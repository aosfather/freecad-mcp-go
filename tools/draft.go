package tools

import (
	"strconv"
	"strings"
)

func init() {
	register("draft",
		ToolDef{Name: "freecad_draft_wire", Handler: hDraftWire},
		ToolDef{Name: "freecad_draft_bspline", Handler: hDraftBspline},
		ToolDef{Name: "freecad_draft_polygon", Handler: hDraftPolygon},
		ToolDef{Name: "freecad_draft_ellipse", Handler: hDraftEllipse},
		ToolDef{Name: "freecad_draft_rectangle", Handler: hDraftRectangle},
		ToolDef{Name: "freecad_draft_facebinder", Handler: hDraftFacebinder},
		ToolDef{Name: "freecad_draft_clone", Handler: hDraftClone},
		ToolDef{Name: "freecad_draft_shapestring", Handler: hDraftShapestring},
		ToolDef{Name: "freecad_draft_move", Handler: hDraftMove},
		ToolDef{Name: "freecad_draft_rotate", Handler: hDraftRotate},
		ToolDef{Name: "freecad_draft_scale", Handler: hDraftScale},
		ToolDef{Name: "freecad_draft_offset", Handler: hDraftOffset},
		ToolDef{Name: "freecad_draft_upgrade", Handler: hDraftUpgrade},
		ToolDef{Name: "freecad_draft_downgrade", Handler: hDraftDowngrade},
		ToolDef{Name: "freecad_draft_path_array", Handler: hDraftPathArray},
		ToolDef{Name: "freecad_draft_dimension", Handler: hDraftDimension},
		ToolDef{Name: "freecad_draft_shape2dview", Handler: hDraftShape2dview},
	)
}

func hDraftWire(args map[string]any, b *Bridge) (Result, error) {
	points, _ := args["points"].([]any)
	closed := ArgBool(args, "closed", false)
	wireName := ArgStr(args, "name", "Wire")

	parts := make([]string, 0, len(points))
	for _, p := range points {
		pm, _ := p.(map[string]any)
		x, _ := pm["x"].(float64)
		y, _ := pm["y"].(float64)
		z, _ := pm["z"].(float64)
		parts = append(parts, "FreeCAD.Vector("+PyNum(x)+", "+PyNum(y)+", "+PyNum(z)+")")
	}
	pointsList := strings.Join(parts, ", ")

	closedPy := "False"
	if closed {
		closedPy = "True"
	}

	return b.Run(`
` + DocPreamble + `
import Draft
pts = [` + pointsList + `]
wire = Draft.make_wire(pts, closed=` + closedPy + `, face=False)
wire.Label = ` + PyQuote(wireName) + `
doc.recompute()
_mcp_result["result"] = {"name": wire.Name, "label": wire.Label, "points": ` + strconv.Itoa(len(points)) + `, "closed": ` + strconv.FormatBool(closed) + `}
`), nil
}

func hDraftBspline(args map[string]any, b *Bridge) (Result, error) {
	points, _ := args["points"].([]any)
	closed := ArgBool(args, "closed", false)
	bsName := ArgStr(args, "name", "BSpline")

	parts := make([]string, 0, len(points))
	for _, p := range points {
		pm, _ := p.(map[string]any)
		x, _ := pm["x"].(float64)
		y, _ := pm["y"].(float64)
		z, _ := pm["z"].(float64)
		parts = append(parts, "FreeCAD.Vector("+PyNum(x)+", "+PyNum(y)+", "+PyNum(z)+")")
	}
	pointsList := strings.Join(parts, ", ")

	closedPy := "False"
	if closed {
		closedPy = "True"
	}

	return b.Run(`
` + DocPreamble + `
import Draft
pts = [` + pointsList + `]
bsp = Draft.make_bspline(pts, closed=` + closedPy + `)
bsp.Label = ` + PyQuote(bsName) + `
doc.recompute()
_mcp_result["result"] = {"name": bsp.Name, "label": bsp.Label, "points": ` + strconv.Itoa(len(points)) + `, "closed": ` + strconv.FormatBool(closed) + `}
`), nil
}

func hDraftPolygon(args map[string]any, b *Bridge) (Result, error) {
	sides, _ := args["sides"].(float64)
	radius, _ := args["radius"].(float64)
	x := ArgNum(args, "x", 0)
	y := ArgNum(args, "y", 0)
	z := ArgNum(args, "z", 0)
	polyName := ArgStr(args, "name", "Polygon")

	return b.Run(`
` + DocPreamble + `
import Draft
poly = Draft.make_polygon(` + PyNum(sides) + `, ` + PyNum(radius) + `)
poly.Label = ` + PyQuote(polyName) + `
poly.Placement.Base = FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)
doc.recompute()
_mcp_result["result"] = {"name": poly.Name, "label": poly.Label, "sides": ` + PyNum(sides) + `, "radius": ` + PyNum(radius) + `}
`), nil
}

func hDraftEllipse(args map[string]any, b *Bridge) (Result, error) {
	majorRadius, _ := args["majorRadius"].(float64)
	minorRadius, _ := args["minorRadius"].(float64)
	x := ArgNum(args, "x", 0)
	y := ArgNum(args, "y", 0)
	z := ArgNum(args, "z", 0)
	ellName := ArgStr(args, "name", "Ellipse")

	return b.Run(`
` + DocPreamble + `
import Draft
ell = Draft.make_ellipse(` + PyNum(majorRadius*2) + `, ` + PyNum(minorRadius*2) + `)
ell.Label = ` + PyQuote(ellName) + `
ell.Placement.Base = FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)
doc.recompute()
_mcp_result["result"] = {"name": ell.Name, "label": ell.Label, "majorRadius": ` + PyNum(majorRadius) + `, "minorRadius": ` + PyNum(minorRadius) + `}
`), nil
}

func hDraftRectangle(args map[string]any, b *Bridge) (Result, error) {
	width, _ := args["width"].(float64)
	height, _ := args["height"].(float64)
	x := ArgNum(args, "x", 0)
	y := ArgNum(args, "y", 0)
	z := ArgNum(args, "z", 0)
	rectName := ArgStr(args, "name", "Rectangle")

	return b.Run(`
` + DocPreamble + `
import Draft
rect = Draft.make_rectangle(` + PyNum(width) + `, ` + PyNum(height) + `)
rect.Label = ` + PyQuote(rectName) + `
rect.Placement.Base = FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)
doc.recompute()
_mcp_result["result"] = {"name": rect.Name, "label": rect.Label, "width": ` + PyNum(width) + `, "height": ` + PyNum(height) + `}
`), nil
}

func hDraftFacebinder(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	faceNames, _ := args["faceNames"].([]any)
	fbName := ArgStr(args, "name", "FaceBinder")

	faceParts := make([]string, 0, len(faceNames))
	for _, f := range faceNames {
		s, _ := f.(string)
		faceParts = append(faceParts, PyQuote(s))
	}

	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
fb = doc.addObject("Part::FeaturePython", ` + PyQuote(fbName) + `)
import Draft
fb_obj = Draft.make_facebinder([(obj, (` + strings.Join(faceParts, ", ") + `,))])
fb_obj.Label = ` + PyQuote(fbName) + `
doc.recompute()
_mcp_result["result"] = {"name": fb_obj.Name, "label": fb_obj.Label, "faces": ` + PyQuoteAny(faceNames) + `}
`), nil
}

func hDraftClone(args map[string]any, b *Bridge) (Result, error) {
	objectNames, _ := args["objectNames"].([]any)
	cloneName := ArgStr(args, "name", "Clone")

	objsParts := make([]string, 0, len(objectNames))
	for _, n := range objectNames {
		s, _ := n.(string)
		objsParts = append(objsParts, "doc.getObject("+PyQuote(s)+")")
	}
	objsCode := strings.Join(objsParts, ", ")

	return b.Run(`
` + DocPreamble + `
import Draft
objs = [` + objsCode + `]
missing = [` + PyQuoteAny(objectNames) + `[i] for i, o in enumerate(objs) if o is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
clone = Draft.make_clone(objs)
clone.Label = ` + PyQuote(cloneName) + `
doc.recompute()
_mcp_result["result"] = {"name": clone.Name, "label": clone.Label, "sources": ` + PyQuoteAny(objectNames) + `}
`), nil
}

func hDraftShapestring(args map[string]any, b *Bridge) (Result, error) {
	text, _ := args["text"].(string)
	size := ArgNum(args, "size", 10)
	fontFile, _ := args["fontFile"].(string)
	x := ArgNum(args, "x", 0)
	y := ArgNum(args, "y", 0)
	z := ArgNum(args, "z", 0)
	ssName := ArgStr(args, "name", "ShapeString")

	fontCode := `"/System/Library/Fonts/Helvetica.ttc"`
	if fontFile != "" {
		fontCode = PyQuote(fontFile)
	}

	return b.Run(`
` + DocPreamble + `
import Draft
import os
font = ` + fontCode + `
if not os.path.exists(font):
    font = "/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf"
    if not os.path.exists(font):
        font = ""
ss = Draft.make_shapestring(` + PyQuote(text) + `, font, ` + PyNum(size) + `)
ss.Label = ` + PyQuote(ssName) + `
ss.Placement.Base = FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)
doc.recompute()
_mcp_result["result"] = {"name": ss.Name, "label": ss.Label, "text": ` + PyQuote(text) + `, "size": ` + PyNum(size) + `}
`), nil
}

func hDraftMove(args map[string]any, b *Bridge) (Result, error) {
	objectNames, _ := args["objectNames"].([]any)
	x, _ := args["x"].(float64)
	y, _ := args["y"].(float64)
	z, _ := args["z"].(float64)
	copyFlag := ArgBool(args, "copy", false)

	objsParts := make([]string, 0, len(objectNames))
	for _, n := range objectNames {
		s, _ := n.(string)
		objsParts = append(objsParts, "doc.getObject("+PyQuote(s)+")")
	}
	objsCode := strings.Join(objsParts, ", ")

	copyPy := "False"
	if copyFlag {
		copyPy = "True"
	}

	return b.Run(`
` + DocPreamble + `
import Draft
objs = [` + objsCode + `]
result = Draft.move(objs, FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `), copy=` + copyPy + `)
doc.recompute()
moved = [r.Name for r in (result if isinstance(result, list) else [result])] if result else ` + PyQuoteAny(objectNames) + `
_mcp_result["result"] = {"moved": moved, "vector": {"x": ` + PyNum(x) + `, "y": ` + PyNum(y) + `, "z": ` + PyNum(z) + `}, "copy": ` + strconv.FormatBool(copyFlag) + `}
`), nil
}

func hDraftRotate(args map[string]any, b *Bridge) (Result, error) {
	objectNames, _ := args["objectNames"].([]any)
	angle, _ := args["angle"].(float64)
	cx := ArgNum(args, "centerX", 0)
	cy := ArgNum(args, "centerY", 0)
	cz := ArgNum(args, "centerZ", 0)
	ax := ArgNum(args, "axisX", 0)
	ay := ArgNum(args, "axisY", 0)
	az := ArgNum(args, "axisZ", 1)
	copyFlag := ArgBool(args, "copy", false)

	objsParts := make([]string, 0, len(objectNames))
	for _, n := range objectNames {
		s, _ := n.(string)
		objsParts = append(objsParts, "doc.getObject("+PyQuote(s)+")")
	}
	objsCode := strings.Join(objsParts, ", ")

	copyPy := "False"
	if copyFlag {
		copyPy = "True"
	}

	return b.Run(`
` + DocPreamble + `
import Draft
objs = [` + objsCode + `]
result = Draft.rotate(objs, ` + PyNum(angle) + `, FreeCAD.Vector(` + PyNum(cx) + `, ` + PyNum(cy) + `, ` + PyNum(cz) + `), FreeCAD.Vector(` + PyNum(ax) + `, ` + PyNum(ay) + `, ` + PyNum(az) + `), copy=` + copyPy + `)
doc.recompute()
rotated = [r.Name for r in (result if isinstance(result, list) else [result])] if result else ` + PyQuoteAny(objectNames) + `
_mcp_result["result"] = {"rotated": rotated, "angle": ` + PyNum(angle) + `, "copy": ` + strconv.FormatBool(copyFlag) + `}
`), nil
}

func hDraftScale(args map[string]any, b *Bridge) (Result, error) {
	objectNames, _ := args["objectNames"].([]any)
	scaleX, _ := args["scaleX"].(float64)
	scaleY := ArgNum(args, "scaleY", scaleX)
	scaleZ := ArgNum(args, "scaleZ", scaleX)
	cx := ArgNum(args, "centerX", 0)
	cy := ArgNum(args, "centerY", 0)
	cz := ArgNum(args, "centerZ", 0)
	copyFlag := ArgBool(args, "copy", true)

	objsParts := make([]string, 0, len(objectNames))
	for _, n := range objectNames {
		s, _ := n.(string)
		objsParts = append(objsParts, "doc.getObject("+PyQuote(s)+")")
	}
	objsCode := strings.Join(objsParts, ", ")

	copyPy := "False"
	if copyFlag {
		copyPy = "True"
	}

	return b.Run(`
` + DocPreamble + `
import Draft
objs = [` + objsCode + `]
result = Draft.scale(objs, FreeCAD.Vector(` + PyNum(scaleX) + `, ` + PyNum(scaleY) + `, ` + PyNum(scaleZ) + `), FreeCAD.Vector(` + PyNum(cx) + `, ` + PyNum(cy) + `, ` + PyNum(cz) + `), copy=` + copyPy + `)
doc.recompute()
scaled = [r.Name for r in (result if isinstance(result, list) else [result])] if result else ` + PyQuoteAny(objectNames) + `
_mcp_result["result"] = {"scaled": scaled, "scale": {"x": ` + PyNum(scaleX) + `, "y": ` + PyNum(scaleY) + `, "z": ` + PyNum(scaleZ) + `}, "copy": ` + strconv.FormatBool(copyFlag) + `}
`), nil
}

func hDraftOffset(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	distance, _ := args["distance"].(float64)
	copyFlag := ArgBool(args, "copy", true)

	copyPy := "False"
	if copyFlag {
		copyPy = "True"
	}

	return b.Run(`
` + DocPreamble + `
import Draft
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
result = Draft.offset(obj, FreeCAD.Vector(` + PyNum(distance) + `, 0, 0), copy=` + copyPy + `)
doc.recompute()
_mcp_result["result"] = {"name": result.Name if result else ` + PyQuote(objectName) + `, "distance": ` + PyNum(distance) + `, "copy": ` + strconv.FormatBool(copyFlag) + `}
`), nil
}

func hDraftUpgrade(args map[string]any, b *Bridge) (Result, error) {
	objectNames, _ := args["objectNames"].([]any)

	objsParts := make([]string, 0, len(objectNames))
	for _, n := range objectNames {
		s, _ := n.(string)
		objsParts = append(objsParts, "doc.getObject("+PyQuote(s)+")")
	}
	objsCode := strings.Join(objsParts, ", ")

	return b.Run(`
` + DocPreamble + `
import Draft
objs = [` + objsCode + `]
result = Draft.upgrade(objs)
doc.recompute()
added = [o.Name for o in result[0]] if result and result[0] else []
deleted = [o.Name for o in result[1]] if result and len(result) > 1 and result[1] else []
_mcp_result["result"] = {"added": added, "deleted": deleted}
`), nil
}

func hDraftDowngrade(args map[string]any, b *Bridge) (Result, error) {
	objectNames, _ := args["objectNames"].([]any)

	objsParts := make([]string, 0, len(objectNames))
	for _, n := range objectNames {
		s, _ := n.(string)
		objsParts = append(objsParts, "doc.getObject("+PyQuote(s)+")")
	}
	objsCode := strings.Join(objsParts, ", ")

	return b.Run(`
` + DocPreamble + `
import Draft
objs = [` + objsCode + `]
result = Draft.downgrade(objs)
doc.recompute()
added = [o.Name for o in result[0]] if result and result[0] else []
deleted = [o.Name for o in result[1]] if result and len(result) > 1 and result[1] else []
_mcp_result["result"] = {"added": added, "deleted": deleted}
`), nil
}

func hDraftPathArray(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	pathName, _ := args["pathName"].(string)
	count, _ := args["count"].(float64)
	arrName := ArgStr(args, "name", "PathArray")

	return b.Run(`
` + DocPreamble + `
import Draft
obj = doc.getObject(` + PyQuote(objectName) + `)
path = doc.getObject(` + PyQuote(pathName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
if path is None:
    raise ValueError("Path not found: ` + pathName + `")
arr = Draft.make_path_array(obj, path, ` + PyNum(count) + `)
arr.Label = ` + PyQuote(arrName) + `
doc.recompute()
_mcp_result["result"] = {"name": arr.Name, "label": arr.Label, "count": ` + PyNum(count) + `, "path": ` + PyQuote(pathName) + `}
`), nil
}

func hDraftDimension(args map[string]any, b *Bridge) (Result, error) {
	p1x, _ := args["point1X"].(float64)
	p1y, _ := args["point1Y"].(float64)
	p1z := ArgNum(args, "point1Z", 0)
	p2x, _ := args["point2X"].(float64)
	p2y, _ := args["point2Y"].(float64)
	p2z := ArgNum(args, "point2Z", 0)
	dimName := ArgStr(args, "name", "Dimension")

	return b.Run(`
` + DocPreamble + `
import Draft
p1 = FreeCAD.Vector(` + PyNum(p1x) + `, ` + PyNum(p1y) + `, ` + PyNum(p1z) + `)
p2 = FreeCAD.Vector(` + PyNum(p2x) + `, ` + PyNum(p2y) + `, ` + PyNum(p2z) + `)
mid = (p1 + p2) * 0.5
mid.z += 5
dim = Draft.make_dimension(p1, p2, mid)
dim.Label = ` + PyQuote(dimName) + `
doc.recompute()
dist = p1.distanceToPoint(p2)
_mcp_result["result"] = {"name": dim.Name, "label": dim.Label, "distance": dist}
`), nil
}

func hDraftShape2dview(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	dx := ArgNum(args, "directionX", 0)
	dy := ArgNum(args, "directionY", 0)
	dz := ArgNum(args, "directionZ", -1)
	viewName := ArgStr(args, "name", "Shape2DView")

	return b.Run(`
` + DocPreamble + `
import Draft
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
view = Draft.make_shape2dview(obj, FreeCAD.Vector(` + PyNum(dx) + `, ` + PyNum(dy) + `, ` + PyNum(dz) + `))
view.Label = ` + PyQuote(viewName) + `
doc.recompute()
_mcp_result["result"] = {"name": view.Name, "label": view.Label, "source": ` + PyQuote(objectName) + `}
`), nil
}
