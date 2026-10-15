package tools

import (
	"strconv"
	"strings"
)

func init() {
	register("sketcher",
		ToolDef{Name: "freecad_create_sketch", Handler: hSktCreateSketch},
		ToolDef{Name: "freecad_add_sketch_line", Handler: hSktAddSketchLine},
		ToolDef{Name: "freecad_add_sketch_circle", Handler: hSktAddSketchCircle},
		ToolDef{Name: "freecad_add_sketch_arc", Handler: hSktAddSketchArc},
		ToolDef{Name: "freecad_add_sketch_rectangle", Handler: hSktAddSketchRectangle},
		ToolDef{Name: "freecad_add_sketch_constraint", Handler: hSktAddSketchConstraint},
		ToolDef{Name: "freecad_add_sketch_ellipse", Handler: hSktAddSketchEllipse},
		ToolDef{Name: "freecad_add_sketch_bspline", Handler: hSktAddSketchBspline},
		ToolDef{Name: "freecad_add_sketch_polygon", Handler: hSktAddSketchPolygon},
		ToolDef{Name: "freecad_add_sketch_slot", Handler: hSktAddSketchSlot},
		ToolDef{Name: "freecad_add_sketch_point", Handler: hSktAddSketchPoint},
		ToolDef{Name: "freecad_add_sketch_external", Handler: hSktAddSketchExternal},
		ToolDef{Name: "freecad_sketch_fillet", Handler: hSktSketchFillet},
		ToolDef{Name: "freecad_sketch_trim", Handler: hSktSketchTrim},
		ToolDef{Name: "freecad_sketch_toggle_construction", Handler: hSktSketchToggleConstruction},
		ToolDef{Name: "freecad_close_sketch", Handler: hSktCloseSketch},
	)
}

func hSktCreateSketch(args map[string]any, b *Bridge) (Result, error) {
	sketchName := ArgStr(args, "name", "Sketch")
	plane := ArgStr(args, "plane", "XY")

	placementMap := map[string]string{
		"XY": "FreeCAD.Placement(FreeCAD.Vector(0,0,0), FreeCAD.Rotation(FreeCAD.Vector(0,0,1), 0))",
		"XZ": "FreeCAD.Placement(FreeCAD.Vector(0,0,0), FreeCAD.Rotation(FreeCAD.Vector(1,0,0), -90))",
		"YZ": "FreeCAD.Placement(FreeCAD.Vector(0,0,0), FreeCAD.Rotation(FreeCAD.Vector(0,1,0), 90))",
	}
	placement, ok := placementMap[plane]
	if !ok {
		placement = placementMap["XY"]
	}

	return b.Run(`
import Sketcher
doc = FreeCAD.ActiveDocument
sketch = doc.addObject("Sketcher::SketchObject", ` + PyQuote(sketchName) + `)
sketch.MapMode = "Deactivated"
sketch.Placement = ` + placement + `
doc.recompute()
_mcp_result["result"] = {"name": sketch.Name, "plane": ` + PyQuote(plane) + `}
`), nil
}

func hSktAddSketchLine(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	x1, _ := args["x1"].(float64)
	y1, _ := args["y1"].(float64)
	x2, _ := args["x2"].(float64)
	y2, _ := args["y2"].(float64)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
idx = sketch.addGeometry(Part.LineSegment(FreeCAD.Vector(` + PyNum(x1) + `, ` + PyNum(y1) + `, 0), FreeCAD.Vector(` + PyNum(x2) + `, ` + PyNum(y2) + `, 0)))
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndex": idx}
`), nil
}

func hSktAddSketchCircle(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	cx, _ := args["centerX"].(float64)
	cy, _ := args["centerY"].(float64)
	radius, _ := args["radius"].(float64)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
idx = sketch.addGeometry(Part.Circle(FreeCAD.Vector(` + PyNum(cx) + `, ` + PyNum(cy) + `, 0), FreeCAD.Vector(0, 0, 1), ` + PyNum(radius) + `))
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndex": idx}
`), nil
}

func hSktAddSketchArc(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	cx, _ := args["centerX"].(float64)
	cy, _ := args["centerY"].(float64)
	radius, _ := args["radius"].(float64)
	startAngle, _ := args["startAngle"].(float64)
	endAngle, _ := args["endAngle"].(float64)

	return b.Run(`
import math
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
circle = Part.Circle(FreeCAD.Vector(` + PyNum(cx) + `, ` + PyNum(cy) + `, 0), FreeCAD.Vector(0, 0, 1), ` + PyNum(radius) + `)
arc = Part.ArcOfCircle(circle, math.radians(` + PyNum(startAngle) + `), math.radians(` + PyNum(endAngle) + `))
idx = sketch.addGeometry(arc)
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndex": idx}
`), nil
}

func hSktAddSketchRectangle(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	x1, _ := args["x1"].(float64)
	y1, _ := args["y1"].(float64)
	x2, _ := args["x2"].(float64)
	y2, _ := args["y2"].(float64)

	return b.Run(`
import Sketcher
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
# Add 4 line segments: bottom, right, top, left
i0 = sketch.addGeometry(Part.LineSegment(FreeCAD.Vector(` + PyNum(x1) + `, ` + PyNum(y1) + `, 0), FreeCAD.Vector(` + PyNum(x2) + `, ` + PyNum(y1) + `, 0)))
i1 = sketch.addGeometry(Part.LineSegment(FreeCAD.Vector(` + PyNum(x2) + `, ` + PyNum(y1) + `, 0), FreeCAD.Vector(` + PyNum(x2) + `, ` + PyNum(y2) + `, 0)))
i2 = sketch.addGeometry(Part.LineSegment(FreeCAD.Vector(` + PyNum(x2) + `, ` + PyNum(y2) + `, 0), FreeCAD.Vector(` + PyNum(x1) + `, ` + PyNum(y2) + `, 0)))
i3 = sketch.addGeometry(Part.LineSegment(FreeCAD.Vector(` + PyNum(x1) + `, ` + PyNum(y2) + `, 0), FreeCAD.Vector(` + PyNum(x1) + `, ` + PyNum(y1) + `, 0)))
# Add coincidence constraints to close the rectangle
# bottom-end to right-start
sketch.addConstraint(Sketcher.Constraint("Coincident", i0, 2, i1, 1))
# right-end to top-start
sketch.addConstraint(Sketcher.Constraint("Coincident", i1, 2, i2, 1))
# top-end to left-start
sketch.addConstraint(Sketcher.Constraint("Coincident", i2, 2, i3, 1))
# left-end to bottom-start
sketch.addConstraint(Sketcher.Constraint("Coincident", i3, 2, i0, 1))
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndices": [i0, i1, i2, i3]}
`), nil
}

func hSktAddSketchConstraint(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	constraintType, _ := args["constraintType"].(string)
	index1, _ := args["index1"].(float64)
	index2, hasIndex2 := args["index2"].(float64)
	value, hasValue := args["value"].(float64)

	// Map user-friendly names to FreeCAD constraint names
	constraintNameMap := map[string]string{
		"coincident":    "Coincident",
		"horizontal":    "Horizontal",
		"vertical":      "Vertical",
		"fixed":         "Block",
		"distance":      "Distance",
		"distanceX":     "DistanceX",
		"distanceY":     "DistanceY",
		"radius":        "Radius",
		"diameter":      "Diameter",
		"angle":         "Angle",
		"equal":         "Equal",
		"tangent":       "Tangent",
		"perpendicular": "Perpendicular",
		"parallel":      "Parallel",
		"pointOnObject": "PointOnObject",
		"symmetric":     "Symmetric",
	}
	fcConstraint, ok := constraintNameMap[constraintType]
	if !ok {
		fcConstraint = constraintType
	}

	// Build constraint arguments based on what's provided
	var constraintArgs string
	if hasValue && hasIndex2 {
		constraintArgs = PyQuote(fcConstraint) + ", " + PyNum(index1) + ", " + PyNum(index2) + ", " + PyNum(value)
	} else if hasIndex2 {
		constraintArgs = PyQuote(fcConstraint) + ", " + PyNum(index1) + ", " + PyNum(index2)
	} else if hasValue {
		constraintArgs = PyQuote(fcConstraint) + ", " + PyNum(index1) + ", " + PyNum(value)
	} else {
		constraintArgs = PyQuote(fcConstraint) + ", " + PyNum(index1)
	}

	return b.Run(`
import Sketcher
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
cid = sketch.addConstraint(Sketcher.Constraint(` + constraintArgs + `))
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "constraintIndex": cid, "type": ` + PyQuote(constraintType) + `}
`), nil
}

func hSktAddSketchEllipse(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	cx, _ := args["centerX"].(float64)
	cy, _ := args["centerY"].(float64)
	majorRadius, _ := args["majorRadius"].(float64)
	minorRadius, _ := args["minorRadius"].(float64)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
import Part
ellipse = Part.Ellipse(FreeCAD.Vector(` + PyNum(cx) + `, ` + PyNum(cy) + `, 0), ` + PyNum(majorRadius) + `, ` + PyNum(minorRadius) + `)
idx = sketch.addGeometry(ellipse)
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndex": idx, "majorRadius": ` + PyNum(majorRadius) + `, "minorRadius": ` + PyNum(minorRadius) + `}
`), nil
}

func hSktAddSketchBspline(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	points, _ := args["points"].([]any)
	closed := ArgBool(args, "closed", false)

	parts := make([]string, 0, len(points))
	for _, p := range points {
		pm, _ := p.(map[string]any)
		x, _ := pm["x"].(float64)
		y, _ := pm["y"].(float64)
		parts = append(parts, "FreeCAD.Vector("+PyNum(x)+", "+PyNum(y)+", 0)")
	}
	pointsList := strings.Join(parts, ", ")

	periodic := "False"
	if closed {
		periodic = "True"
	}

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
import Part
pts = [` + pointsList + `]
bspline = Part.BSplineCurve()
bspline.interpolate(pts, PeriodicFlag=` + periodic + `)
idx = sketch.addGeometry(bspline)
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndex": idx, "points": ` + strconv.Itoa(len(points)) + `, "closed": ` + strconv.FormatBool(closed) + `}
`), nil
}

func hSktAddSketchPolygon(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	cx, _ := args["centerX"].(float64)
	cy, _ := args["centerY"].(float64)
	radius, _ := args["radius"].(float64)
	sides, _ := args["sides"].(float64)

	return b.Run(`
import math, Sketcher
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
n = ` + PyNum(sides) + `
r = ` + PyNum(radius) + `
cx, cy = ` + PyNum(cx) + `, ` + PyNum(cy) + `
indices = []
for i in range(n):
    angle1 = 2 * math.pi * i / n
    angle2 = 2 * math.pi * ((i + 1) % n) / n
    x1 = cx + r * math.cos(angle1)
    y1 = cy + r * math.sin(angle1)
    x2 = cx + r * math.cos(angle2)
    y2 = cy + r * math.sin(angle2)
    idx = sketch.addGeometry(Part.LineSegment(FreeCAD.Vector(x1, y1, 0), FreeCAD.Vector(x2, y2, 0)))
    indices.append(idx)
# Close with coincidence constraints
for i in range(n):
    j = (i + 1) % n
    sketch.addConstraint(Sketcher.Constraint("Coincident", indices[i], 2, indices[j], 1))
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndices": indices, "sides": n}
`), nil
}

func hSktAddSketchSlot(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	x1, _ := args["x1"].(float64)
	y1, _ := args["y1"].(float64)
	x2, _ := args["x2"].(float64)
	y2, _ := args["y2"].(float64)
	width, _ := args["width"].(float64)

	return b.Run(`
import math, Sketcher
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
# Compute slot geometry
dx = ` + PyNum(x2) + ` - ` + PyNum(x1) + `
dy = ` + PyNum(y2) + ` - ` + PyNum(y1) + `
length = math.sqrt(dx*dx + dy*dy)
if length < 1e-6:
    raise ValueError("Slot centers must be different points")
# Normal direction for offset
nx = -dy / length * ` + PyNum(width) + ` / 2
ny = dx / length * ` + PyNum(width) + ` / 2
angle_base = math.atan2(dy, dx)
# Create two lines and two arcs
# Top line
l1 = sketch.addGeometry(Part.LineSegment(
    FreeCAD.Vector(` + PyNum(x1) + `+nx, ` + PyNum(y1) + `+ny, 0),
    FreeCAD.Vector(` + PyNum(x2) + `+nx, ` + PyNum(y2) + `+ny, 0)))
# Bottom line
l2 = sketch.addGeometry(Part.LineSegment(
    FreeCAD.Vector(` + PyNum(x2) + `-nx, ` + PyNum(y2) + `-ny, 0),
    FreeCAD.Vector(` + PyNum(x1) + `-nx, ` + PyNum(y1) + `-ny, 0)))
# Right semicircle
a1 = sketch.addGeometry(Part.ArcOfCircle(
    Part.Circle(FreeCAD.Vector(` + PyNum(x2) + `, ` + PyNum(y2) + `, 0), FreeCAD.Vector(0,0,1), ` + PyNum(width) + `/2),
    angle_base - math.pi/2, angle_base + math.pi/2))
# Left semicircle
a2 = sketch.addGeometry(Part.ArcOfCircle(
    Part.Circle(FreeCAD.Vector(` + PyNum(x1) + `, ` + PyNum(y1) + `, 0), FreeCAD.Vector(0,0,1), ` + PyNum(width) + `/2),
    angle_base + math.pi/2, angle_base + 3*math.pi/2))
# Connect with coincidence
sketch.addConstraint(Sketcher.Constraint("Coincident", l1, 2, a1, 1))
sketch.addConstraint(Sketcher.Constraint("Coincident", a1, 2, l2, 1))
sketch.addConstraint(Sketcher.Constraint("Coincident", l2, 2, a2, 1))
sketch.addConstraint(Sketcher.Constraint("Coincident", a2, 2, l1, 1))
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndices": [l1, l2, a1, a2], "width": ` + PyNum(width) + `}
`), nil
}

func hSktAddSketchPoint(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	x, _ := args["x"].(float64)
	y, _ := args["y"].(float64)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
idx = sketch.addGeometry(Part.Point(FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, 0)))
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryIndex": idx}
`), nil
}

func hSktAddSketchExternal(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	objectName, _ := args["objectName"].(string)
	edgeName, _ := args["edgeName"].(string)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
sketch.addExternal(` + PyQuote(objectName) + `, ` + PyQuote(edgeName) + `)
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "external": ` + PyQuote(objectName) + `, "edge": ` + PyQuote(edgeName) + `}
`), nil
}

func hSktSketchFillet(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	vertexIndex, _ := args["vertexIndex"].(float64)
	radius, _ := args["radius"].(float64)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
idx = sketch.fillet(` + PyNum(vertexIndex) + `, ` + PyNum(radius) + `)
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "filletIndex": idx, "radius": ` + PyNum(radius) + `}
`), nil
}

func hSktSketchTrim(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	geoIndex, _ := args["geoIndex"].(float64)
	posX, _ := args["posX"].(float64)
	posY, _ := args["posY"].(float64)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
sketch.trim(` + PyNum(geoIndex) + `, FreeCAD.Vector(` + PyNum(posX) + `, ` + PyNum(posY) + `, 0))
doc.recompute()
_mcp_result["result"] = {"sketchName": sketch.Name, "trimmedGeoIndex": ` + PyNum(geoIndex) + `}
`), nil
}

func hSktSketchToggleConstruction(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)
	geoIndex, _ := args["geoIndex"].(float64)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
sketch.toggleConstruction(` + PyNum(geoIndex) + `)
doc.recompute()
is_construction = sketch.getConstruction(` + PyNum(geoIndex) + `)
_mcp_result["result"] = {"sketchName": sketch.Name, "geoIndex": ` + PyNum(geoIndex) + `, "isConstruction": is_construction}
`), nil
}

func hSktCloseSketch(args map[string]any, b *Bridge) (Result, error) {
	sketchName, _ := args["sketchName"].(string)

	return b.Run(`
doc = FreeCAD.ActiveDocument
sketch = doc.getObject(` + PyQuote(sketchName) + `)
doc.recompute()
geo_count = sketch.GeometryCount
constraint_count = sketch.ConstraintCount
_mcp_result["result"] = {"sketchName": sketch.Name, "geometryCount": geo_count, "constraintCount": constraint_count, "status": "closed"}
`), nil
}
