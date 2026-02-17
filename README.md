# ha-mcp

[![CI](https://github.com/zorak1103/ha-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/zorak1103/ha-mcp/actions/workflows/ci.yml)
[![Release](https://github.com/zorak1103/ha-mcp/actions/workflows/release.yml/badge.svg)](https://github.com/zorak1103/ha-mcp/actions/workflows/release.yml)
[![Renovate](https://github.com/zorak1103/ha-mcp/actions/workflows/renovate.yml/badge.svg)](https://github.com/zorak1103/ha-mcp/actions/workflows/renovate.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zorak1103/ha-mcp?style=flat)](https://goreportcard.com/report/github.com/zorak1103/ha-mcp)
[![Go Version](https://img.shields.io/github/go-mod/go-version/zorak1103/ha-mcp)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/zorak1103/ha-mcp.svg)](https://pkg.go.dev/github.com/zorak1103/ha-mcp)
[![License](https://img.shields.io/github/license/zorak1103/ha-mcp)](LICENSE)
[![GitHub release](https://img.shields.io/github/v/release/zorak1103/ha-mcp)](https://github.com/zorak1103/ha-mcp/releases/latest)
[![Docker Hub](https://img.shields.io/docker/v/zorak1103/ha-mcp?label=Docker%20Hub&logo=docker)](https://hub.docker.com/r/zorak1103/ha-mcp)

A Model Context Protocol (MCP) server that provides AI assistants with access to Home Assistant, enabling smart home control and automation management.

## Features

- **Hybrid Architecture**: Primary WebSocket communication with REST API for automation/script/scene CRUD operations
- **Entity Management**: Read and control all Home Assistant entities
- **Registry Access**: Query entity, device, and area registries
- **Automation CRUD**: Create, read, update, and delete automations (via REST API) with flexible ID lookup (entity_id, config.id, or friendly name)
- **Helper Management**: Full support for all 14 helper types via consolidated `manage_helper` and `helper_action` tools
- **Script & Scene Control**: Full CRUD operations for scripts and scenes (via REST API) with flexible ID lookup (entity_id or friendly name)
- **Service Calls**: Execute any Home Assistant service
- **History & Statistics**: Query entity state history and recorder statistics
- **Media Browser**: Browse media sources and get camera streams
- **Lovelace Config**: Access dashboard configurations
- **Auto-Reconnect**: Automatic reconnection with exponential backoff
- **Request Retries**: Automatic retries with exponential backoff for transient failures (5xx, 429, network errors)
- **Optional Caching**: TTL-based caching for static data (services, config, registries) to reduce API calls
- **Access Control**: Flexible tool filtering with read-only mode, whitelist/blacklist, and fine-grained action control

## Comparison with Official Home Assistant MCP Server

Home Assistant provides an [official MCP integration](https://www.home-assistant.io/integrations/mcp_server) starting with version 2025.1. While both enable AI control of Home Assistant, they differ significantly in scope and approach:

### ha-mcp Advantages

- **Power User Focus**: 33 specialized tools for advanced automation, scripts, scenes, helpers, devices, areas, labels, floors, zones, persons, and tags
- **Complete CRUD**: Create, read, update, delete automations/scripts/scenes/helpers (not available in official integration)
- **Deep System Access**: Query registries, analyze dependencies, access logbook, validate config
- **Flexible Output**: Natural language (LLM-optimized) and JSON formats
- **No Setup Required**: Works with any Home Assistant installation, no integration needed

### Official Integration Advantages

- **Simplicity**: Intent-based approach, fewer tools, easier for basic scenarios
- **Built-in**: No external server needed, runs inside Home Assistant
- **Entity-Level Security**: Fine-grained entity exposure control (only whitelisted entities visible)
- **Calendar & Todos**: Dedicated calendar and todo list tools

**Note:** ha-mcp now provides tool-level access control (read-only mode, whitelist/blacklist), complementing the official integration's entity-level controls

### When to Use ha-mcp

Choose ha-mcp if you need:
- Full automation/script/scene lifecycle management (create, edit, delete)
- Advanced analysis (dependencies, cross-references, automation targets)
- System administration (registry queries, config validation)
- Historical data (entity history, logbook, statistics)
- Media management (browser, camera streams)
- Template rendering and Lovelace config access

### When to Use Official Integration

Choose the official integration if you need:
- Simple entity control and status queries
- Calendar and todo list management
- Entity-level security (fine-grained entity exposure control)
- No external infrastructure

**Security note:** Both solutions offer access control - ha-mcp at tool/action level, official integration at entity level. Choose based on your security model.

**Full comparison:** See [docs/feature-comparison.md](docs/feature-comparison.md) for detailed feature matrices and architectural differences.

## Installation

### From Binary

Download the latest release for your platform from the [Releases](../../releases) page.

```bash
# Linux/macOS
tar -xzf ha-mcp_linux_amd64.tar.gz
chmod +x ha-mcp
sudo mv ha-mcp /usr/local/bin/

# Windows
# Extract ha-mcp_windows_amd64.zip and add to PATH
```

### From Source

Requires Go 1.26 or later.

```bash
git clone https://github.com/zorak1103/ha-mcp.git
cd ha-mcp
go build -o ha-mcp ./cmd/ha-mcp
```

### Using Docker

Multi-arch Docker images (amd64/arm64) are published to Docker Hub on each release.

```bash
# Pull the latest image
docker pull zorak1103/ha-mcp:latest

# Run container (token provided by clients via Authorization header)
docker run -d \
  --name ha-mcp \
  -p 8080:8080 \
  -e HA_URL=http://homeassistant.local:8123 \
  zorak1103/ha-mcp:latest

# Or with default token for development (optional)
docker run -d \
  --name ha-mcp \
  -p 8080:8080 \
  -e HA_URL=http://homeassistant.local:8123 \
  -e HA_TOKEN=your-long-lived-access-token \
  zorak1103/ha-mcp:latest

# Use a specific version
docker pull zorak1103/ha-mcp:v0.8.0
```

Available tags:
- `zorak1103/ha-mcp:latest` - Latest release (multi-arch)
- `zorak1103/ha-mcp:vX.Y.Z` - Specific version (multi-arch)

### Linux Packages

RPM and DEB packages are available in the releases:

```bash
# Debian/Ubuntu
sudo dpkg -i ha-mcp_amd64.deb

# RHEL/Fedora
sudo rpm -i ha-mcp_amd64.rpm
```

## Configuration

ha-mcp supports configuration via YAML file, environment variables, or command-line flags.

### Connection Requirements

ha-mcp connects to Home Assistant via **WebSocket** (`ws://{host}/api/websocket` or `wss://{host}/api/websocket` for HTTPS). Ensure:

- Home Assistant is running and accessible
- WebSocket connections are allowed (default in HA)
- The URL points to your Home Assistant instance (HTTP/HTTPS URL is converted to WebSocket internally)
- A valid long-lived access token is configured

### HTTPS/WSS Support

ha-mcp fully supports secure connections. The URL scheme is automatically converted:

| Input URL Scheme  | WebSocket Scheme   |
| ----------------- | ------------------ |
| `http://`         | `ws://`            |
| `https://`        | `wss://`           |
| `ws://`           | `ws://`            |
| `wss://`          | `wss://`           |

**Example configurations for secure connections:**

```yaml
# config.yaml with HTTPS
homeassistant:
  url: "https://homeassistant.example.com"  # Converted to wss://
  token: "your-long-lived-access-token"
```

```bash
# Environment variables with HTTPS
export HA_URL=https://homeassistant.example.com
export HA_TOKEN=your-long-lived-access-token
```

```bash
# Command-line with HTTPS
ha-mcp --ha-url https://homeassistant.example.com --ha-token your-token
```

**Important notes for HTTPS/WSS:**

1. **SSL/TLS Certificates**: The system's certificate store is used for validation. Self-signed certificates may require additional configuration on the host system.

2. **Reverse Proxy Setup**: When using a reverse proxy (nginx, Traefik, Caddy), ensure WebSocket upgrade headers are properly forwarded:
   ```nginx
   # nginx example
   location /api/websocket {
       proxy_pass http://homeassistant:8123;
       proxy_http_version 1.1;
       proxy_set_header Upgrade $http_upgrade;
       proxy_set_header Connection "upgrade";
       proxy_set_header Host $host;
   }
   ```

3. **Home Assistant Cloud (Nabu Casa)**: For remote access via Nabu Casa, use your unique URL:
   ```yaml
   homeassistant:
     url: "https://your-instance.ui.nabu.casa"
     token: "your-long-lived-access-token"
   ```

### Proxy Support

ha-mcp supports HTTP/HTTPS proxies via standard environment variables. The underlying WebSocket library (`coder/websocket`) uses Go's standard HTTP client, which automatically respects these proxy settings.

**Supported environment variables:**

| Variable      | Description                                             |
| ------------- | ------------------------------------------------------- |
| `HTTP_PROXY`  | Proxy for HTTP connections (e.g., `http://proxy:8080`)  |
| `HTTPS_PROXY` | Proxy for HTTPS connections (e.g., `http://proxy:8080`) |
| `NO_PROXY`    | Comma-separated list of hosts to bypass proxy           |

**Example usage:**

```bash
# Set proxy environment variables
export HTTP_PROXY=http://proxy.example.com:8080
export HTTPS_PROXY=http://proxy.example.com:8080
export NO_PROXY=localhost,127.0.0.1

# Start ha-mcp (will use proxy for Home Assistant connection)
ha-mcp --ha-url https://homeassistant.example.com --ha-token your-token
```

**Docker with proxy:**

```bash
docker run -d \
  --name ha-mcp \
  -p 8080:8080 \
  -e HA_URL=https://homeassistant.example.com \
  -e HA_TOKEN=your-token \
  -e HTTP_PROXY=http://proxy.example.com:8080 \
  -e HTTPS_PROXY=http://proxy.example.com:8080 \
  ha-mcp:latest
```

**Notes:**

- Proxy authentication is supported via URL format: `http://user:password@proxy:8080`
- SOCKS5 proxies are supported: `socks5://proxy:1080`
- For WebSocket connections over HTTPS (wss://), the `HTTPS_PROXY` variable is used
- Ensure the proxy allows WebSocket upgrade requests (HTTP 101 Switching Protocols)

### Configuration File

Create a config file at one of these locations:
- `./config.yaml`
- `./configs/config.yaml`
- `$HOME/.config/ha-mcp/config.yaml`
- `/etc/ha-mcp/config.yaml`

```yaml
homeassistant:
  url: "http://homeassistant.local:8123"  # WebSocket URL derived automatically
  token: "your-long-lived-access-token"
  rest:
    rate_limit: 10  # Requests per second (0 = unlimited)
    rate_burst: 5   # Maximum burst size
    max_retries: 3  # Retry attempts for transient failures
    retry_initial_delay_ms: 100  # Initial delay between retries
    retry_max_delay_ms: 5000     # Maximum delay between retries
  websocket:
    max_retries: 3  # Retry attempts for transient failures
    retry_initial_delay_ms: 100
    retry_max_delay_ms: 5000
  cache:
    enabled: false  # Enable caching for static data (opt-in)
    services_ttl_min: 60     # Services cache TTL in minutes
    config_ttl_min: 30       # Config cache TTL in minutes
    entity_reg_ttl_min: 10   # Entity registry cache TTL
    device_reg_ttl_min: 10   # Device registry cache TTL
    area_reg_ttl_min: 30     # Area registry cache TTL

server:
  port: 8080
  read_only: false  # Enable read-only mode (blocks all write operations)
  tool_filter:
    whitelist: []   # If non-empty, ONLY these tools/actions are allowed
    blacklist: []   # Block specific tools/actions (only if whitelist is empty)

logging:
  level: "info"  # debug, info, warn, error
```

### Environment Variables

```bash
export HA_URL=http://homeassistant.local:8123
export HA_TOKEN=your-long-lived-access-token
export HA_MCP_PORT=8080
export HA_MCP_LOG_LEVEL=info

# REST API settings (optional)
export HA_REST_RATE_LIMIT=10   # Requests per second (0 = unlimited, default: 10)
export HA_REST_RATE_BURST=5    # Maximum burst size (default: 5)
export HA_REST_MAX_RETRIES=3   # Max retry attempts (default: 3)
export HA_REST_RETRY_INITIAL_DELAY_MS=100  # Initial retry delay in ms (default: 100)
export HA_REST_RETRY_MAX_DELAY_MS=5000     # Max retry delay in ms (default: 5000)

# WebSocket settings (optional)
export HA_WS_MAX_RETRIES=3     # Max retry attempts (default: 3)
export HA_WS_RETRY_INITIAL_DELAY_MS=100    # Initial retry delay in ms (default: 100)
export HA_WS_RETRY_MAX_DELAY_MS=5000       # Max retry delay in ms (default: 5000)

# Caching settings (optional, disabled by default)
export HA_CACHE_ENABLED=false              # Enable caching (default: false)
export HA_CACHE_SERVICES_TTL_MIN=60        # Services cache TTL (default: 60)
export HA_CACHE_CONFIG_TTL_MIN=30          # Config cache TTL (default: 30)
export HA_CACHE_ENTITY_REG_TTL_MIN=10      # Entity registry TTL (default: 10)
export HA_CACHE_DEVICE_REG_TTL_MIN=10      # Device registry TTL (default: 10)
export HA_CACHE_AREA_REG_TTL_MIN=30        # Area registry TTL (default: 30)

# Access control settings (optional)
export HA_MCP_READ_ONLY=false              # Enable read-only mode (default: false)
export HA_MCP_TOOL_FILTER_WHITELIST=""     # Comma-separated whitelist (default: empty)
export HA_MCP_TOOL_FILTER_BLACKLIST=""     # Comma-separated blacklist (default: empty)
```

### Command-Line Flags

```bash
ha-mcp \
  --ha-url http://homeassistant.local:8123 \
  --ha-token your-long-lived-access-token \
  --port 8080 \
  --read-only  # Optional: enable read-only mode
```

### Getting a Home Assistant Token

1. Open Home Assistant web interface
2. Click on your profile (bottom left)
3. Scroll to "Long-Lived Access Tokens"
4. Click "Create Token"
5. Give it a name (e.g., "ha-mcp")
6. Copy the token (it won't be shown again!)

## Access Control & Tool Filtering

ha-mcp provides flexible access control to restrict which tools and actions are available. This is useful for security-conscious deployments, shared instances, or limiting AI capabilities.

### Read-Only Mode

The simplest form of access control is read-only mode, which blocks all write operations while allowing read operations:

```yaml
# config.yaml
server:
  read_only: true
```

```bash
# Command line
ha-mcp --read-only

# Environment variable
export HA_MCP_READ_ONLY=true
```

**What gets blocked in read-only mode:**
- All `create`, `update`, `delete` actions in `manage_*` tools
- Service calls (`call_service`)
- Helper actions (`helper_action` - toggle, set_value, increment, etc.)
- Script execution, scene activation
- Any operation that modifies Home Assistant state

**What remains available:**
- All `list` and `get` actions
- State queries (`get_state`, `query_entities`, `query_devices`)
- History and statistics
- Analysis tools (`analyze_entity`, `get_entity_dependencies`)
- Registry queries (`get_registry`)
- Logbook access

### Tool Filtering

For more granular control, use the tool filter system with whitelist or blacklist:

#### Whitelist Mode

When whitelist is non-empty, **ONLY** the specified tools/actions are allowed:

```yaml
server:
  tool_filter:
    whitelist:
      - "get_state"                    # Allow entire tool
      - "manage_automation:list"       # Allow specific action
      - "manage_automation:get"
      - "manage_script:list"
      - "query_entities"
```

```bash
# Environment variable (comma-separated)
export HA_MCP_TOOL_FILTER_WHITELIST="get_state,manage_automation:list,manage_automation:get"
```

#### Blacklist Mode

When whitelist is empty, use blacklist to block specific tools/actions:

```yaml
server:
  tool_filter:
    blacklist:
      - "call_service"                 # Block entire tool
      - "manage_automation:delete"     # Block specific action
      - "manage_script:delete"
      - "manage_scene:delete"
```

```bash
# Environment variable (comma-separated)
export HA_MCP_TOOL_FILTER_BLACKLIST="call_service,manage_automation:delete"
```

#### Glob Patterns

Use glob patterns to match multiple tools:

```yaml
server:
  tool_filter:
    blacklist:
      - "manage_*:delete"              # Block delete action across all manage_* tools
      - "manage_*:create"              # Block create action across all manage_* tools
      - "get_*"                        # Block all get_* tools
```

#### Category-Based Filtering

Filter by action category (read/write):

```yaml
server:
  tool_filter:
    blacklist:
      - "*:write"                      # Block all write operations (same as read_only: true)
```

```yaml
server:
  tool_filter:
    whitelist:
      - "*:read"                       # Allow only read operations
```

#### Sub-Action Filtering

For tools with nested actions (like `query_entities` with mode=health, action=remove):

```yaml
server:
  tool_filter:
    blacklist:
      - "query_entities:health:remove"   # Block health check removal
      - "query_devices:health:remove"    # Block device health removal
```

### Filter Behavior

- **Tool Removal**: Completely blocked tools disappear from `tools/list` (the AI won't see them)
- **Schema Modification**: Partially blocked tools have their schemas updated to show only allowed actions
- **Runtime Check**: Attempted blocked actions return an error at runtime
- **Mutual Exclusion**: Whitelist and blacklist cannot be used together - configuration validation will fail if both are non-empty
- **Whitelist Mode**: If whitelist is specified, ONLY listed items are allowed (implicit deny-all)
- **Blacklist Mode**: If whitelist is empty, blacklist blocks specific items (implicit allow-all)
- **AI Notification**: When filter is active, clients receive a "⚠️ SERVER IN RESTRICTED MODE" message

### Example Scenarios

**Scenario 1: Read-only for monitoring**
```yaml
server:
  read_only: true
```

**Scenario 2: Allow only safe operations**
```yaml
server:
  tool_filter:
    whitelist:
      - "get_*"                        # All get tools
      - "query_*"                      # All query tools
      - "analyze_*"                    # All analysis tools
      - "manage_automation:list"
      - "manage_automation:get"
```

**Scenario 3: Block dangerous operations**
```yaml
server:
  tool_filter:
    blacklist:
      - "call_service"                 # Block arbitrary service calls
      - "manage_*:delete"              # Block all delete operations
      - "manage_script:execute"        # Block script execution
      - "manage_scene:activate"        # Block scene activation
```

**Scenario 4: Whitelist with specific write operations allowed**
```yaml
server:
  tool_filter:
    whitelist:
      - "get_*"                        # All get tools
      - "query_*"                      # All query tools
      - "analyze_*"                    # All analysis tools
      - "manage_automation:list"       # Read automation
      - "manage_automation:get"
      - "manage_automation:toggle"     # Exception: allow enable/disable
      - "helper_action"                # Exception: allow helper actions
```

**Note:** You cannot combine whitelist and blacklist - the server will refuse to start with a validation error if both are specified.

## Usage

### Quick Start

```bash
# Initialize configuration files in current directory
ha-mcp init

# Edit the generated config.yaml or .env with your settings
# Then verify your configuration
ha-mcp config

# Start the server
ha-mcp
```

### Available Commands

| Command         | Description                                      |
| --------------- | ------------------------------------------------ |
| `ha-mcp`        | Start the MCP server                             |
| `ha-mcp init`   | Create config.yaml and .env in current directory |
| `ha-mcp config` | Display effective configuration (tokens masked)  |
| `ha-mcp --help` | Show help and available flags                    |

### Starting the Server

```bash
# With config file (default: ./config.yaml)
ha-mcp

# With environment variables
HA_URL=http://homeassistant.local:8123 HA_TOKEN=xxx ha-mcp

# With flags
ha-mcp --ha-url http://homeassistant.local:8123 --ha-token xxx
```

### Using with Cline

Add to your Cline MCP configuration (`~/.config/cline/mcp.json`):

```json
{
  "servers": {
    "ha-mcp": {
      "url": "http://localhost:8080",
      "headers": {
        "Authorization": "Bearer your-ha-access-token"
      },
      "description": "Home Assistant MCP Server"
    }
  }
}
```

### Using with Claude Desktop

Add to Claude Desktop's MCP configuration:

```json
{
  "mcpServers": {
    "homeassistant": {
      "type": "http",
      "url": "http://localhost:8080",
      "headers": {
        "Authorization": "Bearer ${HA_TOKEN}"
      }
    }
  }
}
```

### Using with opencode

Configure in your opencode settings:

```yaml
mcp:
  servers:
    - name: homeassistant
      url: http://localhost:8080
      headers:
        Authorization: "Bearer your-ha-access-token"
```

## API Reference

### MCP Endpoint

All MCP requests are sent to:

```
POST http://localhost:8080/
Content-Type: application/json
Authorization: Bearer <your-ha-access-token>
```

### Available Tools

#### Entity Tools

| Tool                      | Description                                                                                                                                                                                                                              |
| ------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `query_entities`          | Consolidated entity queries (mode: current, history, statistics, domains, presence, health; format: natural/json; group_by: domain, area_id, device_class, integration; health: analyze/remove dead entities with multi-category filter) |
| `query_devices`           | Device health check (mode: health; action: analyze/remove; format: natural/json; categories: disabled, orphaned_config_entry, config_entry_error, no_entities, no_config_entries; manufacturer filter)                                   |
| `get_state`               | Get state of a specific entity (format: natural/json)                                                                                                                                                                                    |
| `analyze_entity`          | Analyze entity usage in automations, scripts, and scenes (format: natural/json)                                                                                                                                                          |
| `get_entity_dependencies` | Find all entities an automation/script depends on (format: natural/json)                                                                                                                                                                 |

#### Registry Tools

| Tool                  | Description                                                                                                   |
| --------------------- | ------------------------------------------------------------------------------------------------------------- |
| `get_registry`        | Query registries (type: entities, devices, areas, all; format: natural/json)                                  |
| `manage_area`         | Consolidated area management (actions: list, get, create, update, delete; format: natural/json for list/get)  |
| `manage_label`        | Consolidated label management (actions: list, get, create, update, delete; format: natural/json for list/get) |
| `manage_floor`        | Consolidated floor management (actions: list, get, create, update, delete; format: natural/json for list/get) |
| `manage_zone`         | Consolidated zone management (actions: list, get, create, update, delete; format: natural/json for list/get)  |
| `manage_person`       | Consolidated person management (actions: list, get, create, update, delete; format: natural/json for list/get)|
| `manage_tag`          | Consolidated tag management (actions: list, get, create, update, delete; format: natural/json for list/get)   |
| `manage_entity`       | Entity registry management (actions: get, update; update fields: name, icon, area_id, disabled_by, hidden_by, labels, aliases; format: natural/json) |
| `manage_device`       | Device registry management (actions: get, update; update fields: name_by_user, area_id, disabled_by, labels; format: natural/json) |
| `manage_config_entry` | Consolidated config entry management (actions: list, get; list: optional domain filter; get: requires entry_id; format: natural/json) |

#### Automation Tools

| Tool                | Description                                                                                                                                   |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `manage_automation` | Consolidated automation management (actions: list, get, create, update, delete, toggle, coverage; format: natural/json for list/get/coverage) |

**Flexible ID Lookup**: The `automation_id` parameter accepts multiple formats:
- Entity ID: `automation.morning_lights`
- Config ID (UUID): `abc123-def456-...`
- Alias or friendly name: `morning lights` (case-insensitive partial match)

#### Helper Tools

ha-mcp provides comprehensive support for all 14 Home Assistant helper types through two consolidated tools.

| Tool            | Description                                                                                                                |
| --------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `manage_helper` | List, create, update, delete, or get details for any helper type (format: natural/json) |
| `helper_action` | Execute runtime actions (toggle, set, increment, start, etc.)                                                              |

##### manage_helper

Universal tool for helper lifecycle management:

| Action        | Description                                                                                          |
| ------------- | ---------------------------------------------------------------------------------------------------- |
| `list`        | List all helpers with optional format (natural/json) and verbose mode                                |
| `create`      | Create a new helper (requires `type`, `id`, `name`)                                                  |
| `update`      | Update an existing helper (requires `entity_id`; supports all helper types)                          |
| `delete`      | Delete an existing helper (requires `entity_id`)                                                     |
| `get_details` | Get detailed configuration for any helper type (requires `entity_id`; format: natural/json) |

**Supported helper types:** `input_boolean`, `input_number`, `input_text`, `input_select`, `input_datetime`, `input_button`, `counter`, `timer`, `schedule`, `group`, `template_sensor`, `template_binary_sensor`, `threshold`, `derivative`, `integral`

**ID Parameter Behavior:**
- For WebSocket helpers (`input_*`, `counter`, `timer`, `schedule`): The `id` parameter controls the entity ID (e.g., `id="test_bool"` creates `input_boolean.test_bool`), while `name` sets the display name
- For Config Entry Flow helpers (`threshold`, `derivative`, `integral`, `group`, `template_*`): Entity ID is derived from `name` (Home Assistant limitation)

##### helper_action

Universal tool for runtime helper operations:

| Action            | Applicable To                                     | Description                     |
| ----------------- | ------------------------------------------------- | ------------------------------- |
| `toggle`          | input_boolean                                     | Toggle on/off                   |
| `set`             | input_number, input_text, input_datetime, counter | Set value                       |
| `increment`       | counter                                           | Increment by step               |
| `decrement`       | counter                                           | Decrement by step               |
| `reset`           | counter, integral                                 | Reset to initial/zero           |
| `start`           | timer                                             | Start timer (optional duration) |
| `pause`           | timer                                             | Pause running timer             |
| `cancel`          | timer                                             | Cancel timer                    |
| `finish`          | timer                                             | Finish immediately              |
| `change`          | timer                                             | Change duration while running   |
| `press`           | input_button                                      | Press/trigger button            |
| `select`          | input_select                                      | Select an option                |
| `set_options`     | input_select                                      | Update available options        |
| `reload`          | schedule, group                                   | Reload from configuration       |
| `add_entities`    | group                                             | Add entities to group           |
| `remove_entities` | group                                             | Remove entities from group      |

#### Script Tools

| Tool            | Description                                                                                                             |
| --------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `manage_script` | Consolidated script management (actions: list, get, create, update, delete, execute; format: natural/json for list/get) |

**Flexible ID Lookup**: The `script_id` parameter accepts multiple formats:
- Entity ID: `script.morning_routine`
- Alias or friendly name: `morning routine` (case-insensitive partial match)

#### Scene Tools

| Tool           | Description                                                                                                             |
| -------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `manage_scene` | Consolidated scene management (actions: list, get, create, update, delete, activate; format: natural/json for list/get) |

**Flexible ID Lookup**: The `scene_id` parameter accepts multiple formats:
- Entity ID: `scene.movie_time`
- Friendly name: `movie time` (case-insensitive partial match)

#### Media Tools

| Tool                | Description                                |
| ------------------- | ------------------------------------------ |
| `browse_media`      | Browse media sources and libraries         |
| `get_camera_stream` | Get camera stream URL for an entity        |
| `sign_media_path`   | Sign a media path for authenticated access |

#### Dashboard Tools

| Tool               | Description                                                                        |
| ------------------ | ---------------------------------------------------------------------------------- |
| `manage_dashboard` | Manage Lovelace dashboards - list, get, create, update, delete, save configuration |

#### Template Tools

| Tool              | Description                                                 |
| ----------------- | ----------------------------------------------------------- |
| `render_template` | Render a Jinja2 template using current Home Assistant state |

#### Logbook Tools

| Tool          | Description                                                                                       |
| ------------- | ------------------------------------------------------------------------------------------------- |
| `get_logbook` | Get logbook entries (mode: entries, correlation; format: natural/json) with cause-effect analysis |

#### Configuration Tools

| Tool              | Description                                                  |
| ----------------- | ------------------------------------------------------------ |
| `validate_config` | Validate Home Assistant configuration.yaml for syntax errors |

#### Target Tools

| Tool             | Description                                                                                                             |
| ---------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `analyze_target` | Analyze targets for automation capabilities (info: triggers, conditions, services, entities, all; format: natural/json) |

#### Service Tools

| Tool            | Description                                                            |
| --------------- | ---------------------------------------------------------------------- |
| `call_service`  | Call any Home Assistant service (format: natural/json)                 |
| `list_services` | List all available services with descriptions (optional domain filter) |

#### System Tools

| Tool              | Description                                                                                |
| ----------------- | ------------------------------------------------------------------------------------------ |
| `get_system_info` | Get Home Assistant system configuration (version, timezone, units, etc.)                   |
| `get_datetime`    | Get current date/time in Home Assistant's configured timezone (optional timezone override) |

#### HACS Tools

> **Note**: HACS (Home Assistant Community Store) is an optional third-party add-on. These tools will only work if HACS is installed.

| Tool          | Description                                                                                                                                                                                                 |
| ------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `manage_hacs` | Manage HACS repositories (actions: info, list, get, releases, release_notes, critical, download, uninstall, add_repository, remove_repository, refresh, toggle_beta; format: natural/json for read actions) |

##### Key Features
- **Repository Management**: List installed repositories, check for updates, download/uninstall
- **Custom Repositories**: Add and remove custom GitHub repositories
- **Beta Versions**: Toggle beta version visibility per repository
- **Release Information**: Get release notes and available versions
- **Filters**: Filter by category (integration, plugin, theme, python_script, appdaemon, netdaemon), installed status, or pending updates

### Output Formats

Most tools support two output formats via the `format` parameter:

- **`natural` (default)**: LLM-optimized natural language output for better readability and reduced token usage
  - Example: `"Living Room Light is on at 80% brightness. Changed 2h ago."`

- **`json`**: Structured JSON output for backward compatibility and programmatic access
  - Example: `{"entity_id": "light.living_room", "state": "on", "attributes": {"brightness": 204, ...}}`

**Tools with format support**: `query_entities`, `query_devices`, `get_state`, `analyze_entity`, `get_entity_dependencies`, `call_service`, `get_registry`, `analyze_target`, `manage_automation`, `manage_script`, `manage_scene`, `manage_area`, `manage_label`, `manage_floor`, `manage_zone`, `manage_person`, `manage_tag`, `manage_helper`, `helper_action`, `manage_hacs`

### Example Requests

#### Query All Entity States

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "query_entities",
    "arguments": {
      "mode": "current"
    }
  }
}
```

#### Get Single Entity State

```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get_state",
    "arguments": {
      "entity_id": "light.living_room"
    }
  }
}
```

#### Query Registry

```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "get_registry",
    "arguments": {
      "type": "entities",
      "domain": "light"
    }
  }
}
```

#### Browse Media

```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "browse_media",
    "arguments": {
      "media_content_id": "media-source://media_source"
    }
  }
}
```

#### Get Statistics

```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "get_statistics",
    "arguments": {
      "statistic_ids": ["sensor.temperature", "sensor.humidity"],
      "period": "hour"
    }
  }
}
```

#### Call a Service

```json
{
  "jsonrpc": "2.0",
  "id": 6,
  "method": "tools/call",
  "params": {
    "name": "call_service",
    "arguments": {
      "domain": "light",
      "service": "turn_on",
      "data": {
        "entity_id": "light.living_room",
        "brightness": 255
      }
    }
  }
}
```

#### Create an Automation

```json
{
  "jsonrpc": "2.0",
  "id": 7,
  "method": "tools/call",
  "params": {
    "name": "create_automation",
    "arguments": {
      "alias": "Turn on lights at sunset",
      "trigger": [
        {
          "platform": "sun",
          "event": "sunset"
        }
      ],
      "action": [
        {
          "service": "light.turn_on",
          "target": {
            "entity_id": "light.living_room"
          }
        }
      ]
    }
  }
}
```

## Authentication

ha-mcp supports flexible authentication via HTTP Bearer tokens. The Home Assistant access token can be provided either per-request via HTTP header or as a server default.

### Token via HTTP Header (Recommended)

MCP clients send the Home Assistant token in the `Authorization` header with every request:

```
Authorization: Bearer <your-long-lived-access-token>
```

This approach is recommended because:
- Each client can use their own Home Assistant token
- Tokens are not stored on the server
- Tokens can have different permissions for different clients

**Example with curl:**

```bash
curl -X POST http://localhost:8080/ \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJ0eXAiOiJKV1QiLCJhbGc..." \
  -d '{"jsonrpc":"2.0","id":1,"method":"tools/list"}'
```

### Development Mode (Optional)

For local development or testing, you can configure a default token on the server:

```bash
ha-mcp --ha-url http://homeassistant.local:8123 --ha-token your-token
```

When a default token is configured:
- Requests **with** an `Authorization` header use the header token
- Requests **without** an `Authorization` header use the default token

This allows backwards-compatible operation while supporting per-request authentication.

### Authentication Errors

When no token is provided and no default is configured:

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "error": {
    "code": -32004,
    "message": "authorization header with Bearer token required"
  }
}
```

## Health Check

The server provides a health check endpoint (no authentication required):

```bash
curl http://localhost:8080/health
# Response: {"status":"ok"}
```

## Troubleshooting

### WebSocket Connection Issues

1. **Verify Home Assistant URL**: Ensure the URL is accessible from where ha-mcp runs
2. **Check Token**: Verify the token is valid and not expired
3. **WebSocket Support**: Ensure Home Assistant allows WebSocket connections (default enabled)
4. **Proxy Configuration**: If using a reverse proxy, ensure WebSocket upgrade is allowed
5. **Firewall**: Ensure port 8123 (HA) and 8080 (MCP) are accessible

### Connection States

ha-mcp includes automatic reconnection with exponential backoff:

- **Initial connection**: Establishes WebSocket and authenticates
- **Disconnection**: Automatic reconnect attempts (1s, 2s, 4s, ... up to 60s)
- **Health monitoring**: Periodic ping to detect connection issues

### Debug Mode

Enable debug logging for more detailed output:

```bash
# Via environment variable
export HA_MCP_LOG_LEVEL=debug
ha-mcp

# Or in config.yaml
# logging:
#   level: "debug"
```

Debug logs show:
- WebSocket connection state changes
- Message IDs and responses
- Reconnection attempts
- Authentication flow

### Common Errors

| Error                      | Solution                                          |
| -------------------------- | ------------------------------------------------- |
| `connection refused`       | Check if Home Assistant is running and accessible |
| `401 unauthorized`         | Token is invalid or expired, create a new one     |
| `websocket: bad handshake` | Check URL format and proxy WebSocket support      |
| `auth_invalid`             | Token authentication failed, verify token         |
| `entity not found`         | Verify the entity_id exists in Home Assistant     |
| `connection closed`        | Network issue, ha-mcp will auto-reconnect         |

## Development

### Prerequisites

- Go 1.26+
- golangci-lint v2
- Docker (for container builds)

### Building

```bash
# Build binary
go build -o ha-mcp ./cmd/ha-mcp

# Run unit tests
go test ./...

# Run linter
golangci-lint run ./...
```

### Integration Tests

Integration tests verify the MCP server against a real Home Assistant instance. They are isolated using a unique `__mcptest_` prefix for all created entities.

**Prerequisites:**
- Running Home Assistant instance (version 2023.1+)
- Long-lived access token with full API access

**Configuration:**
```bash
export HA_INTEGRATION_TEST_URL=http://homeassistant.local:8123
export HA_INTEGRATION_TEST_TOKEN=<your-long-lived-access-token>
export HA_INTEGRATION_TEST_TIMEOUT=5m  # optional
```

**Running integration tests:**
```bash
# Run all integration tests
go test -tags=integration -v ./internal/handlers/integration/...

# Run specific test suite
go test -tags=integration -v ./internal/handlers/integration/... -run TestCounterIntegration
```

**Safety guarantees:**
- All test entities use `__mcptest_` prefix to avoid conflicts
- Pre-test cleanup removes leftover entities from previous runs
- Post-test verification ensures no test data remains
- Tests are skipped automatically if environment variables are not set

See [docs/integration-tests.md](docs/integration-tests.md) for detailed documentation.

### Project Structure

```
ha-mcp/
├── cmd/
│   └── ha-mcp/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration handling
│   ├── homeassistant/
│   │   ├── client.go            # Client interface (~70 methods)
│   │   ├── cached_client.go     # TTL-based caching decorator
│   │   ├── factory.go           # Client factory (creates HybridClient)
│   │   ├── hybrid_client.go     # Hybrid client combining WS + REST
│   │   ├── rest_client.go       # REST client for automation/script/scene CRUD
│   │   ├── retry.go             # Retry logic with exponential backoff
│   │   ├── ws_client.go         # WebSocket connection management
│   │   ├── ws_client_impl.go    # WebSocket Client implementation
│   │   ├── ws_messages.go       # WebSocket message types
│   │   ├── ws_reconnect.go      # Reconnection logic
│   │   └── types.go             # Data types
│   ├── mcp/
│   │   ├── server.go            # MCP HTTP server
│   │   ├── registry.go          # Tool registry
│   │   └── types.go             # MCP protocol types
│   ├── handlers/
│   │   ├── formatter/           # Output formatters (natural/json)
│   │   │   ├── formatter.go     # Formatter interfaces
│   │   │   ├── natural.go       # LLM-optimized natural language output
│   │   │   ├── json.go          # Structured JSON output
│   │   │   ├── registry.go      # Registry formatters
│   │   │   ├── target.go        # Target analysis formatters
│   │   │   ├── automations.go   # Automation list/get formatters
│   │   │   ├── scripts.go       # Script list/get formatters
│   │   │   ├── scenes.go        # Scene list/get formatters
│   │   │   └── helpers.go       # Helper list/get_details formatters
│   │   ├── integration/         # Integration tests (build tag: integration)
│   │   │   ├── helpers.go       # Test ID generation, validation
│   │   │   ├── cleanup.go       # Cleanup utilities
│   │   │   ├── suite_test.go    # Base test suite
│   │   │   └── *_integration_test.go  # Domain-specific tests
│   │   ├── analysis_snapshot.go # Parallel data fetching for analysis
│   │   ├── entities.go          # Entity tool handlers (get_state)
│   │   ├── analysis.go          # Analysis tool handlers (analyze_entity, get_entity_dependencies)
│   │   ├── entities_consolidated.go # Consolidated query_entities tool (current/history/statistics/domains/presence/health)
│   │   ├── entities_presence.go # Presence analysis for query_entities (mode=presence)
│   │   ├── entities_health.go   # Health mode: detect and remove dead entities (mode=health)
│   │   ├── devices_consolidated.go # Consolidated query_devices tool (health)
│   │   ├── devices_health.go    # Device health: detect and remove problematic devices
│   │   ├── automations.go       # Consolidated manage_automation tool
│   │   ├── automations_coverage.go # Automation coverage analysis
│   │   ├── helpers_consolidated.go  # manage_helper and helper_action tools
│   │   ├── scripts.go           # Consolidated manage_script tool
│   │   ├── scenes.go            # Consolidated manage_scene tool
│   │   ├── areas.go             # Consolidated manage_area tool
│   │   ├── labels.go            # Consolidated manage_label tool
│   │   ├── floors.go            # Consolidated manage_floor tool
│   │   ├── zones.go             # Consolidated manage_zone tool
│   │   ├── persons.go           # Consolidated manage_person tool
│   │   ├── tags.go              # Consolidated manage_tag tool
│   │   ├── registry.go          # Registry helper functions
│   │   ├── registry_consolidated.go # Consolidated get_registry tool
│   │   ├── media.go             # Media tool handlers
│   │   ├── lovelace.go          # Lovelace tool handler
│   │   ├── targets_consolidated.go  # Consolidated analyze_target tool
│   │   ├── services.go          # Service discovery handler
│   │   ├── system.go            # System info handler
│   │   ├── datetime.go          # Date/time handler
│   │   ├── templates.go         # Template rendering handler
│   │   ├── logbook.go           # Logbook access handler (entries/correlation modes)
│   │   ├── logbook_correlation.go # Logbook correlation analysis
│   │   ├── config.go            # Configuration validation handler
│   │   ├── hacs.go              # HACS (Community Store) management handler
│   │   └── register.go          # Handler registration
│   └── logging/
│       └── logger.go            # Structured logging
├── configs/
│   ├── config.example.yaml      # Example configuration
│   └── .env.example             # Example environment file
├── docs/
│   └── integration-tests.md     # Integration test documentation
├── Dockerfile                   # Container build
├── .golangci.yml               # Linter configuration
└── README.md                   # This file
```

## Architecture

### Hybrid Communication

ha-mcp uses a hybrid approach combining WebSocket and REST APIs:

- **WebSocket (primary)**: Used for most operations including state queries, service calls, helper management (via `manage_helper`/`helper_action`), and registry access
- **REST API**: Used for automation/script/scene CRUD operations, template rendering, logbook, and config validation

```
┌─────────────┐     HTTP/JSON-RPC      ┌─────────────┐
│  AI Client  │ ◄──────────────────────► │   ha-mcp    │
│  (Claude,   │                         │  MCP Server │
│   Cline)    │                         │             │
└─────────────┘                         └──────┬──────┘
                                               │
                           ┌───────────────────┴───────────────────┐
                           │                                       │
                           │ WebSocket (primary)                   │ REST API
                           │ ws://host/api/websocket               │ http://host/api/...
                           │ - State queries, service calls        │ - Automation CRUD
                           │ - Helper CRUD (manage_helper)         │ - Script CRUD
                           │ - Registry access                     │ - Scene CRUD
                           │                                       │
                           └───────────────┬───────────────────────┘
                                           │
                                    ┌──────▼──────┐
                                    │    Home     │
                                    │  Assistant  │
                                    └─────────────┘
```

### Message Flow

1. AI client sends JSON-RPC request to ha-mcp
2. ha-mcp routes to WebSocket (most operations) or REST API (automation/script/scene CRUD)
3. Home Assistant processes and responds
4. ha-mcp returns result to AI client

### API Limitations

Some Home Assistant operations have limitations:

- **Scripts**: REST API works, but `script.reload` service call is needed after create/update for entity to appear immediately
- **Automations/Scenes**: REST API stores config, but entity may not appear until Home Assistant restart or reload
- **Config Entry helpers**: Template, threshold, derivative, integration, and group helpers use HTTP-based Config Entry Flow (automatically handled by ha-mcp)
- **Entity/Device Registry CREATE operations**: `get_registry` and `query_devices` support UPDATE and DELETE but not CREATE. This is by design—entities and devices are created by integrations (Zigbee, Z-Wave, MQTT, etc.), not manually. Creating registry entries without a backing integration produces orphaned entries that cause errors. UPDATE operations (renaming, area assignment, icons, device class, enable/disable) are safe as they modify metadata only. DELETE operations are essential for cleaning up stale/orphaned entries via `query_entities` health mode and `query_devices` health mode.

## License

GPL-3.0 License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository on GitHub
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Ensure all checks pass:
   ```bash
   # Run linter
   golangci-lint run ./...
   
   # Run tests
   go test -race ./...
   ```
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Pull Request Guidelines

- Ensure CI checks pass (lint, test, security scans)
- Update documentation if needed
- Add tests for new functionality
- Keep commits focused and atomic

## Acknowledgments

- [Model Context Protocol](https://modelcontextprotocol.io/) specification
- [Home Assistant WebSocket API](https://developers.home-assistant.io/docs/api/websocket)
- [coder/websocket](https://github.com/coder/websocket) - Pure Go WebSocket library
