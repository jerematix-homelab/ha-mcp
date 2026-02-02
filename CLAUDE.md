# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

ha-mcp is a Model Context Protocol (MCP) server that provides AI assistants with access to Home Assistant. It uses a hybrid architecture: WebSocket for most operations, REST API for automation/script/scene CRUD (create/update/delete). Translates MCP tool calls into Home Assistant API commands.

**Requirements:** Go 1.25+, golangci-lint v2

## Build and Development Commands

```bash
# Build the binary
go build -o ha-mcp.exe ./cmd/ha-mcp

# Run unit tests 
go test ./...

# Run a single test
go test -v ./internal/handlers -run TestEntityHandlers

# Run integration tests (requires HA_INTEGRATION_TEST_URL and HA_INTEGRATION_TEST_TOKEN)
go test -tags=integration -v ./internal/handlers/integration/...

# Run linter (uses golangci-lint v2)
golangci-lint run ./...

# Check formatting
gofmt -l .

# Run security vulnerability scan
govulncheck ./...

# Initialize config files in current directory
./ha-mcp init

# Display effective configuration (tokens masked)
./ha-mcp config

# Run the server
./ha-mcp --ha-url http://homeassistant.local:8123 --ha-token YOUR_TOKEN

# Docker: Pull and run
docker pull zorak1103/ha-mcp:latest
docker run -p 8080:8080 -e HA_URL=http://homeassistant.local:8123 zorak1103/ha-mcp:latest

# Docker: Local snapshot build (builds + creates images)
goreleaser release --snapshot --clean --skip=publish
```

## Architecture

### Request Flow

```
AI Client (Claude, Cline)
    → HTTP POST / (JSON-RPC)
    → MCP Server (internal/mcp/server.go)
    → Tool Registry lookup (internal/mcp/registry.go)
    → Tool Handler (internal/handlers/*.go)
    → HybridClient (internal/homeassistant/hybrid_client.go)
        → WebSocket (most operations) OR REST API (automation/script/scene CRUD)
    → Home Assistant API
```

### Key Packages

- **cmd/ha-mcp**: CLI entry point using Cobra, handles flags and signals
- **internal/mcp**: MCP protocol server, JSON-RPC handling, tool/resource registry
- **internal/homeassistant**: Hybrid client (WS + REST), WebSocket with auto-reconnect, REST for automation/script/scene CRUD
- **internal/handlers**: MCP tool handlers organized by domain (entities, automations, helpers, analysis, etc.)
- **internal/config**: Viper-based config loading (YAML → .env → ENV → CLI flags)
- **internal/logging**: Structured logging with DEBUG/INFO/WARN/ERROR/TRACE levels

### Handler Pattern

Each handler domain follows this pattern:
1. Create handler struct with `New*Handlers()` factory
2. Implement `RegisterTools(registry *mcp.Registry)` method
3. Register in `internal/handlers/register.go` via `RegisterAllTools()`

Tool handlers have signature:
```go
func(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error)
```

### Home Assistant Client Interface

The `homeassistant.Client` interface abstracts all HA operations. Implementation uses a hybrid approach:
- **HybridClient** (`hybrid_client.go`): Routes to WebSocket or REST based on operation type
- **WebSocket** (`ws_client_impl.go`): Persistent connection with auto-reconnect (1s → 60s backoff)
- **REST** (`rest_client.go`): Used for automation/script/scene CRUD operations
- **Retry** (`retry.go`): Automatic retries with exponential backoff for transient failures (5xx, 429, network errors)

**API Routing:**
- WebSocket: Helpers (input_*, counter, timer, schedule), state queries, service calls, registry access, config entries
- REST: Automation/Script/Scene create/update/delete, template rendering, logbook, config validation

**Limitations:**
- Scripts require `script.reload` after REST create/update for entity to appear
- Automations/Scenes: REST stores config but entity may require HA restart

**Config Entry Helpers:** threshold, derivative, integration, group, template use HTTP-based Config Entry Flow (automatically handled by HybridClient).

Factory pattern in `factory.go` creates the appropriate client based on configuration.

### Configuration Priority

`CLI flags > ENV vars > .env file > config.yaml > defaults`

Key environment variables:
- **Connection**: `HA_URL`, `HA_TOKEN`, `HA_MCP_PORT`, `HA_MCP_LOG_LEVEL`
- **REST Rate Limiting**: `HA_REST_RATE_LIMIT`, `HA_REST_RATE_BURST`
- **REST Retry**: `HA_REST_MAX_RETRIES` (default: 3), `HA_REST_RETRY_INITIAL_DELAY_MS` (default: 100), `HA_REST_RETRY_MAX_DELAY_MS` (default: 5000)
- **WebSocket Retry**: `HA_WS_MAX_RETRIES` (default: 3), `HA_WS_RETRY_INITIAL_DELAY_MS` (default: 100), `HA_WS_RETRY_MAX_DELAY_MS` (default: 5000)
- **Caching**: `HA_CACHE_ENABLED` (default: false), `HA_CACHE_SERVICES_TTL_MIN` (default: 60), `HA_CACHE_CONFIG_TTL_MIN` (default: 30), `HA_CACHE_ENTITY_REG_TTL_MIN` (default: 10), `HA_CACHE_DEVICE_REG_TTL_MIN` (default: 10), `HA_CACHE_AREA_REG_TTL_MIN` (default: 30)

### Testing

Test files (`*_test.go`) are excluded from funlen, gocognit, errcheck, gosec, and gocritic linters. Shared test utilities are in `internal/handlers/testing_helpers_test.go` (mock clients, result parsing helpers).

### Integration Tests

Integration tests in `internal/handlers/integration/` verify write operations against a real Home Assistant instance.

**Key files:**
- `helpers.go` - Test ID generation with `__mcptest_` prefix
- `cleanup.go` - Cleanup utilities with retry logic
- `suite_test.go` - Base test suite with config loading
- `*_integration_test.go` - Domain-specific tests (counters, timers, automations, etc.)

**Running integration tests:**
```bash
# Set environment variables
export HA_INTEGRATION_TEST_URL=http://homeassistant.local:8123
export HA_INTEGRATION_TEST_TOKEN=<your-token>

# Run all integration tests
go test -tags=integration -v ./internal/handlers/integration/...

# Run specific test suite
go test -tags=integration -v ./internal/handlers/integration/... -run TestCounterIntegration
```

**Safety:** All test entities use `mcptest_<uuid>_<name>` prefix. Tests are skipped if environment variables are not set.

## Workflow Preferences

**IMPORTANT: ALWAYS use Sub-Agents (Task tool) for the following tasks:**

- **Code Exploration & Research**: `subagent_type=Explore` for codebase navigation, file search, architecture understanding
- **Code Reviews**: `subagent_type=code-reviewer` after implementations
- **Run Tests**: `subagent_type=test-runner` after code changes
- **Build Validation**: `subagent_type=build-validator` before commits
- **Complexity Analysis**: `subagent_type=complexity-analyzer` for refactoring
- **Architecture Planning**: `subagent_type=code-architect` for new features

**Parallelization:** Execute multiple independent agent tasks in parallel (multiple Task tool calls in one message).

### Test-Driven Development (TDD)

- **Required**: Tests MUST be written BEFORE writing or modifying code
- **Workflow**:
  1. Write a test for the desired behavior (test fails - Red)
  2. Implement minimal code to make the test pass (Green)
  3. Refactor code while tests continue to pass (Refactor)
  4. Run `golangci-lint run ./...` to ensure code quality
- Tests define expected behavior BEFORE implementation
- Code is iteratively adjusted until all tests pass
- **Linting**: Always run `golangci-lint run ./...` after implementation to catch issues early
