package tools

import "strings"

func init() {
	register("operations",
		ToolDef{Name: "freecad_boolean_fuse", Handler: hOpBooleanFuse},
		ToolDef{Name: "freecad_boolean_cut", Handler: hOpBooleanCut},
		ToolDef{Name: "freecad_boolean_intersect", Handler: hOpBooleanIntersect},
		ToolDef{Name: "freecad_fillet", Handler: hOpFillet},
		ToolDef{Name: "freecad_chamfer", Handler: hOpChamfer},
		ToolDef{Name: "freecad_move_object", Handler: hOpMoveObject},
		ToolDef{Name: "freecad_rotate_object", Handler: hOpRotateObject},
		ToolDef{Name: "freecad_copy_object", Handler: hOpCopyObject},
		ToolDef{Name: "freecad_delete_object", Handler: hOpDeleteObject},
		ToolDef{Name: "freecad_mirror_object", Handler: hOpMirrorObject},
		ToolDef{Name: "freecad_check_geometry", Handler: hOpCheckGeometry},
		ToolDef{Name: "freecad_refine_shape", Handler: hOpRefineShape},
		ToolDef{Name: "freecad_boolean_fragments", Handler: hOpBooleanFragments},
		ToolDef{Name: "freecad_slice", Handler: hOpSlice},
		ToolDef{Name: "freecad_boolean_xor", Handler: hOpBooleanXor},
		ToolDef{Name: "freecad_join_connect", Handler: hOpJoinConnect},
		ToolDef{Name: "freecad_join_cutout", Handler: hOpJoinCutout},
	)
}

func hOpBooleanFuse(args map[string]any, b *Bridge) (Result, error) {
	obj1, _ := args["object1"].(string)
	obj2, _ := args["object2"].(string)
	fuseName := ArgStr(args, "name", "Fuse")
	return b.Run(`
doc = FreeCAD.ActiveDocument
o1 = doc.getObject(` + PyQuote(obj1) + `)
o2 = doc.getObject(` + PyQuote(obj2) + `)
if o1 is None:
    raise ValueError("Object not found: ` + obj1 + `")
if o2 is None:
    raise ValueError("Object not found: ` + obj2 + `")
fuse = doc.addObject("Part::Fuse", ` + PyQuote(fuseName) + `)
fuse.Shape1 = o1
fuse.Shape2 = o2
doc.recompute()
_mcp_result["result"] = {"name": fuse.Name, "volume": fuse.Shape.Volume}
`), nil
}

func hOpBooleanCut(args map[string]any, b *Bridge) (Result, error) {
	obj1, _ := args["object1"].(string)
	obj2, _ := args["object2"].(string)
	cutName := ArgStr(args, "name", "Cut")
	return b.Run(`
doc = FreeCAD.ActiveDocument
o1 = doc.getObject(` + PyQuote(obj1) + `)
o2 = doc.getObject(` + PyQuote(obj2) + `)
if o1 is None:
    raise ValueError("Object not found: ` + obj1 + `")
if o2 is None:
    raise ValueError("Object not found: ` + obj2 + `")
cut = doc.addObject("Part::Cut", ` + PyQuote(cutName) + `)
cut.Base = o1
cut.Tool = o2
doc.recompute()
_mcp_result["result"] = {"name": cut.Name, "volume": cut.Shape.Volume}
`), nil
}

func hOpBooleanIntersect(args map[string]any, b *Bridge) (Result, error) {
	obj1, _ := args["object1"].(string)
	obj2, _ := args["object2"].(string)
	intName := ArgStr(args, "name", "Common")
	return b.Run(`
doc = FreeCAD.ActiveDocument
o1 = doc.getObject(` + PyQuote(obj1) + `)
o2 = doc.getObject(` + PyQuote(obj2) + `)
if o1 is None:
    raise ValueError("Object not found: ` + obj1 + `")
if o2 is None:
    raise ValueError("Object not found: ` + obj2 + `")
common = doc.addObject("Part::Common", ` + PyQuote(intName) + `)
common.Shape1 = o1
common.Shape2 = o2
doc.recompute()
_mcp_result["result"] = {"name": common.Name, "volume": common.Shape.Volume}
`), nil
}

func hOpFillet(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	radius, err := ValidatePositiveNumber(args["radius"], "radius")
	if err != nil {
		return Result{}, err
	}
	edgeIndices, _ := args["edgeIndices"].([]any)
	var edgesPython string
	if edgeIndices != nil {
		parts := make([]string, 0, len(edgeIndices))
		for _, e := range edgeIndices {
			i, _ := e.(float64)
			parts = append(parts, "("+PyNum(i)+", "+PyNum(radius)+", "+PyNum(radius)+")")
		}
		edgesPython = "[" + strings.Join(parts, ", ") + "]"
	} else {
		edgesPython = "[(i, " + PyNum(radius) + ", " + PyNum(radius) + ") for i in range(1, len(obj.Shape.Edges) + 1)]"
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
edges = ` + edgesPython + `
fillet = doc.addObject("Part::Fillet", "Fillet")
fillet.Base = obj
fillet.Shape = obj.Shape.makeFillet(` + PyNum(radius) + `, [obj.Shape.Edges[e[0]-1] for e in edges])
doc.recompute()
_mcp_result["result"] = {"name": fillet.Name, "edges": len(edges)}
`), nil
}

func hOpChamfer(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	distance, err := ValidatePositiveNumber(args["distance"], "distance")
	if err != nil {
		return Result{}, err
	}
	edgeIndices, _ := args["edgeIndices"].([]any)
	var edgesPython string
	if edgeIndices != nil {
		parts := make([]string, 0, len(edgeIndices))
		for _, e := range edgeIndices {
			i, _ := e.(float64)
			parts = append(parts, "("+PyNum(i)+", "+PyNum(distance)+", "+PyNum(distance)+")")
		}
		edgesPython = "[" + strings.Join(parts, ", ") + "]"
	} else {
		edgesPython = "[(i, " + PyNum(distance) + ", " + PyNum(distance) + ") for i in range(1, len(obj.Shape.Edges) + 1)]"
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
edges = ` + edgesPython + `
chamfer = doc.addObject("Part::Chamfer", "Chamfer")
chamfer.Base = obj
chamfer.Shape = obj.Shape.makeChamfer(` + PyNum(distance) + `, [obj.Shape.Edges[e[0]-1] for e in edges])
doc.recompute()
_mcp_result["result"] = {"name": chamfer.Name, "edges": len(edges)}
`), nil
}

func hOpMoveObject(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	x, _ := args["x"].(float64)
	y, _ := args["y"].(float64)
	z, _ := args["z"].(float64)
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
obj.Placement.Base = FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)
doc.recompute()
_mcp_result["result"] = {"name": obj.Name, "position": {"x": ` + PyNum(x) + `, "y": ` + PyNum(y) + `, "z": ` + PyNum(z) + `}}
`), nil
}

func hOpRotateObject(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	axis, _ := args["axis"].(string)
	angle, _ := args["angle"].(float64)
	axisVec := "FreeCAD.Vector(0, 0, 1)"
	if axis == "x" {
		axisVec = "FreeCAD.Vector(1, 0, 0)"
	} else if axis == "y" {
		axisVec = "FreeCAD.Vector(0, 1, 0)"
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
rot = FreeCAD.Rotation(` + axisVec + `, ` + PyNum(angle) + `)
obj.Placement.Rotation = rot
doc.recompute()
_mcp_result["result"] = {"name": obj.Name, "axis": ` + PyQuote(axis) + `, "angle": ` + PyNum(angle) + `}
`), nil
}

func hOpCopyObject(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	newName, _ := args["newName"].(string)
	labelLine := ""
	if newName != "" {
		labelLine = `copy.Label = ` + PyQuote(newName)
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
copy = doc.copyObject(obj)
` + labelLine + `
doc.recompute()
_mcp_result["result"] = {"name": copy.Name, "label": copy.Label}
`), nil
}

func hOpDeleteObject(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
doc.removeObject(` + PyQuote(objectName) + `)
doc.recompute()
_mcp_result["result"] = {"deleted": ` + PyQuote(objectName) + `}
`), nil
}

func hOpMirrorObject(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	plane, _ := args["plane"].(string)
	mirrorName := ArgStr(args, "name", "Mirror")
	point := "FreeCAD.Vector(0, 0, 0)"
	normal := "FreeCAD.Vector(0, 0, 1)"
	if plane == "XZ" {
		normal = "FreeCAD.Vector(0, 1, 0)"
	} else if plane == "YZ" {
		normal = "FreeCAD.Vector(1, 0, 0)"
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
mirrored_shape = obj.Shape.mirror(` + point + `, ` + normal + `)
mirror_obj = doc.addObject("Part::Feature", ` + PyQuote(mirrorName) + `)
mirror_obj.Shape = mirrored_shape
doc.recompute()
_mcp_result["result"] = {"name": mirror_obj.Name, "plane": ` + PyQuote(plane) + `}
`), nil
}

func hOpCheckGeometry(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
shape = obj.Shape
is_valid = shape.isValid()
is_closed = shape.isClosed() if hasattr(shape, 'isClosed') else None
checks = {
    "isValid": is_valid,
    "isClosed": is_closed,
    "faces": len(shape.Faces),
    "edges": len(shape.Edges),
    "vertices": len(shape.Vertexes),
    "shells": len(shape.Shells),
    "solids": len(shape.Solids),
    "shapeType": shape.ShapeType,
}
errors = []
try:
    shape.check()
except Exception as e:
    errors.append(str(e))
checks["errors"] = errors
_mcp_result["result"] = checks
`), nil
}

func hOpRefineShape(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	refinedName := ArgStr(args, "name", "Refined")
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
refined_shape = obj.Shape.removeSplitter()
result = doc.addObject("Part::Feature", ` + PyQuote(refinedName) + `)
result.Shape = refined_shape
doc.recompute()
_mcp_result["result"] = {"name": result.Name, "faces_before": len(obj.Shape.Faces), "faces_after": len(refined_shape.Faces), "edges_before": len(obj.Shape.Edges), "edges_after": len(refined_shape.Edges)}
`), nil
}

func hOpBooleanFragments(args map[string]any, b *Bridge) (Result, error) {
	objectNames, _ := args["objectNames"].([]any)
	bfName := ArgStr(args, "name", "BooleanFragments")
	return b.Run(`
doc = FreeCAD.ActiveDocument
import BOPTools.SplitFeatures
bf = BOPTools.SplitFeatures.makeBooleanFragments(` + PyQuote(bfName) + `)
bf.Objects = [doc.getObject(n) for n in ` + PyQuoteAny(objectNames) + `]
bf.Mode = "Standard"
doc.recompute()
pieces = len(bf.Shape.Solids) if bf.Shape.Solids else len(bf.Shape.Faces)
_mcp_result["result"] = {"name": bf.Name, "inputObjects": ` + PyQuoteAny(objectNames) + `, "fragments": pieces}
`), nil
}

func hOpSlice(args map[string]any, b *Bridge) (Result, error) {
	baseName, _ := args["baseName"].(string)
	toolName, _ := args["toolName"].(string)
	sliceName := ArgStr(args, "name", "Slice")
	return b.Run(`
doc = FreeCAD.ActiveDocument
import BOPTools.SplitFeatures
sl = BOPTools.SplitFeatures.makeSlice(` + PyQuote(sliceName) + `)
sl.Base = doc.getObject(` + PyQuote(baseName) + `)
sl.Tools = [doc.getObject(` + PyQuote(toolName) + `)]
sl.Mode = "Standard"
doc.recompute()
pieces = len(sl.Shape.Solids) if sl.Shape.Solids else len(sl.Shape.Faces)
_mcp_result["result"] = {"name": sl.Name, "base": ` + PyQuote(baseName) + `, "tool": ` + PyQuote(toolName) + `, "pieces": pieces}
`), nil
}

func hOpBooleanXor(args map[string]any, b *Bridge) (Result, error) {
	obj1, _ := args["object1"].(string)
	obj2, _ := args["object2"].(string)
	xorName := ArgStr(args, "name", "XOR")
	return b.Run(`
doc = FreeCAD.ActiveDocument
import BOPTools.SplitFeatures
xor = BOPTools.SplitFeatures.makeXOR(` + PyQuote(xorName) + `)
xor.Objects = [doc.getObject(` + PyQuote(obj1) + `), doc.getObject(` + PyQuote(obj2) + `)]
doc.recompute()
_mcp_result["result"] = {"name": xor.Name, "volume": xor.Shape.Volume}
`), nil
}

func hOpJoinConnect(args map[string]any, b *Bridge) (Result, error) {
	obj1, _ := args["object1"].(string)
	obj2, _ := args["object2"].(string)
	joinName := ArgStr(args, "name", "JoinConnect")
	return b.Run(`
doc = FreeCAD.ActiveDocument
j = doc.addObject("Part::JoinConnect", ` + PyQuote(joinName) + `)
j.Object1 = doc.getObject(` + PyQuote(obj1) + `)
j.Object2 = doc.getObject(` + PyQuote(obj2) + `)
doc.recompute()
_mcp_result["result"] = {"name": j.Name, "volume": j.Shape.Volume}
`), nil
}

func hOpJoinCutout(args map[string]any, b *Bridge) (Result, error) {
	obj1, _ := args["object1"].(string)
	obj2, _ := args["object2"].(string)
	joinName := ArgStr(args, "name", "JoinCutout")
	return b.Run(`
doc = FreeCAD.ActiveDocument
j = doc.addObject("Part::JoinCutout", ` + PyQuote(joinName) + `)
j.Object1 = doc.getObject(` + PyQuote(obj1) + `)
j.Object2 = doc.getObject(` + PyQuote(obj2) + `)
doc.recompute()
_mcp_result["result"] = {"name": j.Name, "volume": j.Shape.Volume}
`), nil
}
