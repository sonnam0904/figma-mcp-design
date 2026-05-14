# AGENTS.md

## Project Overview

This folder is a Go rewrite of the MCP server and WebSocket relay from the parent repository.

```text
Claude Code / Cursor <-- stdio --> cmd/mcp-server <-- WebSocket --> cmd/relay <-- WebSocket --> plugin/
```

`plugin/` contains the Figma runtime plugin. It remains JavaScript/HTML because Figma plugin code must run in Figma's JavaScript environment to access the `figma.*` API.

## Commands

```bash
go mod tidy
go build ./...
go run ./cmd/relay
go run ./cmd/mcp-server
```

## Architecture

- `cmd/mcp-server`: stdio MCP entrypoint.
- `internal/mcp`: minimal MCP JSON-RPC server, tool schemas, prompt support, and tool dispatch.
- `internal/figma`: WebSocket client that sends command envelopes to the relay and tracks pending request IDs.
- `cmd/relay` and `internal/relay`: channel-isolated WebSocket relay equivalent to `src/socket.ts`.
- `plugin/`: Figma plugin copied from the original source; commands are handled in `code.js`.

## Agent Notes

- Call `join_channel` before other Figma tools.
- Call `get_document_info`, `read_my_design`, or `get_selection` before modifying designs.
- **MCP scope:** The plugin-backed toolset covers frames/rectangles/text, **create_image** (base64 PNG/JPEG), **apply_style** (local styles from **get_styles**), **get_variables** / **set_variable_binding** (local Figma Variables), basic fills and strokes, auto-layout helpers, components/instances, text batch updates, annotations, exports, and reaction→connector flows. It does **not** include URL-based image fetch, arbitrary vector/boolean editing, full prototype authoring, or full library-variable management unless listed in `internal/mcp/tools.go`. Use **read_design_strategy**, **design_strategy**, and **variable_strategy** prompts as needed.
- Keep stdout clean in MCP server code; use stderr/log package for diagnostics.
- Prefer adding tools in `internal/mcp/tools.go` and forwarding through `internal/mcp/server.go`.
