package tools

import (
	"strings"
)

func init() {
	register("bim",
		ToolDef{Name: "freecad_arch_wall", Handler: hBimArchWall},
		ToolDef{Name: "freecad_arch_structure", Handler: hBimArchStructure},
		ToolDef{Name: "freecad_arch_window", Handler: hBimArchWindow},
		ToolDef{Name: "freecad_arch_floor", Handler: hBimArchFloor},
		ToolDef{Name: "freecad_arch_building", Handler: hBimArchBuilding},
		ToolDef{Name: "freecad_arch_site", Handler: hBimArchSite},
		ToolDef{Name: "freecad_arch_roof", Handler: hBimArchRoof},
		ToolDef{Name: "freecad_arch_stairs", Handler: hBimArchStairs},
		ToolDef{Name: "freecad_export_ifc", Handler: hBimExportIfc},
	)
}

func hBimArchWall(args map[string]any, b *Bridge) (Result, error) {
	baselineName, _ := args["baselineName"].(string)
	width := ArgNum(args, "width", 200)
	height := ArgNum(args, "height", 3000)
	wallName := ArgStr(args, "name", "Wall")

	if baselineName != "" {
		return b.Run(`
` + DocPreamble + `
import Arch
baseline = doc.getObject(` + PyQuote(baselineName) + `)
if baseline is None:
    raise ValueError("Baseline not found: ` + baselineName + `")
wall = Arch.makeWall(baseline, width=` + PyNum(width) + `, height=` + PyNum(height) + `)
wall.Label = ` + PyQuote(wallName) + `
doc.recompute()
_mcp_result["result"] = {"name": wall.Name, "label": wall.Label, "width": ` + PyNum(width) + `, "height": ` + PyNum(height) + `}
`), nil
	}

	startX := ArgNum(args, "startX", 0)
	startY := ArgNum(args, "startY", 0)
	endX := ArgNum(args, "endX", 1000)
	endY := ArgNum(args, "endY", 0)
	length, hasLength := args["length"].(float64)

	if hasLength {
		return b.Run(`
` + DocPreamble + `
import Arch
wall = Arch.makeWall(None, length=` + PyNum(length) + `, width=` + PyNum(width) + `, height=` + PyNum(height) + `)
wall.Label = ` + PyQuote(wallName) + `
doc.recompute()
_mcp_result["result"] = {"name": wall.Name, "label": wall.Label, "length": ` + PyNum(length) + `, "width": ` + PyNum(width) + `, "height": ` + PyNum(height) + `}
`), nil
	}

	return b.Run(`
` + DocPreamble + `
import Arch, Draft
line = Draft.makeLine(FreeCAD.Vector(` + PyNum(startX) + `, ` + PyNum(startY) + `, 0), FreeCAD.Vector(` + PyNum(endX) + `, ` + PyNum(endY) + `, 0))
doc.recompute()
wall = Arch.makeWall(line, width=` + PyNum(width) + `, height=` + PyNum(height) + `)
wall.Label = ` + PyQuote(wallName) + `
doc.recompute()
import math
l = math.sqrt((` + PyNum(endX) + `-` + PyNum(startX) + `)**2 + (` + PyNum(endY) + `-` + PyNum(startY) + `)**2)
_mcp_result["result"] = {"name": wall.Name, "label": wall.Label, "length": l, "width": ` + PyNum(width) + `, "height": ` + PyNum(height) + `}
`), nil
}

func hBimArchStructure(args map[string]any, b *Bridge) (Result, error) {
	structureType := ArgStr(args, "structureType", "column")
	length := ArgNum(args, "length", 500)
	width := ArgNum(args, "width", 500)
	defaultHeight := 500.0
	if structureType == "column" {
		defaultHeight = 3000
	}
	height := ArgNum(args, "height", defaultHeight)
	x := ArgNum(args, "x", 0)
	y := ArgNum(args, "y", 0)
	z := ArgNum(args, "z", 0)
	strName := ArgStr(args, "name", "")
	if strName == "" {
		strName = strings.ToUpper(structureType[:1]) + structureType[1:]
	}
	return b.Run(`
` + DocPreamble + `
import Arch
structure = Arch.makeStructure(length=` + PyNum(length) + `, width=` + PyNum(width) + `, height=` + PyNum(height) + `)
structure.Label = ` + PyQuote(strName) + `
structure.Placement.Base = FreeCAD.Vector(` + PyNum(x) + `, ` + PyNum(y) + `, ` + PyNum(z) + `)
doc.recompute()
_mcp_result["result"] = {"name": structure.Name, "label": structure.Label, "type": ` + PyQuote(structureType) + `, "dimensions": {"length": ` + PyNum(length) + `, "width": ` + PyNum(width) + `, "height": ` + PyNum(height) + `}}
`), nil
}

func hBimArchWindow(args map[string]any, b *Bridge) (Result, error) {
	wallName, _ := args["wallName"].(string)
	windowWidth := ArgNum(args, "windowWidth", 1000)
	windowHeight := ArgNum(args, "windowHeight", 1200)
	sillHeight := ArgNum(args, "sillHeight", 900)
	winName := ArgStr(args, "name", "Window")
	return b.Run(`
` + DocPreamble + `
import Arch
wall = doc.getObject(` + PyQuote(wallName) + `)
if wall is None:
    raise ValueError("Wall not found: ` + wallName + `")
# Create a simple rectangular window
import Draft
rect = Draft.makeRectangle(` + PyNum(windowWidth) + `, ` + PyNum(windowHeight) + `)
rect.Placement.Base = FreeCAD.Vector(0, 0, ` + PyNum(sillHeight) + `)
doc.recompute()
window = Arch.makeWindow(rect)
window.Label = ` + PyQuote(winName) + `
window.Hosts = [wall]
doc.recompute()
_mcp_result["result"] = {"name": window.Name, "label": window.Label, "width": ` + PyNum(windowWidth) + `, "height": ` + PyNum(windowHeight) + `, "sillHeight": ` + PyNum(sillHeight) + `}
`), nil
}

func hBimArchFloor(args map[string]any, b *Bridge) (Result, error) {
	height := ArgNum(args, "height", 3000)
	objectNames, _ := args["objectNames"].([]any)
	floorName := ArgStr(args, "name", "Level")
	objCode := ""
	if objectNames != nil {
		objCode = `floor.Group = [doc.getObject(n) for n in ` + PyQuoteAny(objectNames) + ` if doc.getObject(n)]`
	}
	return b.Run(`
` + DocPreamble + `
import Arch
floor = Arch.makeFloor()
floor.Label = ` + PyQuote(floorName) + `
floor.Height = ` + PyNum(height) + `
` + objCode + `
doc.recompute()
_mcp_result["result"] = {"name": floor.Name, "label": floor.Label, "height": ` + PyNum(height) + `}
`), nil
}

func hBimArchBuilding(args map[string]any, b *Bridge) (Result, error) {
	floorNames, _ := args["floorNames"].([]any)
	buildingName := ArgStr(args, "name", "Building")
	floorsCode := ""
	if floorNames != nil {
		floorsCode = `building.Group = [doc.getObject(n) for n in ` + PyQuoteAny(floorNames) + ` if doc.getObject(n)]`
	}
	return b.Run(`
` + DocPreamble + `
import Arch
building = Arch.makeBuilding()
building.Label = ` + PyQuote(buildingName) + `
` + floorsCode + `
doc.recompute()
_mcp_result["result"] = {"name": building.Name, "label": building.Label}
`), nil
}

func hBimArchSite(args map[string]any, b *Bridge) (Result, error) {
	buildingNames, _ := args["buildingNames"].([]any)
	siteName := ArgStr(args, "name", "Site")
	bldgsCode := ""
	if buildingNames != nil {
		bldgsCode = `site.Group = [doc.getObject(n) for n in ` + PyQuoteAny(buildingNames) + ` if doc.getObject(n)]`
	}
	return b.Run(`
` + DocPreamble + `
import Arch
site = Arch.makeSite()
site.Label = ` + PyQuote(siteName) + `
` + bldgsCode + `
doc.recompute()
_mcp_result["result"] = {"name": site.Name, "label": site.Label}
`), nil
}

func hBimArchRoof(args map[string]any, b *Bridge) (Result, error) {
	baseName, _ := args["baseName"].(string)
	angle := ArgNum(args, "angle", 45)
	roofName := ArgStr(args, "name", "Roof")
	return b.Run(`
` + DocPreamble + `
import Arch
base = doc.getObject(` + PyQuote(baseName) + `)
if base is None:
    raise ValueError("Base not found: ` + baseName + `")
roof = Arch.makeRoof(base, angles=[` + PyNum(angle) + `])
roof.Label = ` + PyQuote(roofName) + `
doc.recompute()
_mcp_result["result"] = {"name": roof.Name, "label": roof.Label, "angle": ` + PyNum(angle) + `}
`), nil
}

func hBimArchStairs(args map[string]any, b *Bridge) (Result, error) {
	numberOfSteps := ArgNum(args, "numberOfSteps", 10)
	totalHeight := ArgNum(args, "totalHeight", 3000)
	width := ArgNum(args, "width", 1000)
	stairsName := ArgStr(args, "name", "Stairs")
	return b.Run(`
` + DocPreamble + `
import Arch
stairs = Arch.makeStairs(` + PyNum(numberOfSteps) + `)
stairs.Label = ` + PyQuote(stairsName) + `
stairs.Height = ` + PyNum(totalHeight) + `
stairs.Width = ` + PyNum(width) + `
doc.recompute()
_mcp_result["result"] = {"name": stairs.Name, "label": stairs.Label, "steps": ` + PyNum(numberOfSteps) + `, "height": ` + PyNum(totalHeight) + `, "width": ` + PyNum(width) + `}
`), nil
}

func hBimExportIfc(args map[string]any, b *Bridge) (Result, error) {
	filePath, _ := args["filePath"].(string)
	objectNames, _ := args["objectNames"].([]any)
	objsCode := `objs = doc.Objects`
	if objectNames != nil {
		objsCode = `objs = [doc.getObject(n) for n in ` + PyQuoteAny(objectNames) + ` if doc.getObject(n)]`
	}
	return b.Run(`
` + DocPreamble + `
import importIFC
` + objsCode + `
importIFC.export(objs, ` + PyQuote(filePath) + `)
import os
_mcp_result["result"] = {"filePath": ` + PyQuote(filePath) + `, "objectCount": len(objs), "size_bytes": os.path.getsize(` + PyQuote(filePath) + `)}
`), nil
}
