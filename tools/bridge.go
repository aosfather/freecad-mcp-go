package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	defaultFreeCADApp = "/Applications/FreeCAD.app/Contents/Resources"
	resultMarker      = "__MCP_RESULT__"
	socketPort        = 12345
	socketHost        = "127.0.0.1"
	readyMarker       = "__MCP_READY__"
	execPrefix        = "__MCP_EXEC__"
	endSentinel       = "\n__MCP_END__\n"
)

var resourcesRe = regexp.MustCompile(`^(.+/Contents/Resources)`)

// freecadResult mirrors FreeCADResult from types.ts.
type freecadResult struct {
	Success bool `json:"success"`
	Result  any  `json:"result"`
	Error   string
	Output  string `json:"output"`
}

// rawResult tolerates the snake-free JSON keys used by the FreeCAD side.
// The Python side emits {"success": ..., "result": ..., "error": ...}.
// Result may be absent; Error uses the default Go unmarshal via a custom
// struct below.
type rawResult struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
	Error   string          `json:"error"`
	Output  string          `json:"output"`
}

// Bridge is the Go port of the TS FreeCADBridge. It supports two modes:
//   - socket mode: connects to a running FreeCAD GUI with the MCP server macro
//   - headless mode: spawns the FreeCAD python REPL (fallback)
type Bridge struct {
	freecadApp     string
	timeout        time.Duration
	replScriptPath string

	mu        sync.Mutex // serializes execute + process lifecycle
	useSocket tri

	proc  *exec.Cmd
	stdin io.WriteCloser

	bufMu    sync.Mutex
	outbuf   strings.Builder
	waitFor  string
	notify   chan struct{}
	procDead bool
}

type tri int

const (
	modeAuto tri = iota
	modeSocket
	modeHeadless
)

// NewBridge creates a bridge. freecadCmd mirrors the TS FREECAD_CMD handling.
func NewBridge(freecadCmd string, timeout time.Duration) *Bridge {
	if freecadCmd == "" {
		freecadCmd = os.Getenv("FREECAD_CMD")
	}
	if freecadCmd == "" {
		freecadCmd = defaultFreeCADApp + "/bin/freecadcmd"
	}
	app := defaultFreeCADApp
	if m := resourcesRe.FindStringSubmatch(freecadCmd); m != nil {
		app = m[1]
	}
	return &Bridge{
		freecadApp:     app,
		timeout:        timeout,
		replScriptPath: filepath.Join(os.TempDir(), "freecad-mcp-repl.py"),
		notify:         make(chan struct{}, 1),
	}
}

// checkSocket probes the GUI macro socket with a 1s timeout.
func checkSocket() bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(socketHost, fmt.Sprint(socketPort)), time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// Execute mirrors FreeCADBridge.execute: auto-detects mode on first call and
// falls back to headless if the socket dies mid-session.
func (b *Bridge) Execute(code string) freecadResult {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.useSocket == modeAuto {
		if checkSocket() {
			b.useSocket = modeSocket
			fmt.Fprintln(os.Stderr, "[FreeCAD MCP] Connected to FreeCAD GUI via socket")
		} else {
			b.useSocket = modeHeadless
			fmt.Fprintln(os.Stderr, "[FreeCAD MCP] GUI socket not available, using headless mode")
		}
	}

	if b.useSocket == modeSocket {
		res := b.executeSocket(code)
		if !res.Success && strings.Contains(res.Error, "Socket error") {
			fmt.Fprintln(os.Stderr, "[FreeCAD MCP] Socket connection lost, switching to headless mode")
			b.useSocket = modeHeadless
			return b.executeHeadless(code)
		}
		return res
	}
	return b.executeHeadless(code)
}

// Run mirrors FreeCADBridge.run: executes code and formats a tool Result.
func (b *Bridge) Run(code string) Result {
	res := b.Execute(code)
	if !res.Success {
		return Result{Text: "FreeCAD error: " + res.Error, IsError: true}
	}
	var text string
	if raw, ok := res.Result.(json.RawMessage); ok && len(raw) > 0 && string(raw) != "null" {
		var v any
		if err := json.Unmarshal(raw, &v); err == nil {
			switch v.(type) {
			case map[string]any, []any:
				pretty, _ := json.MarshalIndent(v, "", "  ")
				text = string(pretty)
			default:
				text = fmt.Sprint(v)
			}
		} else {
			text = string(raw)
		}
	}
	if text == "" {
		if res.Output != "" {
			text = res.Output
		} else {
			text = "OK"
		}
	}
	return Result{Text: text}
}

// toResult converts a raw JSON response into a freecadResult, normalizing an
// absent/null result to nil.
func toResult(rr rawResult) freecadResult {
	res := freecadResult{Success: rr.Success, Error: rr.Error, Output: rr.Output}
	if len(rr.Result) > 0 && string(rr.Result) != "null" {
		res.Result = json.RawMessage(rr.Result)
	}
	return res
}

// --- socket mode ---

func (b *Bridge) executeSocket(code string) freecadResult {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(socketHost, fmt.Sprint(socketPort)), b.timeout)
	if err != nil {
		return freecadResult{Success: false, Error: fmt.Sprintf("Socket error: %v", err)}
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(b.timeout))

	if _, err := conn.Write([]byte(code + endSentinel)); err != nil {
		return freecadResult{Success: false, Error: fmt.Sprintf("Socket error: %v", err)}
	}

	data, err := io.ReadAll(conn)
	if err != nil {
		return freecadResult{Success: false, Error: fmt.Sprintf("Socket error: %v", err)}
	}

	var rr rawResult
	if err := json.Unmarshal(data, &rr); err != nil {
		return freecadResult{Success: false, Error: "Failed to parse response from FreeCAD GUI"}
	}
	return toResult(rr)
}

// --- headless mode ---

const replScript = `import sys
import json
import FreeCAD
import Part

MARKER = "` + resultMarker + `"

print("__MCP_READY__", flush=True)

while True:
    try:
        line = sys.stdin.readline()
        if not line:
            break
        line = line.strip()
        if not line.startswith("__MCP_EXEC__"):
            continue

        count = int(line.split(":")[1])
        code_lines = []
        for _ in range(count):
            code_lines.append(sys.stdin.readline().rstrip("\n"))
        code = "\n".join(code_lines)

        _mcp_result = {"success": True}
        _mcp_globals = dict(globals())
        _mcp_globals["_mcp_result"] = _mcp_result
        try:
            exec(code, _mcp_globals)
            _mcp_result = _mcp_globals["_mcp_result"]
        except Exception as e:
            _mcp_result = {"success": False, "error": str(e)}

        print(MARKER, flush=True)
        print(json.dumps(_mcp_result), flush=True)
        print(MARKER + "END", flush=True)
    except Exception as e:
        print(MARKER, flush=True)
        print(json.dumps({"success": False, "error": "REPL error: " + str(e)}), flush=True)
        print(MARKER + "END", flush=True)
`

// ensureProcess spawns the headless REPL once and waits for the ready marker.
// Caller must hold b.mu.
func (b *Bridge) ensureProcess() error {
	if b.proc != nil && b.proc.Process != nil && b.procDead == false {
		return nil
	}

	if err := os.WriteFile(b.replScriptPath, []byte(replScript), 0o644); err != nil {
		return fmt.Errorf("write repl script: %w", err)
	}

	pythonBin := b.freecadApp + "/bin/python"
	libPath := b.freecadApp + "/lib"

	cmd := exec.Command(pythonBin, "-u", b.replScriptPath)
	cmd.Env = append(os.Environ(), "PYTHONPATH="+libPath, "PYTHONUNBUFFERED=1")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	b.proc = cmd
	b.stdin = stdin
	b.procDead = false

	// stdout pump: accumulate into outbuf, signal when waitFor appears.
	go func() {
		buf := make([]byte, 64*1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				b.bufMu.Lock()
				b.outbuf.Write(buf[:n])
				found := b.waitFor != "" && strings.Contains(b.outbuf.String(), b.waitFor)
				b.bufMu.Unlock()
				if found {
					select {
					case b.notify <- struct{}{}:
					default:
					}
				}
			}
			if err != nil {
				b.bufMu.Lock()
				b.procDead = true
				b.bufMu.Unlock()
				select {
				case b.notify <- struct{}{}:
				default:
				}
				return
			}
		}
	}()

	// drain stderr
	stderr, err := cmd.StderrPipe()
	if err == nil {
		go func() {
			_, _ = io.Copy(io.Discard, stderr)
		}()
	}

	// wait for __MCP_READY__
	b.bufMu.Lock()
	b.outbuf.Reset()
	b.waitFor = readyMarker
	b.bufMu.Unlock()

	if err := b.waitSignal(); err != nil {
		return fmt.Errorf("FreeCAD initialization timed out")
	}
	b.clearBuffer()
	return nil
}

// waitSignal waits for the current marker (or process death), with timeout.
func (b *Bridge) waitSignal() error {
	timer := time.NewTimer(b.timeout)
	defer timer.Stop()
	select {
	case <-b.notify:
		b.bufMu.Lock()
		dead := b.procDead
		hasMarker := b.waitFor != "" && strings.Contains(b.outbuf.String(), b.waitFor)
		b.bufMu.Unlock()
		if dead && !hasMarker {
			return fmt.Errorf("FreeCAD process exited unexpectedly")
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}
}

func (b *Bridge) clearBuffer() {
	b.bufMu.Lock()
	b.outbuf.Reset()
	b.waitFor = ""
	b.bufMu.Unlock()
}

// executeHeadless mirrors executeHeadless from the TS bridge.
// Caller must hold b.mu.
func (b *Bridge) executeHeadless(code string) freecadResult {
	if err := b.ensureProcess(); err != nil {
		return freecadResult{Success: false, Error: err.Error()}
	}

	lines := strings.Split(code, "\n")
	header := fmt.Sprintf("%s:%d\n", execPrefix, len(lines))
	payload := code + "\n"

	b.bufMu.Lock()
	b.outbuf.Reset()
	b.waitFor = resultMarker + "END"
	b.bufMu.Unlock()

	if _, err := io.WriteString(b.stdin, header+payload); err != nil {
		return freecadResult{Success: false, Error: err.Error()}
	}

	if err := b.waitSignal(); err != nil {
		return freecadResult{Success: false, Error: fmt.Sprintf("FreeCAD execution timed out after %dms", b.timeout.Milliseconds())}
	}

	b.bufMu.Lock()
	output := b.outbuf.String()
	b.bufMu.Unlock()
	b.clearBuffer()

	startIdx := strings.Index(output, resultMarker)
	if startIdx == -1 {
		return freecadResult{Success: false, Error: "No result marker in output", Output: output}
	}
	jsonStart := startIdx + len(resultMarker)
	endIdx := strings.Index(output[jsonStart:], resultMarker+"END")
	if endIdx == -1 {
		return freecadResult{Success: false, Error: "No result marker in output", Output: output}
	}
	jsonStr := strings.TrimSpace(output[jsonStart : jsonStart+endIdx])
	firstLine := strings.TrimSpace(strings.SplitN(jsonStr, "\n", 2)[0])

	var rr rawResult
	if err := json.Unmarshal([]byte(firstLine), &rr); err != nil {
		return freecadResult{Success: false, Error: err.Error(), Output: output}
	}
	res := toResult(rr)
	res.Output = strings.TrimSpace(output[:startIdx])
	return res
}

// ResetMode forces re-detection of socket vs headless mode.
func (b *Bridge) ResetMode() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.useSocket = modeAuto
}

// Destroy terminates the headless process if running.
func (b *Bridge) Destroy() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.proc != nil && b.proc.Process != nil {
		_ = b.stdin.Close()
		_ = b.proc.Process.Kill()
		b.proc = nil
		b.stdin = nil
	}
}
