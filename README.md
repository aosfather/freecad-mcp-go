# freecad-mcp-go

**中文** | [English](README.en.md)

> 本项目是 [https://github.com/sergiudanstan/freecad-mcp.git](https://github.com/sergiudanstan/freecad-mcp.git) 的 Go 语言移植。

一个使用 Go 编写的 FreeCAD MCP（Model Context Protocol）服务器。它是 [sergiudanstan/freecad-mcp](https://github.com/sergiudanstan/freecad-mcp) TypeScript 版本的移植，提供与原版完全相同的 **165 个工具**（工具描述与 JSON Schema 直接从 TS 源码的构建产物生成，并通过一致性测试保证对齐），让 AI 助手（如 Claude、Trae 等 MCP 客户端）能够直接驱动 FreeCAD 完成参数化建模、装配、有限元分析等任务。

## 功能特性

- **165 个工具全覆盖**：涵盖文档管理、草图、零件设计、布尔/变换操作、网格、曲面、Draft、BIM/Arch、FEM、TechDraw、装配、电子表格等 FreeCAD 主要工作模块。
- **双模式连接 FreeCAD**：
  - **GUI Socket 模式**（推荐）：连接正在运行的 FreeCAD 图形界面（需安装原版 freecad-mcp 插件，监听 `127.0.0.1:12345`），可实时看到建模过程。
  - **无头模式**（自动回退）：GUI 不可用时自动启动 FreeCAD 自带的 `freecadcmd` Python REPL 执行代码，适合服务器/CI 环境。
  - 会话中途 socket 断开会自动切换到无头模式。
- **stdio MCP 协议**：基于 [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) 实现，可直接接入任意支持 stdio 的 MCP 客户端。
- **内置参数校验**：与 TS 版一致的工具参数校验逻辑（见 `tools/validation.go`）。
- **工具数量一致性测试**：`tools/parity_test.go` 保证注册的处理器与 `schemas.json` 中的 schema 一一对应，且总数与 TS 源一致（165）。

## 工作原理

```
┌──────────────┐  MCP (stdio)  ┌──────────────────┐   TCP 12345   ┌─────────────────┐
│  MCP 客户端   │ ◄───────────► │  freecad-mcp-go  │ ◄───────────► │ FreeCAD GUI      │
│ (Claude 等)  │               │                  │  (MCP 插件)   │ (freecad-mcp 宏) │
└──────────────┘               └────────┬─────────┘               └─────────────────┘
                                        │ socket 不可用时自动回退
                                        ▼
                               ┌──────────────────┐
                               │ freecadcmd 无头   │
                               │ Python REPL      │
                               └──────────────────┘
```

- 每次 MCP 工具调用，服务器会生成对应的 FreeCAD Python 代码并通过 Bridge 执行。
- 结果以 JSON 返回，`success/result/error/output` 字段与 TS 版协议保持一致。
- 全局超时 30 秒。

## 环境要求

- **Go** 1.26+（仅构建时需要）
- **FreeCAD**（建议 1.0+）
  - GUI 模式：在 FreeCAD 中执行仓库自带的 [freecad_server.FCMacro](freecad_server.FCMacro) 宏（监听 12345 端口），无需额外安装插件。
  - 无头模式：无需宏，但需要 `freecadcmd` 或 FreeCAD 自带的 Python 解释器可用。

## 构建与运行

```bash
go build -o freecad-mcp-go .
```

### 启动 FreeCAD GUI 宏（GUI 模式）

仓库自带的 [freecad_server.FCMacro](freecad_server.FCMacro) 会在 FreeCAD 内启动一个 socket 服务（`127.0.0.1:12345`），把收到的 Python 代码调度到 GUI 主线程执行。使用步骤：

1. 启动 FreeCAD
2. 菜单 **宏 (Macro) → 宏... (Macros...)**
3. 点击 **"..."** 浏览，选择仓库中的 `freecad_server.FCMacro`
4. 点击 **执行 (Execute)**
5. FreeCAD 控制台出现以下输出即表示就绪：

```
MCP Socket Server listening on 127.0.0.1:12345
MCP Server macro loaded. Waiting for connections...
```

注意：宏需要**先于** MCP 服务器启动（服务器会在首次调用时自动探测 socket；探测不到则回退无头模式）。

### 环境变量

| 变量 | 说明 | 默认值 |
| --- | --- | --- |
| `FREECAD_CMD` | `freecadcmd` 可执行文件路径，用于无头模式及定位 FreeCAD Resources 目录 | `/Applications/FreeCAD.app/Contents/Resources/bin/freecadcmd`（macOS） |

### 配置 MCP 客户端

以 Claude Desktop / Trae 等客户端为例，在 MCP 配置中添加：

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

启动后日志会输出连接模式与工具数量，例如：

```
FreeCAD MCP Server (go) running (cmd: /Applications/FreeCAD.app/.../freecadcmd, tools: 165)
```

## 工具列表（165 个）

| 模块 | 数量 | 工具 |
| --- | --- | --- |
| 文档管理 | 6 | `freecad_new_document`、`freecad_open_document`、`freecad_save_document`、`freecad_close_document`、`freecad_list_objects`、`freecad_get_object_info` |
| 基本体 | 11 | `freecad_create_box`、`freecad_create_cylinder`、`freecad_create_sphere`、`freecad_create_cone`、`freecad_create_torus`、`freecad_create_wedge`、`freecad_create_helix`、`freecad_create_spiral`、`freecad_create_ellipsoid`、`freecad_create_tube`、`freecad_create_prism` |
| 草图 | 16 | `freecad_create_sketch`、`freecad_add_sketch_line`、`freecad_add_sketch_circle`、`freecad_add_sketch_arc`、`freecad_add_sketch_rectangle`、`freecad_add_sketch_constraint`、`freecad_add_sketch_ellipse`、`freecad_add_sketch_bspline`、`freecad_add_sketch_polygon`、`freecad_add_sketch_slot`、`freecad_add_sketch_point`、`freecad_add_sketch_external`、`freecad_sketch_fillet`、`freecad_sketch_trim`、`freecad_sketch_toggle_construction`、`freecad_close_sketch` |
| 建模操作 | 17 | `freecad_boolean_fuse`、`freecad_boolean_cut`、`freecad_boolean_intersect`、`freecad_boolean_fragments`、`freecad_boolean_xor`、`freecad_fillet`、`freecad_chamfer`、`freecad_move_object`、`freecad_rotate_object`、`freecad_copy_object`、`freecad_delete_object`、`freecad_mirror_object`、`freecad_check_geometry`、`freecad_refine_shape`、`freecad_slice`、`freecad_join_connect`、`freecad_join_cutout` |
| 零件设计 | 20 | `freecad_pad`、`freecad_pocket`、`freecad_revolve`、`freecad_loft`、`freecad_sweep`、`freecad_partdesign_fillet`、`freecad_partdesign_chamfer`、`freecad_hole`、`freecad_partdesign_thickness`、`freecad_linear_pattern`、`freecad_polar_pattern`、`freecad_partdesign_mirrored`、`freecad_additive_helix`、`freecad_subtractive_helix`、`freecad_additive_pipe`、`freecad_subtractive_pipe`、`freecad_groove`、`freecad_draft_angle`、`freecad_multi_transform`、`freecad_shape_binder` |
| 高级操作 | 11 | `freecad_thickness`、`freecad_offset_3d`、`freecad_section`、`freecad_compound`、`freecad_linear_array`、`freecad_polar_array`、`freecad_scale_object`、`freecad_extrude`、`freecad_get_center_of_mass`、`freecad_get_face_info`、`freecad_get_edge_info` |
| 导入/导出与测量 | 17 | `freecad_import_step`、`freecad_import_stl`、`freecad_import_iges`、`freecad_import_dxf`、`freecad_import_svg`、`freecad_export_step`、`freecad_export_stl`、`freecad_export_obj`、`freecad_export_iges`、`freecad_export_dxf`、`freecad_export_svg`、`freecad_export_brep`、`freecad_measure_distance`、`freecad_measure_angle`、`freecad_get_volume`、`freecad_get_bounding_box`、`freecad_execute_python` |
| 网格 | 7 | `freecad_mesh_from_shape`、`freecad_mesh_to_shape`、`freecad_mesh_repair`、`freecad_mesh_decimate`、`freecad_mesh_refine`、`freecad_mesh_info`、`freecad_mesh_boolean` |
| 曲面 | 5 | `freecad_surface_filling`、`freecad_surface_geomfill`、`freecad_surface_sections`、`freecad_surface_extend`、`freecad_surface_ruled` |
| Draft（制图） | 17 | `freecad_draft_wire`、`freecad_draft_bspline`、`freecad_draft_polygon`、`freecad_draft_ellipse`、`freecad_draft_rectangle`、`freecad_draft_facebinder`、`freecad_draft_clone`、`freecad_draft_shapestring`、`freecad_draft_move`、`freecad_draft_rotate`、`freecad_draft_scale`、`freecad_draft_offset`、`freecad_draft_upgrade`、`freecad_draft_downgrade`、`freecad_draft_path_array`、`freecad_draft_dimension` |
| BIM/Arch（建筑） | 9 | `freecad_arch_wall`、`freecad_arch_structure`、`freecad_arch_window`、`freecad_arch_floor`、`freecad_arch_building`、`freecad_arch_site`、`freecad_arch_roof`、`freecad_arch_stairs`、`freecad_export_ifc` |
| FEM（有限元） | 8 | `freecad_fem_analysis`、`freecad_fem_material`、`freecad_fem_constraint_fixed`、`freecad_fem_constraint_force`、`freecad_fem_constraint_pressure`、`freecad_fem_mesh`、`freecad_fem_solver`、`freecad_fem_results` |
| TechDraw（工程图） | 6 | `freecad_techdraw_create_page`、`freecad_techdraw_add_view`、`freecad_techdraw_add_projection_group`、`freecad_techdraw_add_dimension`、`freecad_techdraw_export_svg`、`freecad_techdraw_export_dxf` |
| 装配 | 10 | `freecad_assembly_create`、`freecad_assembly_insert_component`、`freecad_assembly_toggle_grounded`、`freecad_assembly_fixed_joint`、`freecad_assembly_revolute_joint`、`freecad_assembly_slider_joint`、`freecad_assembly_cylindrical_joint`、`freecad_assembly_ball_joint`、`freecad_assembly_distance_joint`、`freecad_assembly_solve` |
| 电子表格/表达式 | 5 | `freecad_spreadsheet_create`、`freecad_spreadsheet_set`、`freecad_spreadsheet_get`、`freecad_spreadsheet_alias`、`freecad_set_expression` |

## 项目结构

```
freecad-mcp-go/
├── main.go                 # MCP 服务器入口：注册工具、启动 stdio 服务
├── freecad_server.FCMacro  # FreeCAD GUI 宏：在 FreeCAD 内启动 12345 端口 socket 服务
├── go.mod / go.sum         # Go 模块定义（依赖 mark3labs/mcp-go）
└── tools/                  # 核心实现
    ├── bridge.go           # FreeCAD Bridge：socket / 无头双模式执行
    ├── registry.go         # 工具注册表，嵌入 schemas.json
    ├── schemas.json        # 从 TS 源码生成的 165 个工具描述 + JSON Schema
    ├── validation.go       # 与 TS 版一致的参数校验
    ├── types.go            # 共享类型
    ├── parity_test.go      # 一致性测试（工具数量 / schema 对齐）
    └── *.go                # 各模块工具实现（document、sketcher、partdesign 等）
```

## 测试

```bash
go test ./...
```

`TestToolParity` 会校验：每个 schema 都有对应的 Go 处理器、每个处理器都有 schema、无重复注册，且总数恰好为 165（与 TS 源一致）。

## 致谢

- [sergiudanstan/freecad-mcp](https://github.com/sergiudanstan/freecad-mcp) —— 原始 TypeScript 实现，本项目的工具协议、schema 与 `freecad_server.FCMacro` 宏均源自该项目。
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) —— Go 语言的 MCP 协议实现。
- [FreeCAD](https://www.freecad.org/) —— 开源参数化三维 CAD 建模器。

## 开源协议

本项目以 [MIT License](LICENSE) 分发。由于包含上游 `sergiudanstan/freecad-mcp`（MIT）的衍生代码，LICENSE 中同时保留了上游的版权声明：

```
Copyright (c) 2026 aosfather
Copyright (c) 2024-2025 Dan Stan (freecad-mcp, https://github.com/sergiudanstan/freecad-mcp)
```
