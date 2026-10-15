package tools

import (
	"encoding/json"
	"strconv"
)

// Result mirrors the TS ToolResult: a single text content item with an
// optional error flag.
type Result struct {
	Text    string
	IsError bool
}

// PyNum formats a float the way JS template literals render numbers.
func PyNum(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}

// PyQuote renders a Go string as a double-quoted literal that is valid inside
// generated Python code (same behavior as JS JSON.stringify on a string).
func PyQuote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return `"` + s + `"`
	}
	return string(b)
}

// PyQuoteAny renders an arbitrary JSON value (array/object) as a literal
// usable inside generated Python code (lists/dicts literals).
func PyQuoteAny(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "null"
	}
	return string(b)
}

// PyStr mirrors the TS pattern of `"<escapePythonString(s)}>"` — manual quotes
// around an escaped string.
func PyStr(s string) string {
	return `"` + EscapePythonString(s) + `"`
}

// ArgStr mirrors `(args.x as string) || 'default'`.
func ArgStr(args map[string]any, key, def string) string {
	if v, ok := args[key].(string); ok && v != "" {
		return v
	}
	return def
}

// ArgNum mirrors `(args.x as number) ?? default`.
func ArgNum(args map[string]any, key string, def float64) float64 {
	if v, ok := args[key].(float64); ok {
		return v
	}
	return def
}

// ArgBool mirrors `(args.x as boolean) ?? default`.
func ArgBool(args map[string]any, key string, def bool) bool {
	if v, ok := args[key].(bool); ok {
		return v
	}
	return def
}

// DocPreamble mirrors the shared DOC_PREAMBLE constant in several TS modules.
const DocPreamble = `doc = FreeCAD.ActiveDocument
if doc is None:
    doc = FreeCAD.newDocument("Unnamed")`
