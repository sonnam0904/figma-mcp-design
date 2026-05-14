# figma-mcp-design

A **Model Context Protocol (MCP)** bridge between Cursor / Claude Code (and other MCP clients) and **Figma**, using a WebSocket relay and a Figma plugin that runs in the browser.

## Requirements

| Component | Purpose |
|-----------|---------|
| **Go 1.23+** | If building from source (`go build`, `make`) |
| **Figma Desktop** | Install the development plugin and open your design file |

## Installation

### 1. Download binaries from GitHub Releases

On the repository **Releases** page, download the archive that matches the **OS and architecture of the machine where the agent runs** (for example `figma-mcp-design-v1.0.0-linux-amd64.tar.gz`, or `.zip` on Windows).

Each archive contains the **`figma-mcp-design`** binary (and **`figma-mcp-relay`** if you need to run the relay as a separate process). For MCP configuration in your IDE, use the absolute path to **`figma-mcp-design`** (on Windows: `figma-mcp-design.exe`).

Extract to a fixed folder. Releases usually ship **`SHASUMS256.txt`** for checksum verification.

### 2. From source (recommended for development)

```bash
git clone https://github.com/sonnam0904/figma-mcp-design.git
cd figma-mcp-design
go mod tidy
make build
```

- `make build` produces binaries for the **current machine** and for **six common OS/architecture pairs** under `bin/<goos>-<goarch>/`.
- After `make build`, the MCP server binary for IDE configuration is **`figma-mcp-design`** (on Windows: **`figma-mcp-design.exe`**).

Check your current OS/architecture:

```bash
go env GOOS GOARCH
```

On Linux amd64, the MCP server path is typically:

`figma-mcp-design/bin/linux-amd64/figma-mcp-design`

## Usage

### Pick the right binary

**3 STEPS**

1. Install the binary that matches the OS/architecture where your agent runs, then paste the matching snippet below into the agent's MCP configuration.

2. Locate the binary under `figma-mcp-design/bin/<os-arch>/`. Available targets: `darwin-amd64`, `darwin-arm64`, `linux-amd64`, `linux-arm64`, `windows-amd64`, `windows-arm64`.

   The snippets below are auto-filled with the binary detected for the OS this plugin is running on. If your agent runs on a different machine, swap the path for the matching binary.

   If you installed from **GitHub Release** (archive), use the path to **`figma-mcp-design`** (or **`figma-mcp-design.exe`**) in the folder you extracted instead of `bin/<os-arch>/figma-mcp-design`.

3. Paste the snippet into your agent (Cursor / Codex / Claude…), restart the agent, then return to the **Connection** tab and click **Connect**.

**Detected target:** `linux-amd64`  
**Suggested binary:** `/path-to-project/figma-mcp-design/bin/linux-amd64/figma-mcp-design`

#### Cursor — `mcp.json`

Add this to Cursor's `mcp.json`. Adjust `cwd` if the repository lives in another location.

```json
{
  "mcpServers": {
    "FigmaMCPDesign": {
      "command": "/path-to-project/figma-mcp-design/bin/linux-amd64/figma-mcp-design",
      "args": []
    }
  }
}
```

#### Codex — TOML

Add this server to your Codex MCP configuration. Uses the local stdio server built in Go.

```toml
[mcp_servers.FigmaMCPDesign]
command = "/path-to-project/figma-mcp-design/bin/linux-amd64/figma-mcp-design"
args = []
```

#### Claude — CLI

For Claude Code, register the server with this command from the repository root.

```bash
claude mcp add FigmaMCPDesign -- "/path-to-project/figma-mcp-design/bin/linux-amd64/figma-mcp-design"
```

## Figma plugin setup

1. Start the MCP server from your IDE (by default an embedded relay listens on port **3055** if the port is free).
2. In Figma: **Plugins → Development → Import plugin from manifest…** (wording may vary by Figma version).
3. Select `plugin/manifest.json` in your cloned repository.
4. Run the plugin, connect to **`localhost:3055`**, **join** a **channel** (any name you choose, e.g. `dev`).
5. In the MCP chat session, call the **`join_channel`** tool with the **same channel name** before other Figma tools.

`manifest.json` declares `ws://localhost:3055` for local development; if you change the port, update `networkAccess` accordingly.

## MCP tools (overview)

| Group | Tools |
|-------|-------|
| Channel | `join_channel` |
| Document & selection | `get_document_info`, `get_selection`, `read_my_design`, `get_node_info`, `get_nodes_info` |
| Creation & geometry | `create_rectangle`, `create_frame`, `create_text`, `set_fill_color`, `set_stroke_color`, `move_node`, `resize_node`, `clone_node`, `delete_node`, `delete_multiple_nodes`, `set_corner_radius`, `export_node_as_image` |
| Text | `set_text_content`, `scan_text_nodes`, `set_multiple_text_contents` |
| Styles & components | `get_styles`, `get_local_components`, `create_component_instance`, `get_instance_overrides`, `set_instance_overrides` |
| Annotations | `get_annotations`, `set_annotation`, `set_multiple_annotations` |
| Auto layout | `set_layout_mode`, `set_padding`, `set_axis_align`, `set_layout_sizing`, `set_item_spacing` |
| Tree scan | `scan_nodes_by_types` |
| Prototype & connectors | `get_reactions`, `set_default_connector`, `create_connections` |
| Navigation | `set_focus`, `set_selections` |

Parameter details are in `internal/mcp/tools.go` (MCP schemas).

## Directory layout (abbreviated)

| Path | Contents |
|------|----------|
| `cmd/mcp-server` | MCP entrypoint (stdio) |
| `cmd/relay` | WebSocket relay entrypoint |
| `internal/mcp` | MCP JSON-RPC, tool schemas, dispatch |
| `internal/figma` | WebSocket client to the relay |
| `internal/relay` | Channel-isolated relay hub |
| `plugin/` | Figma plugin (JS/HTML) |
| `scripts/build-release.sh` | Cross-platform release build script |

AI agent notes in the repo: **`AGENTS.md`**.

## License

MIT
