package tools

func init() {
	register("document",
		ToolDef{Name: "freecad_new_document", Handler: hNewDocument},
		ToolDef{Name: "freecad_open_document", Handler: hOpenDocument},
		ToolDef{Name: "freecad_save_document", Handler: hSaveDocument},
		ToolDef{Name: "freecad_close_document", Handler: hCloseDocument},
		ToolDef{Name: "freecad_list_objects", Handler: hListObjects},
		ToolDef{Name: "freecad_get_object_info", Handler: hGetObjectInfo},
	)
}

func hNewDocument(args map[string]any, b *Bridge) (Result, error) {
	docName, _ := args["name"].(string)
	return b.Run(`
doc = FreeCAD.newDocument(` + PyQuote(docName) + `)
_mcp_result["result"] = {"name": doc.Name}
`), nil
}

func hOpenDocument(args map[string]any, b *Bridge) (Result, error) {
	filePath, ok := args["filePath"].(string)
	if !ok {
		filePath = ""
	}
	return b.Run(`
doc = FreeCAD.openDocument(` + PyQuote(filePath) + `)
_mcp_result["result"] = {"name": doc.Name, "objects": len(doc.Objects)}
`), nil
}

func hSaveDocument(args map[string]any, b *Bridge) (Result, error) {
	filePath, _ := args["filePath"].(string)
	if filePath != "" {
		return b.Run(`
doc = FreeCAD.ActiveDocument
doc.saveAs(` + PyQuote(filePath) + `)
_mcp_result["result"] = {"saved": ` + PyQuote(filePath) + `}
`), nil
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
doc.save()
_mcp_result["result"] = {"saved": doc.FileName}
`), nil
}

func hCloseDocument(args map[string]any, b *Bridge) (Result, error) {
	docName, _ := args["name"].(string)
	return b.Run(`
FreeCAD.closeDocument(` + PyQuote(docName) + `)
_mcp_result["result"] = {"closed": ` + PyQuote(docName) + `}
`), nil
}

func hListObjects(args map[string]any, b *Bridge) (Result, error) {
	docName, _ := args["documentName"].(string)
	getDoc := `doc = FreeCAD.ActiveDocument`
	if docName != "" {
		getDoc = `doc = FreeCAD.getDocument(` + PyQuote(docName) + `)`
	}
	return b.Run(`
` + getDoc + `
objects = []
for obj in doc.Objects:
    objects.append({"name": obj.Name, "label": obj.Label, "typeName": obj.TypeId})
_mcp_result["result"] = {"document": doc.Name, "objects": objects}
`), nil
}

func hGetObjectInfo(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	docName, _ := args["documentName"].(string)
	getDoc := `doc = FreeCAD.ActiveDocument`
	if docName != "" {
		getDoc = `doc = FreeCAD.getDocument(` + PyQuote(docName) + `)`
	}
	return b.Run(`
` + getDoc + `
obj = doc.getObject(` + PyQuote(objectName) + `)
info = {"name": obj.Name, "label": obj.Label, "typeName": obj.TypeId}
placement = obj.Placement
info["placement"] = {
    "position": {"x": placement.Base.x, "y": placement.Base.y, "z": placement.Base.z},
    "rotation": {"angle": placement.Rotation.Angle, "axis": {"x": placement.Rotation.Axis.x, "y": placement.Rotation.Axis.y, "z": placement.Rotation.Axis.z}}
}
if hasattr(obj, "Shape"):
    shape = obj.Shape
    bb = shape.BoundBox
    info["shape"] = {
        "volume": shape.Volume,
        "area": shape.Area,
        "boundingBox": {"xMin": bb.XMin, "yMin": bb.YMin, "zMin": bb.ZMin, "xMax": bb.XMax, "yMax": bb.YMax, "zMax": bb.ZMax}
    }
_mcp_result["result"] = info
`), nil
}
