package tools

func init() {
	register("fem",
		ToolDef{Name: "freecad_fem_analysis", Handler: hFemAnalysis},
		ToolDef{Name: "freecad_fem_material", Handler: hFemMaterial},
		ToolDef{Name: "freecad_fem_constraint_fixed", Handler: hFemConstraintFixed},
		ToolDef{Name: "freecad_fem_constraint_force", Handler: hFemConstraintForce},
		ToolDef{Name: "freecad_fem_constraint_pressure", Handler: hFemConstraintPressure},
		ToolDef{Name: "freecad_fem_mesh", Handler: hFemMesh},
		ToolDef{Name: "freecad_fem_solver", Handler: hFemSolver},
		ToolDef{Name: "freecad_fem_results", Handler: hFemResults},
	)
}

func hFemAnalysis(args map[string]any, b *Bridge) (Result, error) {
	analysisName := ArgStr(args, "name", "Analysis")
	return b.Run(`
` + DocPreamble + `
import ObjectsFem
analysis = ObjectsFem.makeAnalysis(doc, ` + PyQuote(analysisName) + `)
doc.recompute()
_mcp_result["result"] = {"name": analysis.Name, "label": analysis.Label}
`), nil
}

func hFemMaterial(args map[string]any, b *Bridge) (Result, error) {
	analysisName, _ := args["analysisName"].(string)
	materialName := ArgStr(args, "materialName", "Steel")
	matObjName := ArgStr(args, "name", "Material")

	materialPresets := map[string]string{
		"Steel":    `mat_obj.Material = {"Name": "Steel", "YoungsModulus": "210000 MPa", "PoissonRatio": "0.3", "Density": "7900 kg/m^3"}`,
		"Aluminum": `mat_obj.Material = {"Name": "Aluminum", "YoungsModulus": "70000 MPa", "PoissonRatio": "0.33", "Density": "2700 kg/m^3"}`,
		"Concrete": `mat_obj.Material = {"Name": "Concrete", "YoungsModulus": "30000 MPa", "PoissonRatio": "0.2", "Density": "2400 kg/m^3"}`,
	}

	youngsModulus, hasYoungsModulus := args["youngsModulus"].(float64)
	poissonRatio := ArgNum(args, "poissonRatio", 0.3)
	density, hasDensity := args["density"].(float64)

	var matCode string
	if materialName == "Custom" && hasYoungsModulus && hasDensity {
		matCode = `mat_obj.Material = {"Name": "Custom", "YoungsModulus": "` + PyNum(youngsModulus) + ` MPa", "PoissonRatio": "` + PyNum(poissonRatio) + `", "Density": "` + PyNum(density) + ` kg/m^3"}`
	} else {
		preset, ok := materialPresets[materialName]
		if !ok {
			preset = materialPresets["Steel"]
		}
		matCode = preset
	}

	return b.Run(`
` + DocPreamble + `
import ObjectsFem
analysis = doc.getObject(` + PyQuote(analysisName) + `)
if analysis is None:
    raise ValueError("Analysis not found: ` + analysisName + `")
mat_obj = ObjectsFem.makeMaterialSolid(doc, ` + PyQuote(matObjName) + `)
` + matCode + `
analysis.addObject(mat_obj)
doc.recompute()
_mcp_result["result"] = {"name": mat_obj.Name, "material": ` + PyQuote(materialName) + `}
`), nil
}

func hFemConstraintFixed(args map[string]any, b *Bridge) (Result, error) {
	analysisName, _ := args["analysisName"].(string)
	objectName, _ := args["objectName"].(string)
	references, _ := args["references"].([]any)
	constraintName := ArgStr(args, "name", "ConstraintFixed")
	return b.Run(`
` + DocPreamble + `
import ObjectsFem
analysis = doc.getObject(` + PyQuote(analysisName) + `)
if analysis is None:
    raise ValueError("Analysis not found: ` + analysisName + `")
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
fixed = ObjectsFem.makeConstraintFixed(doc, ` + PyQuote(constraintName) + `)
fixed.References = [(obj, ` + PyQuoteAny(references) + `)]
analysis.addObject(fixed)
doc.recompute()
_mcp_result["result"] = {"name": fixed.Name, "references": ` + PyQuoteAny(references) + `}
`), nil
}

func hFemConstraintForce(args map[string]any, b *Bridge) (Result, error) {
	analysisName, _ := args["analysisName"].(string)
	objectName, _ := args["objectName"].(string)
	references, _ := args["references"].([]any)
	force, _ := args["force"].(float64)
	// TS computes these defaults but never uses them in the template (Direction is None).
	_, _, _ = ArgNum(args, "directionX", 0), ArgNum(args, "directionY", 0), ArgNum(args, "directionZ", -1)
	constraintName := ArgStr(args, "name", "ConstraintForce")
	return b.Run(`
` + DocPreamble + `
import ObjectsFem
analysis = doc.getObject(` + PyQuote(analysisName) + `)
if analysis is None:
    raise ValueError("Analysis not found: ` + analysisName + `")
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
force_obj = ObjectsFem.makeConstraintForce(doc, ` + PyQuote(constraintName) + `)
force_obj.References = [(obj, ` + PyQuoteAny(references) + `)]
force_obj.Force = ` + PyNum(force) + `
force_obj.Direction = None
force_obj.Reversed = False
analysis.addObject(force_obj)
doc.recompute()
_mcp_result["result"] = {"name": force_obj.Name, "force": ` + PyNum(force) + `, "references": ` + PyQuoteAny(references) + `}
`), nil
}

func hFemConstraintPressure(args map[string]any, b *Bridge) (Result, error) {
	analysisName, _ := args["analysisName"].(string)
	objectName, _ := args["objectName"].(string)
	references, _ := args["references"].([]any)
	pressure, _ := args["pressure"].(float64)
	reversed := ArgBool(args, "reversed", false)
	constraintName := ArgStr(args, "name", "ConstraintPressure")
	reversedPy := "False"
	if reversed {
		reversedPy = "True"
	}
	return b.Run(`
` + DocPreamble + `
import ObjectsFem
analysis = doc.getObject(` + PyQuote(analysisName) + `)
if analysis is None:
    raise ValueError("Analysis not found: ` + analysisName + `")
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
pressure_obj = ObjectsFem.makeConstraintPressure(doc, ` + PyQuote(constraintName) + `)
pressure_obj.References = [(obj, ` + PyQuoteAny(references) + `)]
pressure_obj.Pressure = ` + PyNum(pressure) + `
pressure_obj.Reversed = ` + reversedPy + `
analysis.addObject(pressure_obj)
doc.recompute()
_mcp_result["result"] = {"name": pressure_obj.Name, "pressure": ` + PyNum(pressure) + `, "references": ` + PyQuoteAny(references) + `}
`), nil
}

func hFemMesh(args map[string]any, b *Bridge) (Result, error) {
	analysisName, _ := args["analysisName"].(string)
	objectName, _ := args["objectName"].(string)
	maxSize := ArgNum(args, "maxElementSize", 0)
	minSize := ArgNum(args, "minElementSize", 0)
	meshOrder := ArgNum(args, "meshOrder", 2)
	meshName := ArgStr(args, "name", "FEMMesh")
	maxCode := ""
	if maxSize > 0 {
		maxCode = `mesh.CharacteristicLengthMax = ` + PyNum(maxSize)
	}
	minCode := ""
	if minSize > 0 {
		minCode = `mesh.CharacteristicLengthMin = ` + PyNum(minSize)
	}
	elementOrder := "2nd"
	if meshOrder == 1 {
		elementOrder = "1st"
	}
	return b.Run(`
` + DocPreamble + `
import ObjectsFem
analysis = doc.getObject(` + PyQuote(analysisName) + `)
if analysis is None:
    raise ValueError("Analysis not found: ` + analysisName + `")
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
mesh = ObjectsFem.makeMeshGmsh(doc, ` + PyQuote(meshName) + `)
mesh.Part = obj
` + maxCode + `
` + minCode + `
mesh.ElementOrder = ` + PyQuote(elementOrder) + `
analysis.addObject(mesh)
doc.recompute()
# Try to compute the mesh
from femmesh.gmshtools import GmshTools
gmsh = GmshTools(mesh)
error = gmsh.create_mesh()
doc.recompute()
fem_mesh = mesh.FemMesh
_mcp_result["result"] = {
    "name": mesh.Name,
    "nodes": fem_mesh.NodeCount,
    "elements": fem_mesh.ElementCount,
    "volumes": fem_mesh.VolumeCount,
    "order": ` + PyNum(meshOrder) + `
}
`), nil
}

func hFemSolver(args map[string]any, b *Bridge) (Result, error) {
	analysisName, _ := args["analysisName"].(string)
	solver := ArgStr(args, "solver", "calculix")
	analysisType := ArgStr(args, "analysisType", "static")
	run := ArgBool(args, "run", false)
	solverName := ArgStr(args, "name", "Solver")
	solverCode := `solver = ObjectsFem.makeSolverElmer(doc, ` + PyQuote(solverName) + `)`
	if solver == "calculix" {
		solverCode = `solver = ObjectsFem.makeSolverCalculixCcxTools(doc, ` + PyQuote(solverName) + `)`
	}
	runCode := ""
	if run {
		runCode = `
from femtools import ccxtools
fea = ccxtools.FemToolsCcx(analysis, solver)
fea.purge_results()
fea.run()
result["run"] = True
result["status"] = "completed"
`
	}
	return b.Run(`
` + DocPreamble + `
import ObjectsFem
analysis = doc.getObject(` + PyQuote(analysisName) + `)
if analysis is None:
    raise ValueError("Analysis not found: ` + analysisName + `")
` + solverCode + `
solver.AnalysisType = ` + PyQuote(analysisType) + `
analysis.addObject(solver)
doc.recompute()
result = {"name": solver.Name, "solver": ` + PyQuote(solver) + `, "analysisType": ` + PyQuote(analysisType) + `, "run": False}
` + runCode + `
_mcp_result["result"] = result
`), nil
}

func hFemResults(args map[string]any, b *Bridge) (Result, error) {
	analysisName, _ := args["analysisName"].(string)
	return b.Run(`
` + DocPreamble + `
analysis = doc.getObject(` + PyQuote(analysisName) + `)
if analysis is None:
    raise ValueError("Analysis not found: ` + analysisName + `")
# Find result objects in the analysis
results = []
for obj in analysis.Group:
    if hasattr(obj, "Stats"):
        stats = obj.Stats
        results.append({
            "name": obj.Name,
            "type": obj.TypeId,
            "stats": {
                "maxDisplacement": stats[0] if len(stats) > 0 else None,
                "minStress": stats[1] if len(stats) > 1 else None,
                "maxStress": stats[2] if len(stats) > 2 else None,
                "maxDisplacementAbs": stats[4] if len(stats) > 4 else None,
            }
        })
    elif obj.TypeId == "Fem::FemResultObject":
        res = {
            "name": obj.Name,
            "type": obj.TypeId,
        }
        if hasattr(obj, "DisplacementLengths") and obj.DisplacementLengths:
            res["maxDisplacement"] = max(obj.DisplacementLengths)
        if hasattr(obj, "vonMises") and obj.vonMises:
            res["maxVonMises"] = max(obj.vonMises)
            res["minVonMises"] = min(obj.vonMises)
        results.append(res)
_mcp_result["result"] = {"analysis": analysis.Name, "resultCount": len(results), "results": results}
`), nil
}
