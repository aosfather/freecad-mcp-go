package tools

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
)

// Handler executes one tool. It mirrors the TS handle*Tool functions:
// validation failures return an error (rendered as "Error: ..."), FreeCAD
// failures are wrapped into an error Result by Bridge.Run.
type Handler func(args map[string]any, b *Bridge) (Result, error)

// ToolDef binds a tool name to its handler. Description and JSON schema come
// from schemas.json (generated verbatim from the TS source dist).
type ToolDef struct {
	Name    string
	Module  string
	Handler Handler
}

var allDefs []ToolDef

// register is called from per-module init() functions.
func register(module string, defs ...ToolDef) {
	for _, d := range defs {
		d.Module = module
		allDefs = append(allDefs, d)
	}
}

// Defs returns every registered tool.
func Defs() []ToolDef {
	return allDefs
}

type toolMeta struct {
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
}

//go:embed schemas.json
var schemasJSON []byte

var schemaIndex = mustParseSchemas()

func mustParseSchemas() map[string]toolMeta {
	idx := map[string]toolMeta{}
	if err := json.Unmarshal(schemasJSON, &idx); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: schemas.json is invalid: %v\n", err)
		os.Exit(1)
	}
	return idx
}

// SchemaOf returns the description and raw input schema for a tool.
func SchemaOf(name string) (string, json.RawMessage, bool) {
	meta, ok := schemaIndex[name]
	return meta.Description, meta.Schema, ok
}
