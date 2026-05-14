# figma-mcp-design

Go implementation of the MCP server and WebSocket relay for the Figma bridge in the parent repository.

Pipeline:

```text
Claude Code / Cursor <-- stdio JSON-RPC --> Go MCP server <-- WebSocket --> Go relay <-- WebSocket --> Figma plugin
```

## Important Figma Runtime Note

The MCP server and WebSocket relay are rewritten in Go. The Figma plugin files under `plugin/` remain JavaScript/HTML because Figma plugin main code runs in Figma's JavaScript runtime and talks to the Figma Plugin API directly. Go cannot replace that runtime directly without a separate JS/WASM wrapper, and that wrapper would still need JavaScript glue to access `figma.*`.

## Commands

```bash
go mod tidy
make build

# MCP client config should use the compiled server binary.
./bin/linux-amd64/figma-mcp-design
```

`make build` builds the current platform binary and these 64-bit release targets:

| OS | Architecture | MCP server binary | Standalone relay binary |
| --- | --- | --- | --- |
| Linux | amd64 | `bin/linux-amd64/figma-mcp-design` | `bin/linux-amd64/figma-mcp-relay` |
| Linux | arm64 | `bin/linux-arm64/figma-mcp-design` | `bin/linux-arm64/figma-mcp-relay` |
| Windows | amd64 | `bin/windows-amd64/figma-mcp-design.exe` | `bin/windows-amd64/figma-mcp-relay.exe` |
| Windows | arm64 | `bin/windows-arm64/figma-mcp-design.exe` | `bin/windows-arm64/figma-mcp-relay.exe` |
| macOS | amd64 | `bin/darwin-amd64/figma-mcp-design` | `bin/darwin-amd64/figma-mcp-relay` |
| macOS | arm64 | `bin/darwin-arm64/figma-mcp-design` | `bin/darwin-arm64/figma-mcp-relay` |

Detect your current OS:

```bash
go env GOOS GOARCH
```

This workspace was built on `linux/amd64`, so the local MCP config examples below use `bin/linux-amd64/figma-mcp-design`.

The MCP server starts an embedded relay by default for local usage:

```text
./bin/<os>-<arch>/figma-mcp-design
  -> starts MCP stdio server
  -> starts embedded WebSocket relay on :3055 if the port is available
  -> if :3055 is already in use, keeps running and uses the existing relay
```

To disable the embedded relay:

```bash
./bin/linux-amd64/figma-mcp-design --embedded-relay=false
```

You can still run the relay as a standalone process when you want to share it across multiple MCP server processes or use a custom port. The standalone relay honors `PORT`:

```bash
PORT=3056 ./bin/linux-amd64/figma-mcp-relay
```

The MCP server accepts:

```bash
./bin/linux-amd64/figma-mcp-design --server=localhost --port=3055
./bin/linux-amd64/figma-mcp-design --ws-url=ws://localhost:3055
```

## MCP Client Config

Use the binary for your OS:

- Linux: `/path/to/figma-mcp-design/bin/linux-amd64/figma-mcp-design`
- Linux ARM64: `/path/to/figma-mcp-design/bin/linux-arm64/figma-mcp-design`
- Windows: `C:\path\to\figma-mcp-design\bin\windows-amd64\figma-mcp-design.exe`
- Windows ARM64: `C:\path\to\figma-mcp-design\bin\windows-arm64\figma-mcp-design.exe`
- macOS Intel: `/path/to/figma-mcp-design/bin/darwin-amd64/figma-mcp-design`
- macOS Apple Silicon: `/path/to/figma-mcp-design/bin/darwin-arm64/figma-mcp-design`

Cursor:
```json
{
  "mcpServers": {
    "FigmaMCPDesign": {
      "command": "/home/sonnn/Work/cursor-talk-to-figma-mcp/figma-mcp-design/bin/linux-amd64/figma-mcp-design",
      "args": []
    }
  }
}
```

Codex:

```toml
[mcp_servers.FigmaMCPDesign]
command = "/home/sonnn/Work/cursor-talk-to-figma-mcp/figma-mcp-design/bin/linux-amd64/figma-mcp-design"
args = []
```

Claude Code:

```bash
claude mcp add FigmaMCPDesign -- "/home/sonnn/Work/cursor-talk-to-figma-mcp/figma-mcp-design/bin/linux-amd64/figma-mcp-design"
```

The compiled MCP server binary includes the embedded local relay behavior, so these MCP configurations do not need a separate `cmd/relay` process for normal local usage.

## Figma Setup

1. Start the MCP server from your MCP client configuration. For local usage, this starts the relay automatically if port `3055` is free.
2. In Figma: Plugins -> Development -> Link existing plugin
3. Select `figma-mcp-design/plugin/manifest.json`
4. Run the plugin, connect to `localhost:3055`, and join a channel
5. From the MCP client, call `join_channel` with the same channel before other tools

## Implemented Tools

This Go MCP server exposes the same tool names as the original TypeScript server and forwards them to the Figma plugin:

- document and selection tools: `get_document_info`, `get_selection`, `read_my_design`, `get_node_info`, `get_nodes_info`
- creation and mutation tools: `create_rectangle`, `create_frame`, `create_text`, `set_fill_color`, `set_stroke_color`, `move_node`, `resize_node`, `delete_node`, `delete_multiple_nodes`, `clone_node`
- text and annotation tools: `scan_text_nodes`, `set_text_content`, `set_multiple_text_contents`, `get_annotations`, `set_annotation`, `set_multiple_annotations`
- components, images, layout, reactions, connectors, focus and selection tools from the original server

## Notes

- Logs are written to stderr so stdout stays reserved for MCP JSON-RPC messages.
- `cmd/mcp-server` starts a best-effort embedded relay for `localhost`; if the port is already occupied, it assumes another relay is already running and continues.
- Long-running plugin operations send `progress_update` events; the Go client resets the inactivity timeout for those requests.
- The Go server intentionally validates only required MCP arguments and leaves detailed Figma constraints to the plugin, matching the bridge-style architecture.
