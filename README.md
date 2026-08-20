# remote-agent-mcp

A Go-based [Model Context Protocol](https://modelcontextprotocol.io) (MCP) server
that exposes file and shell tools over **Streamable HTTP**, for remote machine
troubleshooting by a local AI client.

## Tools

| Tool | Purpose |
|------|---------|
| `read` | Read a file with line-based `offset`/`limit` paging and automatic truncation |
| `write` | Create or overwrite a file (creates parent directories) |
| `edit` | Precise `oldString → newString` replacement, optional `replaceAll` |
| `list` | List a directory (single level, includes hidden files) |
| `bash` | Run a shell command (tests, installs, lint, etc.) |

## Build

```sh
go build -o remote-agent-mcp .
```

### Cross-compiling

Use the [`build.sh`](build.sh) script to build for the local platform or
cross-compile for other OS/architectures. All artifacts are written to the
`output/` directory.

```sh
./build.sh                 # build for the current platform
./build.sh linux-arm64     # cross-compile for Linux ARM64
./build.sh all             # build every supported platform
./build.sh clean           # remove the output/ directory
```

Supported targets: `local`, `linux-amd64`, `linux-arm64`, `darwin-amd64`,
`darwin-arm64`, `windows-amd64`, `all`, `clean`. Artifacts are named
`remote-agent-mcp-<os>-<arch>` (Windows adds a `.exe` suffix).

## Run

```sh
./remote-agent-mcp --addr 0.0.0.0:9090 --root /var/log
```

### Configuration

Configuration is read from environment variables, with command-line flags taking
precedence.

| Flag | Env var | Default | Meaning |
|------|---------|---------|---------|
| `--addr` | `MCP_ADDR` | `0.0.0.0:9090` | Listen address |
| `--root` | `MCP_ROOT` | current directory | Root directory file tools are restricted to |
| `--shell` | `MCP_SHELL` | `/bin/bash` | Shell used by `bash` (auto-falls back to `/bin/sh`) |
| `--token` | `MCP_TOKEN` | *(empty)* | Require this bearer token on every request |
| `--allow-all` | `MCP_ALLOW_ALL` | `false` | Allow full filesystem access outside `--root` |
| `--allow-unconfirmed` | `MCP_ALLOW_UNCONFIRMED` | `false` | Skip elicitation confirmation (see below) |

## Safety model

- **Path containment** — `read`/`write`/`edit`/`list` resolve paths inside
  `--root`; `../` escapes are rejected unless `--allow-all` is set. Containment
  is lexical (does not follow symlinks).
- **Elicitation (human-in-the-loop)** — when the client supports it, destructive
  actions ask the user for confirmation:
  - `write` overwriting an existing file,
  - `bash` running a dangerous command (`rm -rf`, `dd`, `mkfs`, `fdisk`,
    `shutdown`, `reboot`, `sudo`, `chmod`/`chown`, overwrite redirection, …).
- **Unsupported-client fallback** — if the client does not support elicitation,
  these actions are **rejected** by default. Set `--allow-unconfirmed` to skip
  confirmation and allow them.

## Connecting an AI client

Add the server to your client's MCP configuration using the
[`examples/mcp.json`](examples/mcp.json) `mcpServers` block:

- **Claude Code** — project-scoped `.mcp.json` at the repo root, or user-scoped
  via `claude mcp add --transport http remote-agent-mcp http://127.0.0.1:9090/ --header "Authorization: Bearer your-token-here"`.
- **Claude Desktop** — `claude_desktop_config.json` (`Settings → Developer →
  Edit Config`); paste the same `mcpServers` entry.

Replace `your-token-here` with the value you passed to `--token`, or omit the
`Authorization` header entirely if you run without authentication.

