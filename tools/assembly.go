package tools

func init() {
	register("assembly",
		ToolDef{Name: "freecad_assembly_create", Handler: hAsmCreate},
		ToolDef{Name: "freecad_assembly_insert_component", Handler: hAsmInsertComponent},
		ToolDef{Name: "freecad_assembly_toggle_grounded", Handler: hAsmToggleGrounded},
		ToolDef{Name: "freecad_assembly_fixed_joint", Handler: hAsmFixedJoint},
		ToolDef{Name: "freecad_assembly_revolute_joint", Handler: hAsmRevoluteJoint},
		ToolDef{Name: "freecad_assembly_slider_joint", Handler: hAsmSliderJoint},
		ToolDef{Name: "freecad_assembly_cylindrical_joint", Handler: hAsmCylindricalJoint},
		ToolDef{Name: "freecad_assembly_ball_joint", Handler: hAsmBallJoint},
		ToolDef{Name: "freecad_assembly_distance_joint", Handler: hAsmDistanceJoint},
		ToolDef{Name: "freecad_assembly_solve", Handler: hAsmSolve},
	)
}

// makeJointCode mirrors the TS makeJointCode helper.
func makeJointCode(jointType, assemblyName, comp1, elem1, comp2, elem2, jointName, extraProps string) string {
	return `
` + DocPreamble + `
assembly = doc.getObject(` + PyQuote(assemblyName) + `)
if assembly is None:
    raise ValueError("Assembly not found: ` + assemblyName + `")
c1 = doc.getObject(` + PyQuote(comp1) + `)
c2 = doc.getObject(` + PyQuote(comp2) + `)
if c1 is None:
    raise ValueError("Component not found: ` + comp1 + `")
if c2 is None:
    raise ValueError("Component not found: ` + comp2 + `")
joint = doc.addObject("Assembly::Joint` + jointType + `", ` + PyQuote(jointName) + `)
joint.ObjectToGround = c1
joint.Part1 = c1
joint.Element1 = ` + PyQuote(elem1) + `
joint.Part2 = c2
joint.Element2 = ` + PyQuote(elem2) + `
` + extraProps + `
assembly.addObject(joint)
doc.recompute()
_mcp_result["result"] = {"name": joint.Name, "type": "` + jointType + `", "component1": ` + PyQuote(comp1) + `, "component2": ` + PyQuote(comp2) + `}
`
}

func hAsmCreate(args map[string]any, b *Bridge) (Result, error) {
	asmName := ArgStr(args, "name", "Assembly")
	return b.Run(`
` + DocPreamble + `
assembly = doc.addObject("Assembly::AssemblyObject", ` + PyQuote(asmName) + `)
doc.recompute()
_mcp_result["result"] = {"name": assembly.Name, "label": assembly.Label}
`), nil
}

func hAsmInsertComponent(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	objectName, _ := args["objectName"].(string)
	x := ArgNum(args, "x", 0)
	y := ArgNum(args, "y", 0)
	z := ArgNum(args, "z", 0)
	return b.Run(`
` + DocPreamble + `
assembly = doc.getObject(` + PyQuote(assemblyName) + `)
if assembly is None:
    raise ValueError("Assembly not found: ` + assemblyName + `")
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
link = doc.addObject("App::Link", ` + PyQuote(objectName+"_Link") + `)
link.LinkedObject = obj
link.Placement.Base = FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)
assembly.addObject(link)
doc.recompute()
_mcp_result["result"] = {"name": link.Name, "linkedTo": obj.Name, "position": {"x": ` + PyNum(x) + `, "y": ` + PyNum(y) + `, "z": ` + PyNum(z) + `}}
`), nil
}

func hAsmToggleGrounded(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	componentName, _ := args["componentName"].(string)
	return b.Run(`
` + DocPreamble + `
assembly = doc.getObject(` + PyQuote(assemblyName) + `)
if assembly is None:
    raise ValueError("Assembly not found: ` + assemblyName + `")
comp = doc.getObject(` + PyQuote(componentName) + `)
if comp is None:
    raise ValueError("Component not found: ` + componentName + `")
# Create a grounded joint
grounded = doc.addObject("Assembly::JointFixed", "Grounded_" + comp.Name)
grounded.ObjectToGround = comp
assembly.addObject(grounded)
doc.recompute()
_mcp_result["result"] = {"name": grounded.Name, "component": comp.Name, "grounded": True}
`), nil
}

func hAsmFixedJoint(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	comp1, _ := args["component1"].(string)
	elem1, _ := args["element1"].(string)
	comp2, _ := args["component2"].(string)
	elem2, _ := args["element2"].(string)
	jName := ArgStr(args, "name", "FixedJoint")
	return b.Run(makeJointCode("Fixed", assemblyName, comp1, elem1, comp2, elem2, jName, "")), nil
}

func hAsmRevoluteJoint(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	comp1, _ := args["component1"].(string)
	elem1, _ := args["element1"].(string)
	comp2, _ := args["component2"].(string)
	elem2, _ := args["element2"].(string)
	jName := ArgStr(args, "name", "RevoluteJoint")
	return b.Run(makeJointCode("Revolute", assemblyName, comp1, elem1, comp2, elem2, jName, "")), nil
}

func hAsmSliderJoint(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	comp1, _ := args["component1"].(string)
	elem1, _ := args["element1"].(string)
	comp2, _ := args["component2"].(string)
	elem2, _ := args["element2"].(string)
	jName := ArgStr(args, "name", "SliderJoint")
	return b.Run(makeJointCode("Slider", assemblyName, comp1, elem1, comp2, elem2, jName, "")), nil
}

func hAsmCylindricalJoint(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	comp1, _ := args["component1"].(string)
	elem1, _ := args["element1"].(string)
	comp2, _ := args["component2"].(string)
	elem2, _ := args["element2"].(string)
	jName := ArgStr(args, "name", "CylindricalJoint")
	return b.Run(makeJointCode("Cylindrical", assemblyName, comp1, elem1, comp2, elem2, jName, "")), nil
}

func hAsmBallJoint(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	comp1, _ := args["component1"].(string)
	elem1, _ := args["element1"].(string)
	comp2, _ := args["component2"].(string)
	elem2, _ := args["element2"].(string)
	jName := ArgStr(args, "name", "BallJoint")
	return b.Run(makeJointCode("Ball", assemblyName, comp1, elem1, comp2, elem2, jName, "")), nil
}

func hAsmDistanceJoint(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	comp1, _ := args["component1"].(string)
	elem1, _ := args["element1"].(string)
	comp2, _ := args["component2"].(string)
	elem2, _ := args["element2"].(string)
	distance, _ := args["distance"].(float64)
	jName := ArgStr(args, "name", "DistanceJoint")
	return b.Run(makeJointCode("Distance", assemblyName, comp1, elem1, comp2, elem2, jName, `joint.Distance = `+PyNum(distance))), nil
}

func hAsmSolve(args map[string]any, b *Bridge) (Result, error) {
	assemblyName, _ := args["assemblyName"].(string)
	return b.Run(`
` + DocPreamble + `
assembly = doc.getObject(` + PyQuote(assemblyName) + `)
if assembly is None:
    raise ValueError("Assembly not found: ` + assemblyName + `")
doc.recompute()
joints = [o for o in assembly.Group if "Joint" in o.TypeId]
components = [o for o in assembly.Group if o.TypeId == "App::Link"]
_mcp_result["result"] = {"assembly": assembly.Name, "joints": len(joints), "components": len(components), "status": "solved"}
`), nil
}
