# agentutil

A standalone Go library for Language Server Protocol (LSP) integration in AI agents. Extracted from [charmbracelet/crush](https://github.com/charmbracelet/crush).

## Features

- **Auto-detection**: Automatically detects installed LSP servers via `exec.LookPath`
- **Lazy initialization**: Starts LSP clients only when relevant files are accessed
- **3 tools**: `lsp_diagnostics`, `lsp_references`, `lsp_restart` — ready for agent integration
- **Pure stdio transport**: Uses [powernap](https://github.com/charmbracelet/x/tree/main/powernap) for JSON-RPC over stdin/stdout
- **Standalone**: No external agent framework required — works with any Go codebase

## Installation

```bash
go get github.com/twistedogic/agentutil
```

## Quick Start

```go
package main

import (
    "context"
    "log/slog"
    "github.com/twistedogic/agentutil/lsp"
)

func main() {
    // Config store providing LSP server definitions
    store := &myConfigStore{}
    workDir := "/path/to/project"

    manager := lsp.NewManager(store, workDir)

    // Set callback to receive client lifecycle events
    manager.SetCallback(func(name string, client *lsp.Client) {
        slog.Info("LSP client event", "name", name, "state", client.GetServerState())
    })

    ctx := context.Background()

    // Start LSP for a file — manager auto-detects the right server
    manager.Start(ctx, "/path/to/project/main.go")

    // Access running clients
    for name, client := range manager.Clients().Seq2() {
        slog.Info("Running LSP", "name", name, "files", client.FileTypes())
        diags := client.GetDiagnostics()
        for uri, diagList := range diags {
            slog.Info("Diagnostics", "uri", uri, "count", len(diagList))
        }
    }
}
```

## Configuration

### ConfigStore Interface

Implement `ConfigStore` to provide LSP server definitions:

```go
type ConfigStore interface {
    LSP() map[string]ServerConfig  // LSP server configs keyed by name
    AutoLSP() *bool                 // nil means auto-start enabled
    Resolver() VariableResolver     // resolves variable references
}

type VariableResolver interface {
    ResolveValue(v string) (string, error)
}

type ServerConfig struct {
    Command     string
    Args        []string
    Environment map[string]string
    FileTypes   []string
    RootMarkers []string
    InitOptions map[string]any
    Settings    map[string]any
    Timeout     int
}
```

### Example: JSON config store

```go
type JSONConfigStore struct {
    servers map[string]lsp.ServerConfig
    autoLSP bool
}

func (c *JSONConfigStore) LSP() map[string]lsp.ServerConfig {
    return c.servers
}

func (c *JSONConfigStore) AutoLSP() *bool {
    return &c.autoLSP
}

func (c *JSONConfigStore) Resolver() lsp.VariableResolver {
    return &envResolver{}
}

type envResolver struct{}

func (e *envResolver) ResolveValue(v string) (string, error) {
    // Expand ${VAR} or $VAR from environment
    if strings.HasPrefix(v, "${") && strings.HasSuffix(v, "}") {
        return os.Getenv(v[2:len(v)-1]), nil
    }
    if strings.HasPrefix(v, "$") {
        return os.Getenv(v[1:]), nil
    }
    return v, nil
}
```

## Architecture

```
agentutil/
├── lsp/
│   ├── manager.go      # Manager: server detection, lifecycle, callbacks
│   ├── client.go       # Client: powernap wrapper, diagnostics cache
│   ├── handlers.go     # LSP notification/request handlers
│   └── versioned.go    # Thread-safe versioned map for diagnostics
├── config/
│   ├── config.go       # ServerConfig type definitions
│   └── types.go        # ResolvedConfig for Goof-based apps
└── tools/
    └── eino/            # (planned) eino adapter for tool integration
```

### Manager

- Loads default LSP servers from powernap's registry
- Merges user-configured servers
- Starts servers lazily when files are accessed
- Tracks unavailable servers with retry backoff
- Skips generic commands (node, python, etc.) unless user-configured

### Client

- Wraps powernap JSON-RPC client (stdio transport)
- Caches diagnostics in a VersionedMap (lock-free reads)
- Supports file open/close/change notifications
- Implements graceful restart

### Handlers

- `workspace/applyEdit` — placeholder for code edit application
- `workspace/configuration` — returns empty config (handled by client)
- `client/registerCapability` — tracks file watcher registrations
- `window/showMessage` — logs server messages
- `textDocument/publishDiagnostics` — updates diagnostic cache

## Integration with Agent Frameworks

### Fantasy (Charm)

Crush uses [charm.land/fantasy](https://charm.sh) for tool registration. The LSP tools are defined in `internal/agent/tools/`:

- `diagnostics.go` — `lsp_diagnostics` tool
- `references.go` — `lsp_references` tool  
- `lsp_restart.go` — `lsp_restart` tool

### eino (CloudWeGo)

Integration via `tools/eino/` package (planned):

```go
// Planned API
manager := lsp.NewManager(store, workDir)
tools := eino.NewLSPRegistry(manager)

agent, _ := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
    Model: chatModel,
    ToolsConfig: tools,
})
```

## Key Design Decisions

1. **Standalone**: No agent framework coupling — works with fantasy, eino, or custom agents
2. **Powernap**: JSON-RPC over stdio — same transport as crush, no custom protocol
3. **ConfigStore interface**: Decouples configuration from LSP logic — any config source works
4. **VersionedMap**: Lock-free reads for diagnostics — high-frequency UI updates without contention
5. **Lazy start**: Only starts servers when files are accessed — no wasted resources

## TODO

- [ ] `tools/eino/` — eino ADK tool adapter
- [ ] `util/edit.go` — workspace edit application (requires filesystem write implementation)
- [ ] Tests for manager and client
- [ ] CI with `go test ./...`
- [ ] GitHub Actions workflow

## Credits

Extracted from [charmbracelet/crush](https://github.com/charmbracelet/crush) with minimal modifications to decouple from the agent framework. Uses [powernap](https://github.com/charmbracelet/x/tree/main/powernap) for the LSP transport layer.

## License

MIT