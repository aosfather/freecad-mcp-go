package tools

func init() {
	register("part-design",
		ToolDef{Name: "freecad_pad", Handler: hPdPad},
		ToolDef{Name: "freecad_pocket", Handler: hPdPocket},
		ToolDef{Name: "freecad_revolve", Handler: hPdRevolve},
		ToolDef{Name: "freecad_loft", Handler: hPdLoft},
		ToolDef{Name: "freecad_sweep", Handler: hPdSweep},
		ToolDef{Name: "freecad_partdesign_fillet", Handler: hPdPartdesignFillet},
		ToolDef{Name: "freecad_partdesign_chamfer", Handler: hPdPartdesignChamfer},
		ToolDef{Name: "freecad_hole", Handler: hPdHole},
		ToolDef{Name: "freecad_partdesign_thickness", Handler: hPdPartdesignThickness},
		ToolDef{Name: "freecad_linear_pattern", Handler: hPdLinearPattern},
		ToolDef{Name: "freecad_polar_pattern", Handler: hPdPolarPattern},
		ToolDef{Name: "freecad_partdesign_mirrored", Handler: hPdPartdesignMirrored},
		ToolDef{Name: "freecad_additive_helix", Handler: hPdAdditiveHelix},
		ToolDef{Name: "freecad_subtractive_helix", Handler: hPdSubtractiveHelix},
		ToolDef{Name: "freecad_additive_pipe", Handler: hPdAdditivePipe},
		ToolDef{Name: "freecad_subtractive_pipe", Handler: hPdSubtractivePipe},
		ToolDef{Name: "freecad_groove", Handler: hPdGroove},
		ToolDef{Name: "freecad_draft_angle", Handler: hPdDraftAngle},
		ToolDef{Name: "freecad_multi_transform", Handler: hPdMultiTransform},
		ToolDef{Name: "freecad_shape_binder", Handler: hPdShapeBinder},
	)
}

func hPdPad(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	length, _ := args["length"].(float64)
	symmetric := ArgBool(args, "symmetric", false)
	reversed := ArgBool(args, "reversed", false)
	padName := ArgStr(args, "name", "Pad")
	symStr := "False"
	if symmetric {
		symStr = "True"
	}
	revStr := "False"
	if reversed {
		revStr = "True"
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
# Create a PartDesign Body if none exists
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    body = doc.addObject("PartDesign::Body", "Body")
if sketch not in body.Group:
    body.addObject(sketch)
pad = doc.addObject("PartDesign::Pad", ` + PyQuote(padName) + `)
pad.Profile = sketch
pad.Length = ` + PyNum(length) + `
pad.Symmetric = ` + symStr + `
pad.Reversed = ` + revStr + `
body.addObject(pad)
doc.recompute()
_mcp_result["result"] = {"name": pad.Name, "length": pad.Length, "type": pad.TypeId}
`), nil
}

func hPdPocket(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	depth, _ := args["depth"].(float64)
	pocketName := ArgStr(args, "name", "Pocket")
	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    body = doc.addObject("PartDesign::Body", "Body")
if sketch not in body.Group:
    body.addObject(sketch)
pocket = doc.addObject("PartDesign::Pocket", ` + PyQuote(pocketName) + `)
pocket.Profile = sketch
pocket.Length = ` + PyNum(depth) + `
body.addObject(pocket)
doc.recompute()
_mcp_result["result"] = {"name": pocket.Name, "depth": pocket.Length, "type": pocket.TypeId}
`), nil
}

func hPdRevolve(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	angle := ArgNum(args, "angle", 360)
	axisX := ArgNum(args, "axisX", 0)
	axisY := ArgNum(args, "axisY", 1)
	axisZ := ArgNum(args, "axisZ", 0)
	revName := ArgStr(args, "name", "Revolution")
	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    body = doc.addObject("PartDesign::Body", "Body")
if sketch not in body.Group:
    body.addObject(sketch)
rev = doc.addObject("PartDesign::Revolution", ` + PyQuote(revName) + `)
rev.Profile = sketch
rev.Angle = ` + PyNum(angle) + `
rev.Axis = FreeCAD.Vector(` + PyNum(axisX) + `, ` + PyNum(axisY) + `, ` + PyNum(axisZ) + `)
body.addObject(rev)
doc.recompute()
_mcp_result["result"] = {"name": rev.Name, "angle": rev.Angle, "type": rev.TypeId}
`), nil
}

func hPdLoft(args map[string]any, b *Bridge) (Result, error) {
	sketchNames, _ := args["sketchNames"].([]any)
	solid := ArgBool(args, "solid", true)
	loftName := ArgStr(args, "name", "Loft")
	sectionsCode := `[doc.getObject(s) for s in ` + PyQuoteAny(sketchNames) + `]`
	pySolid := "False"
	if solid {
		pySolid = "True"
	}
	jsSolid := "false"
	if solid {
		jsSolid = "true"
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
loft = doc.addObject("Part::Loft", ` + PyQuote(loftName) + `)
loft.Sections = ` + sectionsCode + `
loft.Solid = ` + pySolid + `
doc.recompute()
_mcp_result["result"] = {"name": loft.Name, "sections": ` + PyQuoteAny(sketchNames) + `, "solid": ` + jsSolid + `, "type": loft.TypeId}
`), nil
}

func hPdSweep(args map[string]any, b *Bridge) (Result, error) {
	profileName, _ := args["profileName"].(string)
	pathName, _ := args["pathName"].(string)
	solid := ArgBool(args, "solid", true)
	sweepName := ArgStr(args, "name", "Sweep")
	pySolid := "False"
	if solid {
		pySolid = "True"
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
sweep = doc.addObject("Part::Sweep", ` + PyQuote(sweepName) + `)
sweep.Sections = [doc.getObject(` + PyQuote(profileName) + `)]
sweep.Spine = (doc.getObject(` + PyQuote(pathName) + `), [])
sweep.Solid = ` + pySolid + `
doc.recompute()
_mcp_result["result"] = {"name": sweep.Name, "profile": ` + PyQuote(profileName) + `, "path": ` + PyQuote(pathName) + `, "type": sweep.TypeId}
`), nil
}

func hPdPartdesignFillet(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	radius, _ := args["radius"].(float64)
	edgeNames, hasEdgeNames := args["edgeNames"].([]any)
	filletName := ArgStr(args, "name", "Fillet")
	edgesCode := `["Edge" + str(i+1) for i in range(len(obj.Shape.Edges))]`
	if hasEdgeNames {
		edgesCode = PyQuoteAny(edgeNames)
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
edges = ` + edgesCode + `
fillet = doc.addObject("PartDesign::Fillet", ` + PyQuote(filletName) + `)
fillet.Base = (obj, edges)
fillet.Radius = ` + PyNum(radius) + `
doc.recompute()
_mcp_result["result"] = {"name": fillet.Name, "radius": fillet.Radius, "edges": edges, "type": fillet.TypeId}
`), nil
}

func hPdPartdesignChamfer(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	size, _ := args["size"].(float64)
	edgeNames, hasEdgeNames := args["edgeNames"].([]any)
	chamferName := ArgStr(args, "name", "Chamfer")
	edgesCode := `["Edge" + str(i+1) for i in range(len(obj.Shape.Edges))]`
	if hasEdgeNames {
		edgesCode = PyQuoteAny(edgeNames)
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
edges = ` + edgesCode + `
cham = doc.addObject("PartDesign::Chamfer", ` + PyQuote(chamferName) + `)
cham.Base = (obj, edges)
cham.Size = ` + PyNum(size) + `
doc.recompute()
_mcp_result["result"] = {"name": cham.Name, "size": cham.Size, "edges": edges, "type": cham.TypeId}
`), nil
}

func hPdHole(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	diameter, _ := args["diameter"].(float64)
	depth := ArgNum(args, "depth", 0)
	holeType := ArgStr(args, "holeType", "simple")
	counterboreDiameter, hasCounterboreDiameter := args["counterboreDiameter"].(float64)
	countersinkAngle := ArgNum(args, "countersinkAngle", 90)
	threaded := ArgBool(args, "threaded", false)
	holeName := ArgStr(args, "name", "Hole")
	throughAll := depth == 0
	var depthLine string
	if throughAll {
		depthLine = `hole.DepthType = "ThroughAll"`
	} else {
		depthLine = `hole.DepthType = "Dimension"` + "\n" + `hole.Depth = ` + PyNum(depth)
	}
	var counterboreLine string
	if holeType == "counterbore" && hasCounterboreDiameter {
		counterboreLine = `hole.HoleType = "Counterbore"` + "\n" + `hole.HoleCutDiameter = ` + PyNum(counterboreDiameter) + "\n" + `hole.HoleCutDepth = ` + PyNum(ArgNum(args, "counterboreDepth", 3))
	}
	var countersinkLine string
	if holeType == "countersink" {
		countersinkLine = `hole.HoleType = "Countersink"` + "\n" + `hole.HoleCutCountersinkAngle = ` + PyNum(countersinkAngle)
	}
	var threadedLine string
	if threaded {
		threadedLine = `hole.Threaded = True`
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    body = doc.addObject("PartDesign::Body", "Body")
if sketch not in body.Group:
    body.addObject(sketch)
hole = doc.addObject("PartDesign::Hole", ` + PyQuote(holeName) + `)
hole.Profile = sketch
hole.Diameter = ` + PyNum(diameter) + `
` + depthLine + `
` + counterboreLine + `
` + countersinkLine + `
` + threadedLine + `
body.addObject(hole)
doc.recompute()
_mcp_result["result"] = {"name": hole.Name, "diameter": ` + PyNum(diameter) + `, "depth": ` + PyNum(depth) + `, "type": ` + PyQuote(holeType) + `}
`), nil
}

func hPdPartdesignThickness(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	thickness, _ := args["thickness"].(float64)
	faceNames, _ := args["faceNames"].([]any)
	thkName := ArgStr(args, "name", "Thickness")
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
body = None
for b in doc.Objects:
    if b.TypeId == "PartDesign::Body":
        body = b
        break
if body is None:
    raise ValueError("No PartDesign Body found")
thk = doc.addObject("PartDesign::Thickness", ` + PyQuote(thkName) + `)
thk.Base = (obj, ` + PyQuoteAny(faceNames) + `)
thk.Value = ` + PyNum(thickness) + `
thk.Reversed = False
body.addObject(thk)
doc.recompute()
_mcp_result["result"] = {"name": thk.Name, "thickness": ` + PyNum(thickness) + `, "faces": ` + PyQuoteAny(faceNames) + `}
`), nil
}

func hPdLinearPattern(args map[string]any, b *Bridge) (Result, error) {
	featureName, _ := args["featureName"].(string)
	directionAxis := ArgStr(args, "directionAxis", "x")
	length, _ := args["length"].(float64)
	occurrences, _ := args["occurrences"].(float64)
	lpName := ArgStr(args, "name", "LinearPattern")
	return b.Run(`
doc = FreeCAD.ActiveDocument
feature = doc.getObject(` + PyQuote(featureName) + `)
if feature is None:
    raise ValueError("Feature not found: ` + featureName + `")
body = None
for b in doc.Objects:
    if b.TypeId == "PartDesign::Body":
        body = b
        break
if body is None:
    raise ValueError("No PartDesign Body found")
lp = doc.addObject("PartDesign::LinearPattern", ` + PyQuote(lpName) + `)
lp.Originals = [feature]
lp.Direction = (feature, ["N_Axis"])
lp.Length = ` + PyNum(length) + `
lp.Occurrences = ` + PyNum(occurrences) + `
body.addObject(lp)
doc.recompute()
_mcp_result["result"] = {"name": lp.Name, "length": ` + PyNum(length) + `, "occurrences": ` + PyNum(occurrences) + `, "axis": ` + PyQuote(directionAxis) + `}
`), nil
}

func hPdPolarPattern(args map[string]any, b *Bridge) (Result, error) {
	featureName, _ := args["featureName"].(string)
	axis := ArgStr(args, "axis", "z")
	angle := ArgNum(args, "angle", 360)
	occurrences, _ := args["occurrences"].(float64)
	ppName := ArgStr(args, "name", "PolarPattern")
	return b.Run(`
doc = FreeCAD.ActiveDocument
feature = doc.getObject(` + PyQuote(featureName) + `)
if feature is None:
    raise ValueError("Feature not found: ` + featureName + `")
body = None
for b in doc.Objects:
    if b.TypeId == "PartDesign::Body":
        body = b
        break
if body is None:
    raise ValueError("No PartDesign Body found")
pp = doc.addObject("PartDesign::PolarPattern", ` + PyQuote(ppName) + `)
pp.Originals = [feature]
pp.Axis = (feature, ["N_Axis"])
pp.Angle = ` + PyNum(angle) + `
pp.Occurrences = ` + PyNum(occurrences) + `
body.addObject(pp)
doc.recompute()
_mcp_result["result"] = {"name": pp.Name, "angle": ` + PyNum(angle) + `, "occurrences": ` + PyNum(occurrences) + `, "axis": ` + PyQuote(axis) + `}
`), nil
}

func hPdPartdesignMirrored(args map[string]any, b *Bridge) (Result, error) {
	featureName, _ := args["featureName"].(string)
	plane := ArgStr(args, "plane", "YZ")
	mirName := ArgStr(args, "name", "Mirrored")
	return b.Run(`
doc = FreeCAD.ActiveDocument
feature = doc.getObject(` + PyQuote(featureName) + `)
if feature is None:
    raise ValueError("Feature not found: ` + featureName + `")
body = None
for b in doc.Objects:
    if b.TypeId == "PartDesign::Body":
        body = b
        break
if body is None:
    raise ValueError("No PartDesign Body found")
mir = doc.addObject("PartDesign::Mirrored", ` + PyQuote(mirName) + `)
mir.Originals = [feature]
mir.MirrorPlane = (feature, ["N_Axis"])
body.addObject(mir)
doc.recompute()
_mcp_result["result"] = {"name": mir.Name, "plane": ` + PyQuote(plane) + `}
`), nil
}

func hPdAdditiveHelix(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	pitch, _ := args["pitch"].(float64)
	height, hasHeight := args["height"].(float64)
	turns, hasTurns := args["turns"].(float64)
	leftHanded := ArgBool(args, "leftHanded", false)
	reversed := ArgBool(args, "reversed", false)
	helixName := ArgStr(args, "name", "AdditiveHelix")
	lhStr := "False"
	if leftHanded {
		lhStr = "True"
	}
	revStr := "False"
	if reversed {
		revStr = "True"
	}
	var heightLine string
	if hasHeight {
		heightLine = `helix.Height = ` + PyNum(height)
	}
	var turnsLine string
	if hasTurns {
		turnsLine = `helix.Turns = ` + PyNum(turns)
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    body = doc.addObject("PartDesign::Body", "Body")
if sketch not in body.Group:
    body.addObject(sketch)
helix = doc.addObject("PartDesign::AdditiveHelix", ` + PyQuote(helixName) + `)
helix.Profile = sketch
helix.Pitch = ` + PyNum(pitch) + `
` + heightLine + `
` + turnsLine + `
helix.LeftHanded = ` + lhStr + `
helix.Reversed = ` + revStr + `
body.addObject(helix)
doc.recompute()
_mcp_result["result"] = {"name": helix.Name, "pitch": ` + PyNum(pitch) + `, "type": helix.TypeId}
`), nil
}

func hPdSubtractiveHelix(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	pitch, _ := args["pitch"].(float64)
	height, hasHeight := args["height"].(float64)
	turns, hasTurns := args["turns"].(float64)
	leftHanded := ArgBool(args, "leftHanded", false)
	helixName := ArgStr(args, "name", "SubtractiveHelix")
	lhStr := "False"
	if leftHanded {
		lhStr = "True"
	}
	var heightLine string
	if hasHeight {
		heightLine = `helix.Height = ` + PyNum(height)
	}
	var turnsLine string
	if hasTurns {
		turnsLine = `helix.Turns = ` + PyNum(turns)
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    raise ValueError("No PartDesign Body found — need existing solid to cut from")
if sketch not in body.Group:
    body.addObject(sketch)
helix = doc.addObject("PartDesign::SubtractiveHelix", ` + PyQuote(helixName) + `)
helix.Profile = sketch
helix.Pitch = ` + PyNum(pitch) + `
` + heightLine + `
` + turnsLine + `
helix.LeftHanded = ` + lhStr + `
body.addObject(helix)
doc.recompute()
_mcp_result["result"] = {"name": helix.Name, "pitch": ` + PyNum(pitch) + `, "type": helix.TypeId}
`), nil
}

func hPdAdditivePipe(args map[string]any, b *Bridge) (Result, error) {
	profileName, _ := args["profileName"].(string)
	spineName, _ := args["spineName"].(string)
	pipeName := ArgStr(args, "name", "AdditivePipe")
	return b.Run(`
doc = FreeCAD.ActiveDocument
profile = doc.getObject(` + PyQuote(profileName) + `)
spine = doc.getObject(` + PyQuote(spineName) + `)
if profile is None:
    raise ValueError("Profile not found: ` + profileName + `")
if spine is None:
    raise ValueError("Spine not found: ` + spineName + `")
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    body = doc.addObject("PartDesign::Body", "Body")
if profile not in body.Group:
    body.addObject(profile)
pipe = doc.addObject("PartDesign::AdditivePipe", ` + PyQuote(pipeName) + `)
pipe.Profile = profile
pipe.Spine = (spine, ["Edge1"])
body.addObject(pipe)
doc.recompute()
_mcp_result["result"] = {"name": pipe.Name, "type": pipe.TypeId}
`), nil
}

func hPdSubtractivePipe(args map[string]any, b *Bridge) (Result, error) {
	profileName, _ := args["profileName"].(string)
	spineName, _ := args["spineName"].(string)
	pipeName := ArgStr(args, "name", "SubtractivePipe")
	return b.Run(`
doc = FreeCAD.ActiveDocument
profile = doc.getObject(` + PyQuote(profileName) + `)
spine = doc.getObject(` + PyQuote(spineName) + `)
if profile is None:
    raise ValueError("Profile not found: ` + profileName + `")
if spine is None:
    raise ValueError("Spine not found: ` + spineName + `")
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    raise ValueError("No PartDesign Body found")
if profile not in body.Group:
    body.addObject(profile)
pipe = doc.addObject("PartDesign::SubtractivePipe", ` + PyQuote(pipeName) + `)
pipe.Profile = profile
pipe.Spine = (spine, ["Edge1"])
body.addObject(pipe)
doc.recompute()
_mcp_result["result"] = {"name": pipe.Name, "type": pipe.TypeId}
`), nil
}

func hPdGroove(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	angle := ArgNum(args, "angle", 360)
	axisX := ArgNum(args, "axisX", 0)
	axisY := ArgNum(args, "axisY", 1)
	axisZ := ArgNum(args, "axisZ", 0)
	grooveName := ArgStr(args, "name", "Groove")
	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
body = None
for obj in doc.Objects:
    if obj.TypeId == "PartDesign::Body":
        body = obj
        break
if body is None:
    raise ValueError("No PartDesign Body found")
if sketch not in body.Group:
    body.addObject(sketch)
groove = doc.addObject("PartDesign::Groove", ` + PyQuote(grooveName) + `)
groove.Profile = sketch
groove.Angle = ` + PyNum(angle) + `
groove.Axis = FreeCAD.Vector(` + PyNum(axisX) + `, ` + PyNum(axisY) + `, ` + PyNum(axisZ) + `)
body.addObject(groove)
doc.recompute()
_mcp_result["result"] = {"name": groove.Name, "angle": ` + PyNum(angle) + `, "type": groove.TypeId}
`), nil
}

func hPdDraftAngle(args map[string]any, b *Bridge) (Result, error) {
	featureName, _ := args["featureName"].(string)
	faceNames, _ := args["faceNames"].([]any)
	angle := ArgNum(args, "angle", 5)
	pullDirection, _ := args["pullDirection"].(string)
	draftName := ArgStr(args, "name", "Draft")
	var pullLine string
	if pullDirection != "" {
		pullLine = `draft.NeutralPlane = (feature, [` + PyQuote(pullDirection) + `])`
	}
	return b.Run(`
doc = FreeCAD.ActiveDocument
feature = doc.getObject(` + PyQuote(featureName) + `)
if feature is None:
    raise ValueError("Feature not found: ` + featureName + `")
body = None
for b in doc.Objects:
    if b.TypeId == "PartDesign::Body":
        body = b
        break
if body is None:
    raise ValueError("No PartDesign Body found")
draft = doc.addObject("PartDesign::Draft", ` + PyQuote(draftName) + `)
draft.Base = (feature, ` + PyQuoteAny(faceNames) + `)
draft.Angle = ` + PyNum(angle) + `
` + pullLine + `
body.addObject(draft)
doc.recompute()
_mcp_result["result"] = {"name": draft.Name, "angle": ` + PyNum(angle) + `, "faces": ` + PyQuoteAny(faceNames) + `}
`), nil
}

func hPdMultiTransform(args map[string]any, b *Bridge) (Result, error) {
	featureName, _ := args["featureName"].(string)
	transformNames, _ := args["transformNames"].([]any)
	mtName := ArgStr(args, "name", "MultiTransform")
	return b.Run(`
doc = FreeCAD.ActiveDocument
feature = doc.getObject(` + PyQuote(featureName) + `)
if feature is None:
    raise ValueError("Feature not found: ` + featureName + `")
body = None
for b in doc.Objects:
    if b.TypeId == "PartDesign::Body":
        body = b
        break
if body is None:
    raise ValueError("No PartDesign Body found")
mt = doc.addObject("PartDesign::MultiTransform", ` + PyQuote(mtName) + `)
mt.Originals = [feature]
transforms = [doc.getObject(n) for n in ` + PyQuoteAny(transformNames) + `]
mt.Transformations = transforms
body.addObject(mt)
doc.recompute()
_mcp_result["result"] = {"name": mt.Name, "transforms": ` + PyQuoteAny(transformNames) + `}
`), nil
}

func hPdShapeBinder(args map[string]any, b *Bridge) (Result, error) {
	objectName, _ := args["objectName"].(string)
	subElements, ok := args["subElements"].([]any)
	if !ok {
		subElements = []any{}
	}
	sbName := ArgStr(args, "name", "ShapeBinder")
	return b.Run(`
doc = FreeCAD.ActiveDocument
obj = doc.getObject(` + PyQuote(objectName) + `)
if obj is None:
    raise ValueError("Object not found: ` + objectName + `")
body = None
for b in doc.Objects:
    if b.TypeId == "PartDesign::Body":
        body = b
        break
if body is None:
    body = doc.addObject("PartDesign::Body", "Body")
sb = doc.addObject("PartDesign::ShapeBinder", ` + PyQuote(sbName) + `)
subs = ` + PyQuoteAny(subElements) + `
if subs:
    sb.Support = [(obj, tuple(subs))]
else:
    sb.Support = [(obj, "")]
body.addObject(sb)
doc.recompute()
_mcp_result["result"] = {"name": sb.Name, "source": ` + PyQuote(objectName) + `, "subElements": ` + PyQuoteAny(subElements) + `}
`), nil
}
