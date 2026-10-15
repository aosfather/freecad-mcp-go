package tools

import (
	"strconv"
	"strings"
)

func init() {
	register("surface",
		ToolDef{Name: "freecad_surface_filling", Handler: hSurfFilling},
		ToolDef{Name: "freecad_surface_geomfill", Handler: hSurfGeomfill},
		ToolDef{Name: "freecad_surface_sections", Handler: hSurfSections},
		ToolDef{Name: "freecad_surface_extend", Handler: hSurfExtend},
		ToolDef{Name: "freecad_surface_ruled", Handler: hSurfRuled},
	)
}

func hSurfFilling(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	edgeNames, _ := args["edgeNames"].([]any)
	fillName := ArgStr(args, "name", "Filling")
	refs := make([]string, 0, len(edgeNames))
	for _, e := range edgeNames {
		s, _ := e.(string)
		refs = append(refs, "(obj, "+PyQuote(s)+")")
	}
	edgeRefs := strings.Join(refs, ", ")
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
fill = doc.addObject("Surface::Filling", ` + PyQuote(fillName) + `)
fill.BoundEdges = [` + edgeRefs + `]
doc.recompute()
_mcp_result["result"] = {"name": fill.Name, "edges": ` + PyQuoteAny(edgeNames) + `}
`), nil
}

func hSurfGeomfill(args map[string]any, b *Bridge) (Result, error) {
	edges, _ := args["edges"].([]any)
	gfName := ArgStr(args, "name", "GeomFillSurface")
	refs := make([]string, 0, len(edges))
	for _, e := range edges {
		m, _ := e.(map[string]any)
		objectName, _ := m["objectName"].(string)
		edgeName, _ := m["edgeName"].(string)
		refs = append(refs, "(doc.getObject("+PyQuote(objectName)+"), "+PyQuote(edgeName)+")")
	}
	edgeRefs := strings.Join(refs, ", ")
	return b.Run(`
` + DocPreamble + `
gf = doc.addObject("Surface::GeomFillSurface", ` + PyQuote(gfName) + `)
gf.BoundEdges = [` + edgeRefs + `]
doc.recompute()
_mcp_result["result"] = {"name": gf.Name, "edgeCount": ` + strconv.Itoa(len(edges)) + `}
`), nil
}

func hSurfSections(args map[string]any, b *Bridge) (Result, error) {
	sectionNames, _ := args["sectionNames"].([]any)
	secName := ArgStr(args, "name", "Sections")
	return b.Run(`
` + DocPreamble + `
sections = [doc.getObject(n) for n in ` + PyQuoteAny(sectionNames) + `]
missing = [n for n, s in zip(` + PyQuoteAny(sectionNames) + `, sections) if s is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
sec = doc.addObject("Surface::Sections", ` + PyQuote(secName) + `)
sec.NSections = sections
doc.recompute()
_mcp_result["result"] = {"name": sec.Name, "sectionCount": ` + strconv.Itoa(len(sectionNames)) + `}
`), nil
}

func hSurfExtend(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	faceName, _ := args["faceName"].(string)
	extensionLength := ArgNum(args, "extensionLength", 10)
	extName := ArgStr(args, "name", "ExtendSurface")
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
ext = doc.addObject("Surface::Extend", ` + PyQuote(extName) + `)
ext.Face = (obj, [` + PyQuote(faceName) + `])
ext.Tolerance = 0.1
ext.ExtendUNeg = ` + PyNum(extensionLength) + `
ext.ExtendUPos = ` + PyNum(extensionLength) + `
ext.ExtendVNeg = ` + PyNum(extensionLength) + `
ext.ExtendVPos = ` + PyNum(extensionLength) + `
doc.recompute()
_mcp_result["result"] = {"name": ext.Name, "face": ` + PyQuote(faceName) + `, "extension": ` + PyNum(extensionLength) + `}
`), nil
}

func hSurfRuled(args map[string]any, b *Bridge) (Result, error) {
	e1Obj, _ := args["edge1ObjectName"].(string)
	e1Name, _ := args["edge1Name"].(string)
	e2Obj, _ := args["edge2ObjectName"].(string)
	e2Name, _ := args["edge2Name"].(string)
	rsName := ArgStr(args, "name", "RuledSurface")
	return b.Run(`
` + DocPreamble + `
obj1 = doc.getObject(` + PyQuote(e1Obj) + `)
obj2 = doc.getObject(` + PyQuote(e2Obj) + `)
if obj1 is None:
    raise ValueError("Object not found: ` + e1Obj + `")
if obj2 is None:
    raise ValueError("Object not found: ` + e2Obj + `")
edge1 = getattr(obj1.Shape, ` + PyQuote(e1Name) + `)
edge2 = getattr(obj2.Shape, ` + PyQuote(e2Name) + `)
ruled = Part.makeRuledSurface(edge1, edge2)
result = doc.addObject("Part::Feature", ` + PyQuote(rsName) + `)
result.Shape = ruled
doc.recompute()
_mcp_result["result"] = {"name": result.Name, "area": result.Shape.Area}
`), nil
}
