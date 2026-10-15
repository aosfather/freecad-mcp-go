package tools

import (
	"strconv"
)

func init() {
	register("advanced-operations",
		ToolDef{Name: "freecad_thickness", Handler: hAdvThickness},
		ToolDef{Name: "freecad_offset_3d", Handler: hAdvOffset3D},
		ToolDef{Name: "freecad_section", Handler: hAdvSection},
		ToolDef{Name: "freecad_compound", Handler: hAdvCompound},
		ToolDef{Name: "freecad_linear_array", Handler: hAdvLinearArray},
		ToolDef{Name: "freecad_polar_array", Handler: hAdvPolarArray},
		ToolDef{Name: "freecad_scale_object", Handler: hAdvScaleObject},
		ToolDef{Name: "freecad_extrude", Handler: hAdvExtrude},
		ToolDef{Name: "freecad_get_center_of_mass", Handler: hAdvGetCenterOfMass},
		ToolDef{Name: "freecad_get_face_info", Handler: hAdvGetFaceInfo},
		ToolDef{Name: "freecad_get_edge_info", Handler: hAdvGetEdgeInfo},
	)
}

func hAdvThickness(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	thickness, _ := args["thickness"].(float64)
	faceNames, _ := args["faceNames"].([]any)
	resultName := ArgStr(args, "name", "Thickness")
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
faces = [getattr(obj.Shape, f) for f in ` + PyQuoteAny(faceNames) + `]
shell_shape = obj.Shape.makeThickness(faces, ` + PyNum(thickness) + `, 1e-3)
result = doc.addObject("Part::Feature", ` + PyQuote(resultName) + `)
result.Shape = shell_shape
doc.recompute()
_mcp_result["result"] = {"name": result.Name, "thickness": ` + PyNum(thickness) + `, "removedFaces": ` + PyQuoteAny(faceNames) + `, "volume": result.Shape.Volume}
`), nil
}

func hAdvOffset3D(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	distance, _ := args["distance"].(float64)
	resultName := ArgStr(args, "name", "Offset3D")
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
offset_shape = obj.Shape.makeOffsetShape(` + PyNum(distance) + `, 1e-3, fill=True)
result = doc.addObject("Part::Feature", ` + PyQuote(resultName) + `)
result.Shape = offset_shape
doc.recompute()
_mcp_result["result"] = {"name": result.Name, "offset": ` + PyNum(distance) + `, "volume": result.Shape.Volume}
`), nil
}

func hAdvSection(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	plane := ArgStr(args, "plane", "XY")
	offset := ArgNum(args, "offset", 0)
	resultName := ArgStr(args, "name", "Section")
	base := "FreeCAD.Vector(0, 0, " + PyNum(offset) + ")"
	normal := "FreeCAD.Vector(0, 0, 1)"
	switch plane {
	case "XZ":
		base = "FreeCAD.Vector(0, " + PyNum(offset) + ", 0)"
		normal = "FreeCAD.Vector(0, 1, 0)"
	case "YZ":
		base = "FreeCAD.Vector(" + PyNum(offset) + ", 0, 0)"
		normal = "FreeCAD.Vector(1, 0, 0)"
	}
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
section_shape = obj.Shape.section(Part.makePlane(1000, 1000, ` + base + `, ` + normal + `))
result = doc.addObject("Part::Feature", ` + PyQuote(resultName) + `)
result.Shape = section_shape
doc.recompute()
edges = len(result.Shape.Edges)
_mcp_result["result"] = {"name": result.Name, "plane": ` + PyQuote(plane) + `, "offset": ` + PyNum(offset) + `, "edges": edges}
`), nil
}

func hAdvCompound(args map[string]any, b *Bridge) (Result, error) {
	objectNames, _ := args["objectNames"].([]any)
	compName := ArgStr(args, "name", "Compound")
	return b.Run(`
` + DocPreamble + `
obj_names = ` + PyQuoteAny(objectNames) + `
objs = [doc.getObject(n) for n in obj_names]
missing = [n for n, o in zip(obj_names, objs) if o is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
comp = doc.addObject("Part::Compound", ` + PyQuote(compName) + `)
comp.Links = objs
doc.recompute()
_mcp_result["result"] = {"name": comp.Name, "objects": obj_names, "count": len(obj_names)}
`), nil
}

func hAdvLinearArray(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	dx, _ := args["directionX"].(float64)
	dy, _ := args["directionY"].(float64)
	dz, _ := args["directionZ"].(float64)
	count, _ := args["count"].(float64)
	arrName := ArgStr(args, "name", "LinearArray")
	return b.Run(`
` + DocPreamble + `
import Draft
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
arr = Draft.make_ortho_array(obj, v_x=FreeCAD.Vector(` + PyNum(dx) + `, ` + PyNum(dy) + `, ` + PyNum(dz) + `), n_x=` + PyNum(count) + `, n_y=1, n_z=1)
arr.Label = ` + PyQuote(arrName) + `
doc.recompute()
_mcp_result["result"] = {"name": arr.Name, "label": arr.Label, "count": ` + PyNum(count) + `, "spacing": {"x": ` + PyNum(dx) + `, "y": ` + PyNum(dy) + `, "z": ` + PyNum(dz) + `}}
`), nil
}

func hAdvPolarArray(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	count, _ := args["count"].(float64)
	angle := ArgNum(args, "angle", 360)
	axisX := ArgNum(args, "axisX", 0)
	axisY := ArgNum(args, "axisY", 0)
	axisZ := ArgNum(args, "axisZ", 1)
	centerX := ArgNum(args, "centerX", 0)
	centerY := ArgNum(args, "centerY", 0)
	centerZ := ArgNum(args, "centerZ", 0)
	arrName := ArgStr(args, "name", "PolarArray")
	return b.Run(`
` + DocPreamble + `
import Draft
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
arr = Draft.make_polar_array(obj, ` + PyNum(count) + `, ` + PyNum(angle) + `, FreeCAD.Vector(` + PyNum(centerX) + `, ` + PyNum(centerY) + `, ` + PyNum(centerZ) + `), FreeCAD.Vector(` + PyNum(axisX) + `, ` + PyNum(axisY) + `, ` + PyNum(axisZ) + `))
arr.Label = ` + PyQuote(arrName) + `
doc.recompute()
_mcp_result["result"] = {"name": arr.Name, "label": arr.Label, "count": ` + PyNum(count) + `, "angle": ` + PyNum(angle) + `}
`), nil
}

func hAdvScaleObject(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	scaleX, _ := args["scaleX"].(float64)
	scaleY := ArgNum(args, "scaleY", scaleX)
	scaleZ := ArgNum(args, "scaleZ", scaleX)
	resultName := ArgStr(args, "name", "Scaled")
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
import FreeCAD
mat = FreeCAD.Matrix()
mat.scale(` + PyNum(scaleX) + `, ` + PyNum(scaleY) + `, ` + PyNum(scaleZ) + `)
scaled_shape = obj.Shape.transformGeometry(mat)
result = doc.addObject("Part::Feature", ` + PyQuote(resultName) + `)
result.Shape = scaled_shape
doc.recompute()
_mcp_result["result"] = {"name": result.Name, "scale": {"x": ` + PyNum(scaleX) + `, "y": ` + PyNum(scaleY) + `, "z": ` + PyNum(scaleZ) + `}, "volume": result.Shape.Volume}
`), nil
}

func hAdvExtrude(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	dirX := ArgNum(args, "dirX", 0)
	dirY := ArgNum(args, "dirY", 0)
	dirZ := ArgNum(args, "dirZ", 10)
	solid := ArgBool(args, "solid", true)
	resultName := ArgStr(args, "name", "Extrude")
	solidPy := "False"
	if solid {
		solidPy = "True"
	}
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
ext = doc.addObject("Part::Extrusion", ` + PyQuote(resultName) + `)
ext.Base = obj
ext.Dir = FreeCAD.Vector(` + PyNum(dirX) + `, ` + PyNum(dirY) + `, ` + PyNum(dirZ) + `)
ext.Solid = ` + solidPy + `
doc.recompute()
_mcp_result["result"] = {"name": ext.Name, "direction": {"x": ` + PyNum(dirX) + `, "y": ` + PyNum(dirY) + `, "z": ` + PyNum(dirZ) + `}, "solid": ` + strconv.FormatBool(solid) + `, "volume": ext.Shape.Volume}
`), nil
}

func hAdvGetCenterOfMass(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
com = obj.Shape.CenterOfGravity
sm = obj.Shape.MatrixOfInertia
_mcp_result["result"] = {
    "objectName": ` + PyQuote(objectName) + `,
    "centerOfMass": {"x": com.x, "y": com.y, "z": com.z},
    "volume": obj.Shape.Volume,
    "area": obj.Shape.Area,
    "inertiaMatrix": {
        "Ixx": sm.A11, "Ixy": sm.A12, "Ixz": sm.A13,
        "Iyx": sm.A21, "Iyy": sm.A22, "Iyz": sm.A23,
        "Izx": sm.A31, "Izy": sm.A32, "Izz": sm.A33
    }
}
`), nil
}

func hAdvGetFaceInfo(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	faceIndex, hasFaceIndex := args["faceIndex"].(float64)
	if hasFaceIndex {
		return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
face = obj.Shape.Faces[` + PyNum(faceIndex) + ` - 1]
com = face.CenterOfMass
normal = face.normalAt(0, 0)
_mcp_result["result"] = {
    "objectName": ` + PyQuote(objectName) + `,
    "faceIndex": ` + PyNum(faceIndex) + `,
    "faceName": "Face` + PyNum(faceIndex) + `",
    "type": face.Surface.__class__.__name__,
    "area": face.Area,
    "center": {"x": com.x, "y": com.y, "z": com.z},
    "normal": {"x": normal.x, "y": normal.y, "z": normal.z}
}
`), nil
	}
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
faces = []
for i, face in enumerate(obj.Shape.Faces):
    com = face.CenterOfMass
    normal = face.normalAt(0, 0)
    faces.append({
        "faceIndex": i + 1,
        "faceName": f"Face{i+1}",
        "type": face.Surface.__class__.__name__,
        "area": face.Area,
        "center": {"x": com.x, "y": com.y, "z": com.z},
        "normal": {"x": normal.x, "y": normal.y, "z": normal.z}
    })
_mcp_result["result"] = {"objectName": ` + PyQuote(objectName) + `, "faceCount": len(faces), "faces": faces}
`), nil
}

func hAdvGetEdgeInfo(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	edgeIndex, hasEdgeIndex := args["edgeIndex"].(float64)
	if hasEdgeIndex {
		return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
edge = obj.Shape.Edges[` + PyNum(edgeIndex) + ` - 1]
v1 = edge.Vertexes[0].Point if edge.Vertexes else FreeCAD.Vector(0,0,0)
v2 = edge.Vertexes[-1].Point if len(edge.Vertexes) > 1 else v1
_mcp_result["result"] = {
    "objectName": ` + PyQuote(objectName) + `,
    "edgeIndex": ` + PyNum(edgeIndex) + `,
    "edgeName": "Edge` + PyNum(edgeIndex) + `",
    "type": edge.Curve.__class__.__name__,
    "length": edge.Length,
    "start": {"x": v1.x, "y": v1.y, "z": v1.z},
    "end": {"x": v2.x, "y": v2.y, "z": v2.z}
}
`), nil
	}
	return b.Run(`
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
edges = []
for i, edge in enumerate(obj.Shape.Edges):
    v1 = edge.Vertexes[0].Point if edge.Vertexes else FreeCAD.Vector(0,0,0)
    v2 = edge.Vertexes[-1].Point if len(edge.Vertexes) > 1 else v1
    edges.append({
        "edgeIndex": i + 1,
        "edgeName": f"Edge{i+1}",
        "type": edge.Curve.__class__.__name__,
        "length": edge.Length,
        "start": {"x": v1.x, "y": v1.y, "z": v1.z},
        "end": {"x": v2.x, "y": v2.y, "z": v2.z}
    })
_mcp_result["result"] = {"objectName": ` + PyQuote(objectName) + `, "edgeCount": len(edges), "edges": edges}
`), nil
}
