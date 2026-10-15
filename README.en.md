[中文](README.md) | **English**

# freecad-mcp-go

> This project is a Go port of [https://github.com/sergiudanstan/freecad-mcp.git](https://github.com/sergiudanstan/freecad-mcp.git).

A FreeCAD MCP (Model Context Protocol) server written in Go. It is a port of the TypeScript-based [sergiudanstan/freecad-mcp](https://github.com/sergiudanstan/freecad-mcp), providing the exact same **165 tools** as the original (tool descriptions and JSON Schemas are generated directly from the TS build output, with a parity test guaranteeing alignment). It lets AI assistants (MCP clients such as Claude, Trae, etc.) drive FreeCAD directly for parametric modeling, assembly, FEM analysis, and more.

## Features

- **All 165 tools**: covering FreeCAD's major workbenches — document management, sketching, Part Design, boolean/transform operations, mesh, surface, Draft, BIM/Arch, FEM, TechDraw, assembly, spreadsheets, and more.
- **Dual-mode connection to FreeCAD**:
  - **GUI Socket mode** (recommended): connects to a running FreeCAD GUI (via the bundled `freecad_server.FCMacro` macro, listening on `127.0.0.1:12345`) so you can watch modeling happen live.
  - **Headless mode** (automatic fallback): when the GUI is unavailable, it automatically launches FreeCAD's built-in `freecadcmd` Python REPL — great for servers/CI.
  - If the socket drops mid-session, it switches to headless mode automatically.
- **stdio MCP protocol**: built on [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go), works with any MCP client that supports stdio.
- **Built-in argument validation**: validation logic consistent with the TS version (see `tools/validation.go`).
- **Tool-count parity test**: `tools/parity_test.go` ensures every registered handler has a matching schema in `schemas.json`, with the exact same tool count as the TS source (165).

## How It Works

```
┌──────────────┐  MCP (stdio)  ┌──────────────────┐   TCP 12345   ┌─────────────────┐
│  MCP Client  │ ◄───────────► │  freecad-mcp-go  │ ◄───────────► │ FreeCAD GUI      │
│ (Claude etc.)│               │                  │  (FCMacro)    │ (MCP macro)      │
└──────────────┘               └────────┬─────────┘               └─────────────────┘
                                        │ falls back if socket unavailable
                                        ▼
                               ┌──────────────────┐
                               │ freecadcmd       │
                               │ headless Python  │
                               │ REPL             │
                               └──────────────────┘
```

- Each MCP tool call generates FreeCAD Python code, which is executed through the Bridge.
- Results are returned as JSON with `success/result/error/output` fields, identical to the TS protocol.
- Global timeout: 30 seconds.

## Requirements

- **Go** 1.26+ (build only)
- **FreeCAD** (1.0+ recommended)
  - GUI mode: run the bundled [freecad_server.FCMacro](freecad_server.FCMacro) macro inside FreeCAD (listens on port 12345) — no extra addon installation required.
  - Headless mode: no macro needed, but `freecadcmd` or FreeCAD's bundled Python interpreter must be available.

## Build & Run

```bash
go build -o freecad-mcp-go .
```

### Start the FreeCAD GUI Macro (GUI Mode)

The bundled [freecad_server.FCMacro](freecad_server.FCMacro) starts a socket server inside FreeCAD (`127.0.0.1:12345`) and dispatches received Python code to the GUI main thread. Steps:

1. Launch FreeCAD
2. Menu **Macro → Macros...**
3. Click **"..."** to browse and select `freecad_server.FCMacro` from this repo
4. Click **Execute**
5. The following console output in FreeCAD means it's ready:

```
MCP Socket Server listening on 127.0.0.1:12345
MCP Server macro loaded. Waiting for connections...
```

Note: the macro must be started **before** the MCP server connects (the server auto-detects the socket on first call; if not found, it falls back to headless mode).

### Environment Variables

| Variable | Description | Default |
| --- | --- | --- |
| `FREECAD_CMD` | Path to the `freecadcmd` executable, used for headless mode and locating the FreeCAD Resources directory | `/Applications/FreeCAD.app/Contents/Resources/bin/freecadcmd` (macOS) |

### MCP Client Configuration

For Claude Desktop / Trae and similar clients, add to your MCP config:

```json
{
  "mcpServers": {
    "freecad": {
      "command": "/path/to/freecad-mcp-go",
      "env": {
        "FREECAD_CMD": "/Applications/FreeCAD.app/Contents/Resources/bin/freecadcmd"
      }
    }
  }
}
```

On startup the log prints the connection mode and tool count, e.g.:

```
FreeCAD MCP Server (go) running (cmd: /Applications/FreeCAD.app/.../freecadcmd, tools: 165)
```

## Tool List (165)

| Module | Count | Tools |
| --- | --- | --- |
| Document | 6 | `freecad_new_document`, `freecad_open_document`, `freecad_save_document`, `freecad_close_document`, `freecad_list_objects`, `freecad_get_object_info` |
| Primitives | 11 | `freecad_create_box`, `freecad_create_cylinder`, `freecad_create_sphere`, `freecad_create_cone`, `freecad_create_torus`, `freecad_create_wedge`, `freecad_create_helix`, `freecad_create_spiral`, `freecad_create_ellipsoid`, `freecad_create_tube`, `freecad_create_prism` |
| Sketcher | 16 | `freecad_create_sketch`, `freecad_add_sketch_line`, `freecad_add_sketch_circle`, `freecad_add_sketch_arc`, `freecad_add_sketch_rectangle`, `freecad_add_sketch_constraint`, `freecad_add_sketch_ellipse`, `freecad_add_sketch_bspline`, `freecad_add_sketch_polygon`, `freecad_add_sketch_slot`, `freecad_add_sketch_point`, `freecad_add_sketch_external`, `freecad_sketch_fillet`, `freecad_sketch_trim`, `freecad_sketch_toggle_construction`, `freecad_close_sketch` |
| Operations | 17 | `freecad_boolean_fuse`, `freecad_boolean_cut`, `freecad_boolean_intersect`, `freecad_boolean_fragments`, `freecad_boolean_xor`, `freecad_fillet`, `freecad_chamfer`, `freecad_move_object`, `freecad_rotate_object`, `freecad_copy_object`, `freecad_delete_object`, `freecad_mirror_object`, `freecad_check_geometry`, `freecad_refine_shape`, `freecad_slice`, `freecad_join_connect`, `freecad_join_cutout` |
| Part Design | 20 | `freecad_pad`, `freecad_pocket`, `freecad_revolve`, `freecad_loft`, `freecad_sweep`, `freecad_partdesign_fillet`, `freecad_partdesign_chamfer`, `freecad_hole`, `freecad_partdesign_thickness`, `freecad_linear_pattern`, `freecad_polar_pattern`, `freecad_partdesign_mirrored`, `freecad_additive_helix`, `freecad_subtractive_helix`, `freecad_additive_pipe`, `freecad_subtractive_pipe`, `freecad_groove`, `freecad_draft_angle`, `freecad_multi_transform`, `freecad_shape_binder` |
| Advanced | 11 | `freecad_thickness`, `freecad_offset_3d`, `freecad_section`, `freecad_compound`, `freecad_linear_array`, `freecad_polar_array`, `freecad_scale_object`, `freecad_extrude`, `freecad_get_center_of_mass`, `freecad_get_face_info`, `freecad_get_edge_info` |
| Import/Export & Measurement | 17 | `freecad_import_step`, `freecad_import_stl`, `freecad_import_iges`, `freecad_import_dxf`, `freecad_import_svg`, `freecad_export_step`, `freecad_export_stl`, `freecad_export_obj`, `freecad_export_iges`, `freecad_export_dxf`, `freecad_export_svg`, `freecad_export_brep`, `freecad_measure_distance`, `freecad_measure_angle`, `freecad_get_volume`, `freecad_get_bounding_box`, `freecad_execute_python` |
| Mesh | 7 | `freecad_mesh_from_shape`, `freecad_mesh_to_shape`, `freecad_mesh_repair`, `freecad_mesh_decimate`, `freecad_mesh_refine`, `freecad_mesh_info`, `freecad_mesh_boolean` |
| Surface | 5 | `freecad_surface_filling`, `freecad_surface_geomfill`, `freecad_surface_sections`, `freecad_surface_extend`, `freecad_surface_ruled` |
| Draft | 17 | `freecad_draft_wire`, `freecad_draft_bspline`, `freecad_draft_polygon`, `freecad_draft_ellipse`, `freecad_draft_rectangle`, `freecad_draft_facebinder`, `freecad_draft_clone`, `freecad_draft_shapestring`, `freecad_draft_move`, `freecad_draft_rotate`, `freecad_draft_scale`, `freecad_draft_offset`, `freecad_draft_upgrade`, `freecad_draft_downgrade`, `freecad_draft_path_array`, `freecad_draft_dimension` |
| BIM/Arch | 9 | `freecad_arch_wall`, `freecad_arch_structure`, `freecad_arch_window`, `freecad_arch_floor`, `freecad_arch_building`, `freecad_arch_site`, `freecad_arch_roof`, `freecad_arch_stairs`, `freecad_export_ifc` |
| FEM | 8 | `freecad_fem_analysis`, `freecad_fem_material`, `freecad_fem_constraint_fixed`, `freecad_fem_constraint_force`, `freecad_fem_constraint_pressure`, `freecad_fem_mesh`, `freecad_fem_solver`, `freecad_fem_results` |
| TechDraw | 6 | `freecad_techdraw_create_page`, `freecad_techdraw_add_view`, `freecad_techdraw_add_projection_group`, `freecad_techdraw_add_dimension`, `freecad_techdraw_export_svg`, `freecad_techdraw_export_dxf` |
| Assembly | 10 | `freecad_assembly_create`, `freecad_assembly_insert_component`, `freecad_assembly_toggle_grounded`, `freecad_assembly_fixed_joint`, `freecad_assembly_revolute_joint`, `freecad_assembly_slider_joint`, `freecad_assembly_cylindrical_joint`, `freecad_assembly_ball_joint`, `freecad_assembly_distance_joint`, `freecad_assembly_solve` |
| Spreadsheet/Expressions | 5 | `freecad_spreadsheet_create`, `freecad_spreadsheet_set`, `freecad_spreadsheet_get`, `freecad_spreadsheet_alias`, `freecad_set_expression` |

## Project Structure

```
freecad-mcp-go/
├── main.go                 # MCP server entry: registers tools, serves over stdio
├── freecad_server.FCMacro  # FreeCAD GUI macro: starts the port-12345 socket server
├── go.mod / go.sum         # Go module (depends on mark3labs/mcp-go)
└── tools/                  # Core implementation
    ├── bridge.go           # FreeCAD Bridge: socket / headless dual-mode execution
    ├── registry.go         # Tool registry, embeds schemas.json
    ├── schemas.json        # 165 tool descriptions + JSON Schemas generated from the TS source
    ├── validation.go       # Argument validation consistent with the TS version
    ├── types.go            # Shared types
    ├── parity_test.go      # Parity tests (tool count / schema alignment)
    └── *.go                # Per-module tool implementations (document, sketcher, partdesign, etc.)
```

## Tests

```bash
go test ./...
```

`TestToolParity` verifies: every schema has a Go handler, every handler has a schema, no duplicates, and the total is exactly 165 (same as the TS source).

## Acknowledgements

- [sergiudanstan/freecad-mcp](https://github.com/sergiudanstan/freecad-mcp) — the original TypeScript implementation; this project's tool protocol, schemas, and the `freecad_server.FCMacro` macro all derive from it.
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) — Go implementation of the MCP protocol.
- [FreeCAD](https://www.freecad.org/) — the open-source parametric 3D CAD modeler.

## License

Distributed under the [MIT License](LICENSE). Because this project contains derivative code from upstream `sergiudanstan/freecad-mcp` (MIT), the LICENSE also preserves the upstream copyright notice:

```
Copyright (c) 2026 aosfather
Copyright (c) 2024-2025 Dan Stan (freecad-mcp, https://github.com/sergiudanstan/freecad-mcp)
```
