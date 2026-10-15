package tools

func init() {
	register("mesh",
		ToolDef{Name: "freecad_mesh_from_shape", Handler: hMeshFromShape},
		ToolDef{Name: "freecad_mesh_to_shape", Handler: hMeshToShape},
		ToolDef{Name: "freecad_mesh_repair", Handler: hMeshRepair},
		ToolDef{Name: "freecad_mesh_decimate", Handler: hMeshDecimate},
		ToolDef{Name: "freecad_mesh_refine", Handler: hMeshRefine},
		ToolDef{Name: "freecad_mesh_info", Handler: hMeshInfo},
		ToolDef{Name: "freecad_mesh_boolean", Handler: hMeshBoolean},
	)
}

func hMeshFromShape(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	linearDeflection, err := ValidatePositiveNumber(ArgNum(args, "linearDeflection", 0.1), "linearDeflection", 1000)
	if err != nil {
		return Result{}, err
	}
	angularDeflection, err := ValidatePositiveNumber(ArgNum(args, "angularDeflection", 30), "angularDeflection", 180)
	if err != nil {
		return Result{}, err
	}
	meshName := ArgStr(args, "name", "Mesh")
	return b.Run(`
` + DocPreamble + `
import Mesh, MeshPart
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
shape = obj.Shape
mesh_data = MeshPart.meshFromShape(Shape=shape, LinearDeflection=` + PyNum(linearDeflection) + `, AngularDeflection=` + PyNum(angularDeflection) + `)
mesh_obj = doc.addObject("Mesh::Feature", ` + PyQuote(meshName) + `)
mesh_obj.Mesh = mesh_data
doc.recompute()
m = mesh_obj.Mesh
_mcp_result["result"] = {"name": mesh_obj.Name, "vertices": m.CountPoints, "facets": m.CountFacets, "area": m.Area}
`), nil
}

func hMeshToShape(args map[string]any, b *Bridge) (Result, error) {
	meshName, _ := args["meshName"].(string)
	sewing := ArgBool(args, "sewing", true)
	shapeName := ArgStr(args, "name", "Shape")
	sewingStr := "False"
	if sewing {
		sewingStr = "True"
	}
	return b.Run(`
` + DocPreamble + `
mesh_obj = doc.getObject(` + PyQuote(meshName) + `)
if mesh_obj is None:
    raise ValueError("Mesh not found: ` + meshName + `")
shape = Part.Shape()
shape.makeShapeFromMesh(mesh_obj.Mesh.Topology, 0.1)
if ` + sewingStr + `:
    solid = Part.makeSolid(shape)
    shape_obj = doc.addObject("Part::Feature", ` + PyQuote(shapeName) + `)
    shape_obj.Shape = solid
else:
    shape_obj = doc.addObject("Part::Feature", ` + PyQuote(shapeName) + `)
    shape_obj.Shape = shape
doc.recompute()
_mcp_result["result"] = {"name": shape_obj.Name, "volume": shape_obj.Shape.Volume, "area": shape_obj.Shape.Area}
`), nil
}

func hMeshRepair(args map[string]any, b *Bridge) (Result, error) {
	meshName, _ := args["meshName"].(string)
	fillHoles := ArgBool(args, "fillHoles", true)
	removeDuplicates := ArgBool(args, "removeDuplicates", true)
	fixNormals := ArgBool(args, "fixNormals", true)
	removeDuplicatesStr := "False"
	if removeDuplicates {
		removeDuplicatesStr = "True"
	}
	fixNormalsStr := "False"
	if fixNormals {
		fixNormalsStr = "True"
	}
	fillHolesStr := "False"
	if fillHoles {
		fillHolesStr = "True"
	}
	return b.Run(`
` + DocPreamble + `
mesh_obj = doc.getObject(` + PyQuote(meshName) + `)
if mesh_obj is None:
    raise ValueError("Mesh not found: ` + meshName + `")
m = mesh_obj.Mesh.copy()
before_facets = m.CountFacets
before_points = m.CountPoints
if ` + removeDuplicatesStr + `:
    m.removeDuplicatedPoints()
    m.removeDuplicatedFacets()
if ` + fixNormalsStr + `:
    m.harmonizeNormals()
    m.fixIndices()
if ` + fillHolesStr + `:
    m.fillupHoles(1000)
mesh_obj.Mesh = m
doc.recompute()
_mcp_result["result"] = {
    "name": mesh_obj.Name,
    "before": {"facets": before_facets, "vertices": before_points},
    "after": {"facets": m.CountFacets, "vertices": m.CountPoints}
}
`), nil
}

func hMeshDecimate(args map[string]any, b *Bridge) (Result, error) {
	meshName, _ := args["meshName"].(string)
	targetReduction, err := ValidateNumber(args["targetReduction"], "targetReduction", 0.01, 0.99)
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
mesh_obj = doc.getObject(` + PyQuote(meshName) + `)
if mesh_obj is None:
    raise ValueError("Mesh not found: ` + meshName + `")
m = mesh_obj.Mesh.copy()
before = m.CountFacets
target = int(before * (1.0 - ` + PyNum(targetReduction) + `))
m.decimate(target)
mesh_obj.Mesh = m
doc.recompute()
_mcp_result["result"] = {
    "name": mesh_obj.Name,
    "facets_before": before,
    "facets_after": m.CountFacets,
    "reduction": round(1.0 - m.CountFacets / before, 3) if before > 0 else 0
}
`), nil
}

func hMeshRefine(args map[string]any, b *Bridge) (Result, error) {
	meshName, _ := args["meshName"].(string)
	maxEdgeLength, err := ValidatePositiveNumber(args["maxEdgeLength"], "maxEdgeLength")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
mesh_obj = doc.getObject(` + PyQuote(meshName) + `)
if mesh_obj is None:
    raise ValueError("Mesh not found: ` + meshName + `")
m = mesh_obj.Mesh.copy()
before = m.CountFacets
m.refine()
mesh_obj.Mesh = m
doc.recompute()
_mcp_result["result"] = {
    "name": mesh_obj.Name,
    "facets_before": before,
    "facets_after": m.CountFacets,
    "maxEdgeLength": ` + PyNum(maxEdgeLength) + `
}
`), nil
}

func hMeshInfo(args map[string]any, b *Bridge) (Result, error) {
	meshName, _ := args["meshName"].(string)
	return b.Run(`
` + DocPreamble + `
mesh_obj = doc.getObject(` + PyQuote(meshName) + `)
if mesh_obj is None:
    raise ValueError("Mesh not found: ` + meshName + `")
m = mesh_obj.Mesh
bb = m.BoundBox
_mcp_result["result"] = {
    "name": mesh_obj.Name,
    "vertices": m.CountPoints,
    "facets": m.CountFacets,
    "edges": m.CountEdges,
    "area": m.Area,
    "volume": m.Volume,
    "isSolid": m.isSolid(),
    "hasNonManifolds": m.hasNonManifolds(),
    "hasSelfIntersections": m.hasSelfIntersections(),
    "boundingBox": {
        "xMin": bb.XMin, "yMin": bb.YMin, "zMin": bb.ZMin,
        "xMax": bb.XMax, "yMax": bb.YMax, "zMax": bb.ZMax,
        "xLength": bb.XLength, "yLength": bb.YLength, "zLength": bb.ZLength,
    }
}
`), nil
}

func hMeshBoolean(args map[string]any, b *Bridge) (Result, error) {
	mesh1Name, _ := args["mesh1Name"].(string)
	mesh2Name, _ := args["mesh2Name"].(string)
	operation, _ := args["operation"].(string)
	resultName := ArgStr(args, "name", "MeshBool")
	opMethod := "unite"
	if operation == "intersection" {
		opMethod = "intersect"
	} else if operation == "difference" {
		opMethod = "difference"
	}
	return b.Run(`
` + DocPreamble + `
import Mesh
m1_obj = doc.getObject(` + PyQuote(mesh1Name) + `)
m2_obj = doc.getObject(` + PyQuote(mesh2Name) + `)
if m1_obj is None:
    raise ValueError("Mesh not found: ` + mesh1Name + `")
if m2_obj is None:
    raise ValueError("Mesh not found: ` + mesh2Name + `")
m1 = m1_obj.Mesh.copy()
m2 = m2_obj.Mesh.copy()
result = Mesh.Mesh()
result.addMesh(m1)
result.` + opMethod + `(m2)
result_obj = doc.addObject("Mesh::Feature", ` + PyQuote(resultName) + `)
result_obj.Mesh = result
doc.recompute()
_mcp_result["result"] = {"name": result_obj.Name, "operation": ` + PyQuote(operation) + `, "facets": result.CountFacets}
`), nil
}
