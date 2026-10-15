package tools

import (
	"errors"
)

func init() {
	register("primitives",
		ToolDef{Name: "freecad_create_box", Handler: hPrimCreateBox},
		ToolDef{Name: "freecad_create_cylinder", Handler: hPrimCreateCylinder},
		ToolDef{Name: "freecad_create_sphere", Handler: hPrimCreateSphere},
		ToolDef{Name: "freecad_create_cone", Handler: hPrimCreateCone},
		ToolDef{Name: "freecad_create_torus", Handler: hPrimCreateTorus},
		ToolDef{Name: "freecad_create_wedge", Handler: hPrimCreateWedge},
		ToolDef{Name: "freecad_create_helix", Handler: hPrimCreateHelix},
		ToolDef{Name: "freecad_create_spiral", Handler: hPrimCreateSpiral},
		ToolDef{Name: "freecad_create_ellipsoid", Handler: hPrimCreateEllipsoid},
		ToolDef{Name: "freecad_create_tube", Handler: hPrimCreateTube},
		ToolDef{Name: "freecad_create_prism", Handler: hPrimCreatePrism},
	)
}

func hPrimCreateBox(args map[string]any, b *Bridge) (Result, error) {
	objName := ArgStr(args, "name", "Box")
	length, err := ValidatePositiveNumber(args["length"], "length")
	if err != nil {
		return Result{}, err
	}
	width, err := ValidatePositiveNumber(args["width"], "width")
	if err != nil {
		return Result{}, err
	}
	height, err := ValidatePositiveNumber(args["height"], "height")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
box = doc.addObject("Part::Box", ` + PyStr(objName) + `)
box.Length = ` + PyNum(length) + `
box.Width = ` + PyNum(width) + `
box.Height = ` + PyNum(height) + `
doc.recompute()
_mcp_result["result"] = {"name": box.Name, "label": box.Label}
`), nil
}

func hPrimCreateCylinder(args map[string]any, b *Bridge) (Result, error) {
	objName := ArgStr(args, "name", "Cylinder")
	radius, err := ValidatePositiveNumber(args["radius"], "radius")
	if err != nil {
		return Result{}, err
	}
	height, err := ValidatePositiveNumber(args["height"], "height")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
cyl = doc.addObject("Part::Cylinder", ` + PyStr(objName) + `)
cyl.Radius = ` + PyNum(radius) + `
cyl.Height = ` + PyNum(height) + `
doc.recompute()
_mcp_result["result"] = {"name": cyl.Name, "label": cyl.Label}
`), nil
}

func hPrimCreateSphere(args map[string]any, b *Bridge) (Result, error) {
	objName := ArgStr(args, "name", "Sphere")
	radius, err := ValidatePositiveNumber(args["radius"], "radius")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
sph = doc.addObject("Part::Sphere", ` + PyStr(objName) + `)
sph.Radius = ` + PyNum(radius) + `
doc.recompute()
_mcp_result["result"] = {"name": sph.Name, "label": sph.Label}
`), nil
}

func hPrimCreateCone(args map[string]any, b *Bridge) (Result, error) {
	objName := ArgStr(args, "name", "Cone")
	radius1, err := ValidateNonNegativeNumber(args["radius1"], "radius1")
	if err != nil {
		return Result{}, err
	}
	radius2, err := ValidateNonNegativeNumber(args["radius2"], "radius2")
	if err != nil {
		return Result{}, err
	}
	height, err := ValidatePositiveNumber(args["height"], "height")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
cone = doc.addObject("Part::Cone", ` + PyStr(objName) + `)
cone.Radius1 = ` + PyNum(radius1) + `
cone.Radius2 = ` + PyNum(radius2) + `
cone.Height = ` + PyNum(height) + `
doc.recompute()
_mcp_result["result"] = {"name": cone.Name, "label": cone.Label}
`), nil
}

func hPrimCreateTorus(args map[string]any, b *Bridge) (Result, error) {
	objName := ArgStr(args, "name", "Torus")
	radius1, err := ValidatePositiveNumber(args["radius1"], "radius1")
	if err != nil {
		return Result{}, err
	}
	radius2, err := ValidatePositiveNumber(args["radius2"], "radius2")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
tor = doc.addObject("Part::Torus", ` + PyStr(objName) + `)
tor.Radius1 = ` + PyNum(radius1) + `
tor.Radius2 = ` + PyNum(radius2) + `
doc.recompute()
_mcp_result["result"] = {"name": tor.Name, "label": tor.Label}
`), nil
}

func hPrimCreateWedge(args map[string]any, b *Bridge) (Result, error) {
	objName := ArgStr(args, "name", "Wedge")
	keys := []string{"xmin", "ymin", "zmin", "x2min", "z2min", "xmax", "ymax", "zmax", "x2max", "z2max"}
	vals := make(map[string]float64, len(keys))
	for _, key := range keys {
		v, err := ValidateNumber(args[key], key, -1e6, 1e6)
		if err != nil {
			return Result{}, err
		}
		vals[key] = v
	}
	return b.Run(`
` + DocPreamble + `
wedge_shape = Part.makeWedge(` + PyNum(vals["xmin"]) + `, ` + PyNum(vals["ymin"]) + `, ` + PyNum(vals["zmin"]) + `, ` + PyNum(vals["x2min"]) + `, ` + PyNum(vals["z2min"]) + `, ` + PyNum(vals["xmax"]) + `, ` + PyNum(vals["ymax"]) + `, ` + PyNum(vals["zmax"]) + `, ` + PyNum(vals["x2max"]) + `, ` + PyNum(vals["z2max"]) + `)
wedge_obj = doc.addObject("Part::Feature", ` + PyStr(objName) + `)
wedge_obj.Shape = wedge_shape
doc.recompute()
_mcp_result["result"] = {"name": wedge_obj.Name, "label": wedge_obj.Label}
`), nil
}

func hPrimCreateHelix(args map[string]any, b *Bridge) (Result, error) {
	objName := EscapePythonString(ArgStr(args, "name", "Helix"))
	pitch, err := ValidatePositiveNumber(args["pitch"], "pitch")
	if err != nil {
		return Result{}, err
	}
	height, err := ValidatePositiveNumber(args["height"], "height")
	if err != nil {
		return Result{}, err
	}
	radius, err := ValidatePositiveNumber(args["radius"], "radius")
	if err != nil {
		return Result{}, err
	}
	angleRaw := any(float64(0))
	if v, ok := args["angle"]; ok && v != nil {
		angleRaw = v
	}
	angle, err := ValidateNumber(angleRaw, "angle", -89, 89)
	if err != nil {
		return Result{}, err
	}
	leftHanded := ArgBool(args, "leftHanded", false)
	localCoord := "0"
	if leftHanded {
		localCoord = "1"
	}
	return b.Run(`
` + DocPreamble + `
helix = doc.addObject("Part::Helix", ` + PyQuote(objName) + `)
helix.Pitch = ` + PyNum(pitch) + `
helix.Height = ` + PyNum(height) + `
helix.Radius = ` + PyNum(radius) + `
helix.Angle = ` + PyNum(angle) + `
helix.LocalCoord = ` + localCoord + `
doc.recompute()
_mcp_result["result"] = {"name": helix.Name, "label": helix.Label, "pitch": ` + PyNum(pitch) + `, "height": ` + PyNum(height) + `, "radius": ` + PyNum(radius) + `}
`), nil
}

func hPrimCreateSpiral(args map[string]any, b *Bridge) (Result, error) {
	objName := EscapePythonString(ArgStr(args, "name", "Spiral"))
	growth, err := ValidatePositiveNumber(args["growth"], "growth")
	if err != nil {
		return Result{}, err
	}
	rotations, err := ValidatePositiveNumber(args["rotations"], "rotations", 10000)
	if err != nil {
		return Result{}, err
	}
	radius, err := ValidateNonNegativeNumber(args["radius"], "radius")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
spiral = doc.addObject("Part::Spiral", ` + PyQuote(objName) + `)
spiral.Growth = ` + PyNum(growth) + `
spiral.Rotations = ` + PyNum(rotations) + `
spiral.Radius = ` + PyNum(radius) + `
doc.recompute()
_mcp_result["result"] = {"name": spiral.Name, "label": spiral.Label, "growth": ` + PyNum(growth) + `, "rotations": ` + PyNum(rotations) + `, "radius": ` + PyNum(radius) + `}
`), nil
}

func hPrimCreateEllipsoid(args map[string]any, b *Bridge) (Result, error) {
	objName := EscapePythonString(ArgStr(args, "name", "Ellipsoid"))
	r1, err := ValidatePositiveNumber(args["radius1"], "radius1")
	if err != nil {
		return Result{}, err
	}
	r2, err := ValidatePositiveNumber(args["radius2"], "radius2")
	if err != nil {
		return Result{}, err
	}
	r3Raw := any(r2)
	if v, ok := args["radius3"]; ok && v != nil {
		r3Raw = v
	}
	r3, err := ValidatePositiveNumber(r3Raw, "radius3")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
ell = doc.addObject("Part::Ellipsoid", ` + PyQuote(objName) + `)
ell.Radius1 = ` + PyNum(r1) + `
ell.Radius2 = ` + PyNum(r2) + `
ell.Radius3 = ` + PyNum(r3) + `
doc.recompute()
_mcp_result["result"] = {"name": ell.Name, "label": ell.Label, "radius1": ` + PyNum(r1) + `, "radius2": ` + PyNum(r2) + `, "radius3": ` + PyNum(r3) + `}
`), nil
}

func hPrimCreateTube(args map[string]any, b *Bridge) (Result, error) {
	objName := EscapePythonString(ArgStr(args, "name", "Tube"))
	outerRadius, err := ValidatePositiveNumber(args["outerRadius"], "outerRadius")
	if err != nil {
		return Result{}, err
	}
	innerRadius, err := ValidatePositiveNumber(args["innerRadius"], "innerRadius")
	if err != nil {
		return Result{}, err
	}
	height, err := ValidatePositiveNumber(args["height"], "height")
	if err != nil {
		return Result{}, err
	}
	if innerRadius >= outerRadius {
		return Result{}, errors.New("innerRadius must be less than outerRadius")
	}
	return b.Run(`
` + DocPreamble + `
outer = Part.makeCylinder(` + PyNum(outerRadius) + `, ` + PyNum(height) + `)
inner = Part.makeCylinder(` + PyNum(innerRadius) + `, ` + PyNum(height) + `)
tube_shape = outer.cut(inner)
tube = doc.addObject("Part::Feature", ` + PyQuote(objName) + `)
tube.Shape = tube_shape
doc.recompute()
_mcp_result["result"] = {"name": tube.Name, "label": tube.Label, "outerRadius": ` + PyNum(outerRadius) + `, "innerRadius": ` + PyNum(innerRadius) + `, "height": ` + PyNum(height) + `}
`), nil
}

func hPrimCreatePrism(args map[string]any, b *Bridge) (Result, error) {
	objName := EscapePythonString(ArgStr(args, "name", "Prism"))
	sides, err := ValidateNumber(args["sides"], "sides", 3, 360)
	if err != nil {
		return Result{}, err
	}
	radius, err := ValidatePositiveNumber(args["radius"], "radius")
	if err != nil {
		return Result{}, err
	}
	height, err := ValidatePositiveNumber(args["height"], "height")
	if err != nil {
		return Result{}, err
	}
	return b.Run(`
` + DocPreamble + `
prism = doc.addObject("Part::Prism", ` + PyQuote(objName) + `)
prism.Polygon = ` + PyNum(sides) + `
prism.Circumradius = ` + PyNum(radius) + `
prism.Height = ` + PyNum(height) + `
doc.recompute()
_mcp_result["result"] = {"name": prism.Name, "label": prism.Label, "sides": ` + PyNum(sides) + `, "radius": ` + PyNum(radius) + `, "height": ` + PyNum(height) + `}
`), nil
}
