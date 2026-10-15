// freecad-mcp-go is a Go port of the freecad-mcp TypeScript MCP server.
// It exposes the same tools (schemas generated from the TS source) and speaks
// the same protocol to FreeCAD: GUI socket macro on 127.0.0.1:12345, with a
// headless freecadcmd REPL fallback.
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"freecad-mcp-go/tools"
)

const defaultFreeCADCmd = "/Applications/FreeCAD.app/Contents/Resources/bin/freecadcmd"

func main() {
	cmd := os.Getenv("FREECAD_CMD")
	if cmd == "" {
		cmd = defaultFreeCADCmd
	}
	bridge := tools.NewBridge(cmd, 30*time.Second)
	defer bridge.Destroy()

	s := server.NewMCPServer("freecad-mcp", "2.0.0")

	count := 0
	for _, def := range tools.Defs() {
		desc, schema, ok := tools.SchemaOf(def.Name)
		if !ok {
			fmt.Fprintf(os.Stderr, "fatal: no schema for tool %q\n", def.Name)
			os.Exit(1)
		}
		s.AddTool(mcp.NewToolWithRawSchema(def.Name, desc, schema), makeHandler(def, bridge))
		count++
	}

	fmt.Fprintf(os.Stderr, "FreeCAD MCP Server (go) running (cmd: %s, tools: %d)\n", cmd, count)

	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func makeHandler(def tools.ToolDef, b *tools.Bridge) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		res, err := def.Handler(req.GetArguments(), b)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Error: %v", err)), nil
		}
		if res.IsError {
			return mcp.NewToolResultError(res.Text), nil
		}
		return mcp.NewToolResultText(res.Text), nil
	}
}
