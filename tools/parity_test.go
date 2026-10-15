package tools

import (
	"encoding/json"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestToolParity ensures every schema in schemas.json has a registered Go
// handler and every registered handler has a schema, with the exact same
// tool count as the TS source (165).
func TestToolParity(t *testing.T) {
	defs := Defs()
	if len(defs) != len(schemaIndex) {
		t.Errorf("tool count mismatch: %d handlers vs %d schemas", len(defs), len(schemaIndex))
	}
	if len(defs) != 165 {
		t.Errorf("expected 165 tools, got %d", len(defs))
	}

	seen := map[string]bool{}
	for _, d := range defs {
		if seen[d.Name] {
			t.Errorf("duplicate tool registration: %s", d.Name)
		}
		seen[d.Name] = true
		if _, _, ok := SchemaOf(d.Name); !ok {
			t.Errorf("handler registered without schema: %s", d.Name)
		}
		if d.Handler == nil {
			t.Errorf("tool %s has nil handler", d.Name)
		}
	}
	for name := range schemaIndex {
		if !seen[name] {
			t.Errorf("schema without registered handler: %s", name)
		}
	}
}

// TestSchemasAreValidJSON validates every embedded schema parses and looks
// like a JSON Schema object.
func TestSchemasAreValidJSON(t *testing.T) {
	for name, meta := range schemaIndex {
		var schema map[string]any
		if err := json.Unmarshal(meta.Schema, &schema); err != nil {
			t.Errorf("schema for %s is not valid JSON: %v", name, err)
			continue
		}
		if meta.Description == "" {
			t.Errorf("tool %s has empty description", name)
		}
		if _, ok := schema["type"]; !ok {
			t.Errorf("schema for %s has no type", name)
		}
	}
}

// fakeFreeCAD serves one request: reads code until the end sentinel, replies
// with a fixed JSON result, closes.
func fakeFreeCAD(t *testing.T, response string) (done chan struct{}) {
	ln, err := net.Listen("tcp", "127.0.0.1:12345")
	if err != nil {
		t.Skipf("port 12345 not free (real FreeCAD running?): %v", err)
	}
	done = make(chan struct{})
	var once sync.Once
	go func() {
		// The bridge probes with a check connection, then opens a second
		// connection to execute — serve both.
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			buf := make([]byte, 1<<20)
			var got strings.Builder
			for {
				n, err := conn.Read(buf)
				if n > 0 {
					got.Write(buf[:n])
				}
				if err != nil || strings.Contains(got.String(), endSentinel) {
					break
				}
			}
			if strings.Contains(got.String(), endSentinel) {
				once.Do(func() { close(done) })
			}
			_, _ = conn.Write([]byte(response))
			_ = conn.Close()
		}
	}()
	t.Cleanup(func() { _ = ln.Close() })
	return done
}

// TestBridgeSocketSmoke exercises the socket path end-to-end with a fake
// FreeCAD GUI macro, through a real registered handler.
func TestBridgeSocketSmoke(t *testing.T) {
	done := fakeFreeCAD(t, `{"success": true, "result": {"name": "Box", "volume": 1000.0}}`)

	b := NewBridge("/Applications/FreeCAD.app/Contents/Resources/bin/freecadcmd", 5*time.Second)
	defer b.Destroy()

	var def ToolDef
	for _, d := range Defs() {
		if d.Name == "freecad_create_box" {
			def = d
			break
		}
	}
	res, err := def.Handler(map[string]any{
		"length": float64(10), "width": float64(20), "height": float64(30),
		"name": "MyBox",
	}, b)
	if err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %s", res.Text)
	}
	if !strings.Contains(res.Text, `"volume": 1000`) {
		t.Fatalf("unexpected result text: %s", res.Text)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("fake server never got a request")
	}
}

// TestBridgeSocketError verifies FreeCAD-side failures become error Results.
func TestBridgeSocketError(t *testing.T) {
	done := fakeFreeCAD(t, `{"success": false, "error": "Object not found: Foo"}`)

	b := NewBridge("", 5*time.Second)
	defer b.Destroy()
	res := b.Run(`pass`)
	if !res.IsError {
		t.Fatalf("expected error result, got: %+v", res)
	}
	if !strings.Contains(res.Text, "Object not found: Foo") {
		t.Fatalf("error text mismatch: %s", res.Text)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("fake server never got a request")
	}
}

// TestValidationRules spot-checks the ported validation behavior.
func TestValidationRules(t *testing.T) {
	if _, err := ValidatePositiveNumber(float64(0), "length"); err == nil {
		t.Error("zero should be rejected by ValidatePositiveNumber")
	}
	if _, err := ValidatePositiveNumber(float64(-1), "length"); err == nil {
		t.Error("negative should be rejected")
	}
	if _, err := ValidateNumber(float64(90), "angle", -89, 89); err == nil {
		t.Error("90 should exceed max 89")
	}
	if _, err := ValidateObjectName("My Box_1", "name"); err != nil {
		t.Errorf("valid object name rejected: %v", err)
	}
	if _, err := ValidateObjectName("Bad;Name", "name"); err == nil {
		t.Error("invalid characters should be rejected")
	}
	if _, err := ValidateFilePath("/tmp/ok.step", "path"); err != nil {
		t.Errorf("valid path rejected: %v", err)
	}
	for _, bad := range []string{"relative.step", "/etc/passwd", "/tmp/../etc/passwd"} {
		if _, err := ValidateFilePath(bad, "path"); err == nil {
			t.Errorf("path %q should be rejected", bad)
		}
	}
	if _, err := ValidateArray([]any{}, "items"); err == nil {
		t.Error("empty array should be rejected")
	}
}
