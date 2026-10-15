package tools

func init() {
	register("import-export",
		ToolDef{Name: "freecad_import_step", Handler: hIeImportStep},
		ToolDef{Name: "freecad_import_stl", Handler: hIeImportStl},
		ToolDef{Name: "freecad_export_step", Handler: hIeExportStep},
		ToolDef{Name: "freecad_export_stl", Handler: hIeExportStl},
		ToolDef{Name: "freecad_export_obj", Handler: hIeExportObj},
		ToolDef{Name: "freecad_measure_distance", Handler: hIeMeasureDistance},
		ToolDef{Name: "freecad_measure_angle", Handler: hIeMeasureAngle},
		ToolDef{Name: "freecad_get_volume", Handler: hIeGetVolume},
		ToolDef{Name: "freecad_get_bounding_box", Handler: hIeGetBoundingBox},
		ToolDef{Name: "freecad_execute_python", Handler: hIeExecutePython},
		ToolDef{Name: "freecad_import_iges", Handler: hIeImportIges},
		ToolDef{Name: "freecad_import_dxf", Handler: hIeImportDxf},
		ToolDef{Name: "freecad_import_svg", Handler: hIeImportSvg},
		ToolDef{Name: "freecad_export_iges", Handler: hIeExportIges},
		ToolDef{Name: "freecad_export_dxf", Handler: hIeExportDxf},
		ToolDef{Name: "freecad_export_svg", Handler: hIeExportSvg},
		ToolDef{Name: "freecad_export_brep", Handler: hIeExportBrep},
	)
}

func hIeImportStep(args map[string]any, b *Bridge) (Result, error) {
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	var docName string
	if raw, ok := args["documentName"].(string); ok && raw != "" {
		dn, err := ValidateString(raw, "documentName")
		if err != nil {
			return Result{}, err
		}
		docName = dn
	}
	docLine := ""
	if docName != "" {
		docLine = `doc = FreeCAD.getDocument(` + PyStr(docName) + `) if ` + PyStr(docName) + ` in [d.Name for d in FreeCAD.listDocuments().values()] else doc` + "\n"
	}
	code := `
` + DocPreamble + `
` + docLine + `Part.insert(` + PyStr(filePath) + `, doc.Name)
doc.recompute()
_mcp_result["result"] = {"document": doc.Name, "objects": [o.Name for o in doc.Objects]}
`
	return b.Run(code), nil
}

func hIeImportStl(args map[string]any, b *Bridge) (Result, error) {
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	var docName string
	if raw, ok := args["documentName"].(string); ok && raw != "" {
		dn, err := ValidateString(raw, "documentName")
		if err != nil {
			return Result{}, err
		}
		docName = dn
	}
	docLine := ""
	if docName != "" {
		docLine = `doc = FreeCAD.getDocument(` + PyStr(docName) + `) if ` + PyStr(docName) + ` in [d.Name for d in FreeCAD.listDocuments().values()] else doc` + "\n"
	}
	code := `
` + DocPreamble + `
` + docLine + `import Mesh
Mesh.insert(` + PyStr(filePath) + `, doc.Name)
doc.recompute()
_mcp_result["result"] = {"document": doc.Name, "objects": [o.Name for o in doc.Objects]}
`
	return b.Run(code), nil
}

func hIeExportStep(args map[string]any, b *Bridge) (Result, error) {
	objectNames, err := ValidateArray(args["objectNames"], "objectNames")
	if err != nil {
		return Result{}, err
	}
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	code := `
` + DocPreamble + `
obj_names = ` + PyQuoteAny(objectNames) + `
objs = [doc.getObject(n) for n in obj_names]
missing = [n for n, o in zip(obj_names, objs) if o is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
Part.export(objs, ` + PyStr(filePath) + `)
import os
_mcp_result["result"] = {"exported": obj_names, "filePath": ` + PyStr(filePath) + `, "size_bytes": os.path.getsize(` + PyStr(filePath) + `)}
`
	return b.Run(code), nil
}

func hIeExportStl(args map[string]any, b *Bridge) (Result, error) {
	objectNames, err := ValidateArray(args["objectNames"], "objectNames")
	if err != nil {
		return Result{}, err
	}
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	code := `
` + DocPreamble + `
import Mesh
obj_names = ` + PyQuoteAny(objectNames) + `
objs = [doc.getObject(n) for n in obj_names]
missing = [n for n, o in zip(obj_names, objs) if o is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
shapes = [o.Shape for o in objs]
mesh = Mesh.Mesh()
for s in shapes:
    mesh.addFacets(s.tessellate(0.1))
mesh.write(` + PyStr(filePath) + `)
_mcp_result["result"] = {"exported": obj_names, "filePath": ` + PyStr(filePath) + `}
`
	return b.Run(code), nil
}

func hIeExportObj(args map[string]any, b *Bridge) (Result, error) {
	objectNames, err := ValidateArray(args["objectNames"], "objectNames")
	if err != nil {
		return Result{}, err
	}
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	code := `
` + DocPreamble + `
import Mesh
obj_names = ` + PyQuoteAny(objectNames) + `
objs = [doc.getObject(n) for n in obj_names]
missing = [n for n, o in zip(obj_names, objs) if o is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
shapes = [o.Shape for o in objs]
mesh = Mesh.Mesh()
for s in shapes:
    mesh.addFacets(s.tessellate(0.1))
mesh.write(` + PyStr(filePath) + `)
_mcp_result["result"] = {"exported": obj_names, "filePath": ` + PyStr(filePath) + `}
`
	return b.Run(code), nil
}

func hIeMeasureDistance(args map[string]any, b *Bridge) (Result, error) {
	point1, hasP1 := args["point1"].(map[string]any)
	point2, hasP2 := args["point2"].(map[string]any)
	obj1, _ := args["object1Name"].(string)
	obj2, _ := args["object2Name"].(string)

	var p1Code string
	if hasP1 {
		x, _ := point1["x"].(float64)
		y, _ := point1["y"].(float64)
		z, _ := point1["z"].(float64)
		p1Code = `FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)`
	} else if obj1 != "" {
		p1Code = `doc.getObject(` + PyStr(obj1) + `).Shape.BoundBox.Center`
	} else {
		return Result{Text: "Must provide either point1 or object1Name", IsError: true}, nil
	}

	var p2Code string
	if hasP2 {
		x, _ := point2["x"].(float64)
		y, _ := point2["y"].(float64)
		z, _ := point2["z"].(float64)
		p2Code = `FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)`
	} else if obj2 != "" {
		p2Code = `doc.getObject(` + PyStr(obj2) + `).Shape.BoundBox.Center`
	} else {
		return Result{Text: "Must provide either point2 or object2Name", IsError: true}, nil
	}

	code := `
` + DocPreamble + `
p1 = ` + p1Code + `
p2 = ` + p2Code + `
diff = p2.sub(p1)
distance = diff.Length
_mcp_result["result"] = {
    "distance": distance,
    "point1": {"x": p1.x, "y": p1.y, "z": p1.z},
    "point2": {"x": p2.x, "y": p2.y, "z": p2.z},
    "delta": {"x": diff.x, "y": diff.y, "z": diff.z}
}
`
	return b.Run(code), nil
}

func hIeMeasureAngle(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	face1, _ := args["face1"].(string)
	face2, _ := args["face2"].(string)
	code := `
` + DocPreamble + `
import math
obj = doc.getObject(` + PyStr(objectName) + `)
if obj is None:
    raise ValueError("Object '"` + EscapePythonString(objectName) + `'" not found")
shape = obj.Shape
f1 = getattr(shape, ` + PyStr(face1) + `)
f2 = getattr(shape, ` + PyStr(face2) + `)
n1 = f1.normalAt(0, 0)
n2 = f2.normalAt(0, 0)
dot = n1.dot(n2)
dot = max(-1.0, min(1.0, dot))
angle_rad = math.acos(dot)
angle_deg = math.degrees(angle_rad)
_mcp_result["result"] = {
    "angle_degrees": angle_deg,
    "angle_radians": angle_rad,
    "normal1": {"x": n1.x, "y": n1.y, "z": n1.z},
    "normal2": {"x": n2.x, "y": n2.y, "z": n2.z}
}
`
	return b.Run(code), nil
}

func hIeGetVolume(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	code := `
` + DocPreamble + `
obj = doc.getObject(` + PyStr(objectName) + `)
if obj is None:
    raise ValueError("Object '"` + EscapePythonString(objectName) + `'" not found")
vol = obj.Shape.Volume
area = obj.Shape.Area
_mcp_result["result"] = {"objectName": ` + PyStr(objectName) + `, "volume_mm3": vol, "area_mm2": area}
`
	return b.Run(code), nil
}

func hIeGetBoundingBox(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	code := `
` + DocPreamble + `
obj = doc.getObject(` + PyStr(objectName) + `)
if obj is None:
    raise ValueError("Object '"` + EscapePythonString(objectName) + `'" not found")
bb = obj.Shape.BoundBox
_mcp_result["result"] = {
    "objectName": ` + PyStr(objectName) + `,
    "xMin": bb.XMin, "yMin": bb.YMin, "zMin": bb.ZMin,
    "xMax": bb.XMax, "yMax": bb.YMax, "zMax": bb.ZMax,
    "xLength": bb.XLength, "yLength": bb.YLength, "zLength": bb.ZLength,
    "center": {"x": bb.Center.x, "y": bb.Center.y, "z": bb.Center.z}
}
`
	return b.Run(code), nil
}

func hIeExecutePython(args map[string]any, b *Bridge) (Result, error) {
	code, _ := args["code"].(string)
	return b.Run(`
` + DocPreamble + `
` + code + `
`), nil
}

func hIeImportIges(args map[string]any, b *Bridge) (Result, error) {
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
import ImportGui
ImportGui.insert(` + PyStr(filePath) + `, doc.Name)
doc.recompute()
_mcp_result["result"] = {"document": doc.Name, "objects": [o.Name for o in doc.Objects], "file": ` + PyStr(filePath) + `}
`), nil
}

func hIeImportDxf(args map[string]any, b *Bridge) (Result, error) {
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
import importDXF
importDXF.insert(` + PyStr(filePath) + `, doc.Name)
doc.recompute()
_mcp_result["result"] = {"document": doc.Name, "objects": [o.Name for o in doc.Objects], "file": ` + PyStr(filePath) + `}
`), nil
}

func hIeImportSvg(args map[string]any, b *Bridge) (Result, error) {
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
import importSVG
importSVG.insert(` + PyStr(filePath) + `, doc.Name)
doc.recompute()
_mcp_result["result"] = {"document": doc.Name, "objects": [o.Name for o in doc.Objects], "file": ` + PyStr(filePath) + `}
`), nil
}

func hIeExportIges(args map[string]any, b *Bridge) (Result, error) {
	objectNames, err := ValidateArray(args["objectNames"], "objectNames")
	if err != nil {
		return Result{}, err
	}
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	code := `
` + DocPreamble + `
obj_names = ` + PyQuoteAny(objectNames) + `
objs = [doc.getObject(n) for n in obj_names]
missing = [n for n, o in zip(obj_names, objs) if o is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
Part.export(objs, ` + PyStr(filePath) + `)
import os
_mcp_result["result"] = {"exported": obj_names, "filePath": ` + PyStr(filePath) + `, "size_bytes": os.path.getsize(` + PyStr(filePath) + `)}
`
	return b.Run(code), nil
}

func hIeExportDxf(args map[string]any, b *Bridge) (Result, error) {
	objectNames, err := ValidateArray(args["objectNames"], "objectNames")
	if err != nil {
		return Result{}, err
	}
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	code := `
` + DocPreamble + `
import importDXF
obj_names = ` + PyQuoteAny(objectNames) + `
objs = [doc.getObject(n) for n in obj_names]
missing = [n for n, o in zip(obj_names, objs) if o is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
importDXF.export(objs, ` + PyStr(filePath) + `)
import os
_mcp_result["result"] = {"exported": obj_names, "filePath": ` + PyStr(filePath) + `, "size_bytes": os.path.getsize(` + PyStr(filePath) + `)}
`
	return b.Run(code), nil
}

func hIeExportSvg(args map[string]any, b *Bridge) (Result, error) {
	objectNames, err := ValidateArray(args["objectNames"], "objectNames")
	if err != nil {
		return Result{}, err
	}
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	direction := ArgStr(args, "direction", "Front")
	dirMap := map[string]string{
		"Front":     "FreeCAD.Vector(0, 0, -1)",
		"Top":       "FreeCAD.Vector(0, -1, 0)",
		"Right":     "FreeCAD.Vector(1, 0, 0)",
		"Isometric": "FreeCAD.Vector(1, 1, 1)",
	}
	dirVec := dirMap[direction]
	if dirVec == "" {
		dirVec = dirMap["Front"]
	}
	_ = dirVec
	code := `
` + DocPreamble + `
import importSVG
obj_names = ` + PyQuoteAny(objectNames) + `
objs = [doc.getObject(n) for n in obj_names]
missing = [n for n, o in zip(obj_names, objs) if o is None]
if missing:
    raise ValueError(f"Objects not found: {missing}")
importSVG.export(objs, ` + PyStr(filePath) + `)
import os
_mcp_result["result"] = {"exported": obj_names, "filePath": ` + PyStr(filePath) + `, "size_bytes": os.path.getsize(` + PyStr(filePath) + `)}
`
	return b.Run(code), nil
}

func hIeExportBrep(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	filePath, err := ValidateFilePath(args["filePath"], "filePath")
	if err != nil {
		return Result{}, err
	}
	code := `
` + DocPreamble + `
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
obj.Shape.exportBrep(` + PyStr(filePath) + `)
import os
_mcp_result["result"] = {"exported": ` + PyQuote(objectName) + `, "filePath": ` + PyStr(filePath) + `, "size_bytes": os.path.getsize(` + PyStr(filePath) + `)}
`
	return b.Run(code), nil
}
