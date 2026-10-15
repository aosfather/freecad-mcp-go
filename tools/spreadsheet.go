package tools

import (
	"strconv"
	"strings"
)

func init() {
	register("spreadsheet",
		ToolDef{Name: "freecad_spreadsheet_create", Handler: hSheetCreate},
		ToolDef{Name: "freecad_spreadsheet_set", Handler: hSheetSet},
		ToolDef{Name: "freecad_spreadsheet_get", Handler: hSheetGet},
		ToolDef{Name: "freecad_spreadsheet_alias", Handler: hSheetAlias},
		ToolDef{Name: "freecad_set_expression", Handler: hSheetSetExpression},
	)
}

func hSheetCreate(args map[string]any, b *Bridge) (Result, error) {
	sheetName := ArgStr(args, "name", "Spreadsheet")
	return b.Run(`
` + DocPreamble + `
sheet = doc.addObject("Spreadsheet::Sheet", ` + PyQuote(sheetName) + `)
doc.recompute()
_mcp_result["result"] = {"name": sheet.Name, "label": sheet.Label}
`), nil
}

func hSheetSet(args map[string]any, b *Bridge) (Result, error) {
	sheetName, _ := args["sheetName"].(string)
	cells, _ := args["cells"].([]any)
	parts := make([]string, 0, len(cells))
	for _, c := range cells {
		m, _ := c.(map[string]any)
		cell, _ := m["cell"].(string)
		value, _ := m["value"].(string)
		parts = append(parts, "sheet.set("+PyQuote(cell)+", "+PyQuote(value)+")")
	}
	cellsCode := strings.Join(parts, "\n")
	return b.Run(`
` + DocPreamble + `
sheet = doc.getObject(` + PyQuote(sheetName) + `)
if sheet is None:
    raise ValueError("Spreadsheet not found: ` + sheetName + `")
` + cellsCode + `
doc.recompute()
_mcp_result["result"] = {"sheet": sheet.Name, "cellsSet": ` + strconv.Itoa(len(cells)) + `}
`), nil
}

func hSheetGet(args map[string]any, b *Bridge) (Result, error) {
	sheetName, _ := args["sheetName"].(string)
	cells, _ := args["cells"].([]any)
	return b.Run(`
` + DocPreamble + `
sheet = doc.getObject(` + PyQuote(sheetName) + `)
if sheet is None:
    raise ValueError("Spreadsheet not found: ` + sheetName + `")
results = {}
for cell in ` + PyQuoteAny(cells) + `:
    try:
        val = sheet.get(cell)
        results[cell] = str(val) if val is not None else None
    except:
        results[cell] = None
_mcp_result["result"] = {"sheet": sheet.Name, "values": results}
`), nil
}

func hSheetAlias(args map[string]any, b *Bridge) (Result, error) {
	sheetName, _ := args["sheetName"].(string)
	cell, _ := args["cell"].(string)
	alias, _ := args["alias"].(string)
	return b.Run(`
` + DocPreamble + `
sheet = doc.getObject(` + PyQuote(sheetName) + `)
if sheet is None:
    raise ValueError("Spreadsheet not found: ` + sheetName + `")
sheet.setAlias(` + PyQuote(cell) + `, ` + PyQuote(alias) + `)
doc.recompute()
_mcp_result["result"] = {"sheet": sheet.Name, "cell": ` + PyQuote(cell) + `, "alias": ` + PyQuote(alias) + `}
`), nil
}

func hSheetSetExpression(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	property, _ := args["property"].(string)
	expression, _ := args["expression"].(string)
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
obj.setExpression(` + PyQuote(property) + `, ` + PyQuote(expression) + `)
doc.recompute()
_mcp_result["result"] = {"object": obj.Name, "property": ` + PyQuote(property) + `, "expression": ` + PyQuote(expression) + `}
`), nil
}
