package tools

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func init() {
	register("techdraw",
		ToolDef{Name: "freecad_techdraw_create_page", Handler: hTdCreatePage},
		ToolDef{Name: "freecad_techdraw_add_view", Handler: hTdAddView},
		ToolDef{Name: "freecad_techdraw_add_projection_group", Handler: hTdAddProjectionGroup},
		ToolDef{Name: "freecad_techdraw_add_dimension", Handler: hTdAddDimension},
		ToolDef{Name: "freecad_techdraw_export_svg", Handler: hTdExportSvg},
		ToolDef{Name: "freecad_techdraw_export_dxf", Handler: hTdExportDxf},
	)
}

func hTdCreatePage(args map[string]any, b *Bridge) (Result, error) {
	template := ArgStr(args, "template", "A4_Landscape")
	pageName := ArgStr(args, "name", "Page")
	code := `
` + DocPreamble + `
page = doc.addObject("TechDraw::DrawPage", ` + PyQuote(pageName) + `)
# Find the template file
import os
template_name = ` + PyQuote(template) + ` + ".svg"
# Try standard FreeCAD template locations
td_paths = [
    os.path.join(FreeCAD.getResourceDir(), "Mod", "TechDraw", "Templates"),
    os.path.join(FreeCAD.getResourceDir(), "share", "Mod", "TechDraw", "Templates"),
]
template_path = None
for tp in td_paths:
    candidate = os.path.join(tp, template_name)
    if os.path.exists(candidate):
        template_path = candidate
        break
if template_path:
    tmpl = doc.addObject("TechDraw::DrawSVGTemplate", "Template")
    tmpl.Template = template_path
    page.Template = tmpl
doc.recompute()
_mcp_result["result"] = {"name": page.Name, "template": ` + PyQuote(template) + `, "hasTemplate": template_path is not None}
`
	return b.Run(code), nil
}

func hTdAddView(args map[string]any, b *Bridge) (Result, error) {
	pageName, _ := args["pageName"].(string)
	objectName, _ := args["objectName"].(string)
	direction := ArgStr(args, "direction", "Front")
	scale := ArgNum(args, "scale", 1.0)
	x, hasX := args["x"].(float64)
	y, hasY := args["y"].(float64)
	viewName := ArgStr(args, "name", "View")
	dirMap := map[string]string{
		"Front":     "FreeCAD.Vector(0, 0, 1)",
		"Back":      "FreeCAD.Vector(0, 0, -1)",
		"Top":       "FreeCAD.Vector(0, 1, 0)",
		"Bottom":    "FreeCAD.Vector(0, -1, 0)",
		"Left":      "FreeCAD.Vector(-1, 0, 0)",
		"Right":     "FreeCAD.Vector(1, 0, 0)",
		"Isometric": "FreeCAD.Vector(1, 1, 1)",
	}
	dirVec := dirMap[direction]
	if dirVec == "" {
		dirVec = dirMap["Front"]
	}
	xyLines := ""
	if hasX {
		xyLines += `view.X = ` + PyNum(x) + "\n"
	}
	if hasY {
		xyLines += `view.Y = ` + PyNum(y) + "\n"
	}
	code := `
` + DocPreamble + `
page = doc.getObject(` + PyQuote(pageName) + `)
if page is None:
    raise ValueError("Page not found: ` + pageName + `")
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
view = doc.addObject("TechDraw::DrawViewPart", ` + PyQuote(viewName) + `)
view.Source = [obj]
view.Direction = ` + dirVec + `
view.Scale = ` + PyNum(scale) + `
` + xyLines + `page.addView(view)
doc.recompute()
_mcp_result["result"] = {"name": view.Name, "page": page.Name, "direction": ` + PyQuote(direction) + `, "scale": ` + PyNum(scale) + `}
`
	return b.Run(code), nil
}

func hTdAddProjectionGroup(args map[string]any, b *Bridge) (Result, error) {
	pageName, _ := args["pageName"].(string)
	objectName, _ := args["objectName"].(string)
	views, ok := args["views"].([]any)
	if !ok {
		views = []any{"Front", "Top", "Right"}
	}
	scale := ArgNum(args, "scale", 1.0)
	pgName := ArgStr(args, "name", "ProjGroup")
	parts := make([]string, len(views))
	for i, v := range views {
		parts[i] = PyQuoteAny(v)
	}
	viewsList := strings.Join(parts, ", ")
	code := `
` + DocPreamble + `
page = doc.getObject(` + PyQuote(pageName) + `)
if page is None:
    raise ValueError("Page not found: ` + pageName + `")
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
pg = doc.addObject("TechDraw::DrawProjGroup", ` + PyQuote(pgName) + `)
pg.Source = [obj]
pg.Scale = ` + PyNum(scale) + `
page.addView(pg)
# Add the anchor (front) view first
pg.addProjection("Front")
# Add additional views
for v in [` + viewsList + `]:
    if v != "Front":
        pg.addProjection(v)
doc.recompute()
_mcp_result["result"] = {"name": pg.Name, "page": page.Name, "views": [` + viewsList + `], "scale": ` + PyNum(scale) + `}
`
	return b.Run(code), nil
}

func hTdAddDimension(args map[string]any, b *Bridge) (Result, error) {
	pageName, _ := args["pageName"].(string)
	viewName, _ := args["viewName"].(string)
	dimensionType, _ := args["dimensionType"].(string)
	edge, _ := args["edge"].(string)
	vertex1, _ := args["vertex1"].(string)
	vertex2, _ := args["vertex2"].(string)
	dimName := ArgStr(args, "name", "Dimension")

	typeMap := map[string]string{
		"distance":   "TechDraw::DrawViewDimension",
		"horizontal": "TechDraw::DrawViewDimension",
		"vertical":   "TechDraw::DrawViewDimension",
		"radius":     "TechDraw::DrawViewDimension",
		"diameter":   "TechDraw::DrawViewDimension",
		"angle":      "TechDraw::DrawViewDimension",
	}
	dimType := typeMap[dimensionType]
	if dimType == "" {
		dimType = "TechDraw::DrawViewDimension"
	}

	var refsCode string
	if edge != "" {
		refsCode = `dim.References2D = [(view, ` + PyQuote(edge) + `)]`
	} else if vertex1 != "" && vertex2 != "" {
		refsCode = `dim.References2D = [(view, ` + PyQuote(vertex1) + `), (view, ` + PyQuote(vertex2) + `)]`
	} else {
		return Result{Text: "Must provide either edge or both vertex1 and vertex2", IsError: true}, nil
	}

	dimTypeCap := dimensionType
	if dimensionType != "" {
		r, size := utf8.DecodeRuneInString(dimensionType)
		dimTypeCap = string(unicode.ToUpper(r)) + dimensionType[size:]
	}

	code := `
` + DocPreamble + `
page = doc.getObject(` + PyQuote(pageName) + `)
if page is None:
    raise ValueError("Page not found: ` + pageName + `")
view = doc.getObject(` + PyQuote(viewName) + `)
if view is None:
    raise ValueError("View not found: ` + viewName + `")
dim = doc.addObject(` + PyQuote(dimType) + `, ` + PyQuote(dimName) + `)
dim.Type = ` + PyQuote(dimTypeCap) + `
` + refsCode + `
page.addView(dim)
doc.recompute()
_mcp_result["result"] = {"name": dim.Name, "type": ` + PyQuote(dimensionType) + `, "page": page.Name}
`
	return b.Run(code), nil
}

func hTdExportSvg(args map[string]any, b *Bridge) (Result, error) {
	pageName, _ := args["pageName"].(string)
	filePath, _ := args["filePath"].(string)
	code := `
` + DocPreamble + `
import TechDraw
page = doc.getObject(` + PyQuote(pageName) + `)
if page is None:
    raise ValueError("Page not found: ` + pageName + `")
doc.recompute()
TechDraw.writeSVGPage(page, ` + PyQuote(filePath) + `)
import os
_mcp_result["result"] = {"page": page.Name, "filePath": ` + PyQuote(filePath) + `, "size_bytes": os.path.getsize(` + PyQuote(filePath) + `)}
`
	return b.Run(code), nil
}

func hTdExportDxf(args map[string]any, b *Bridge) (Result, error) {
	pageName, _ := args["pageName"].(string)
	filePath, _ := args["filePath"].(string)
	code := `
` + DocPreamble + `
import TechDraw
page = doc.getObject(` + PyQuote(pageName) + `)
if page is None:
    raise ValueError("Page not found: ` + pageName + `")
doc.recompute()
TechDraw.writeDXFPage(page, ` + PyQuote(filePath) + `)
import os
_mcp_result["result"] = {"page": page.Name, "filePath": ` + PyQuote(filePath) + `, "size_bytes": os.path.getsize(` + PyQuote(filePath) + `)}
`
	return b.Run(code), nil
}
