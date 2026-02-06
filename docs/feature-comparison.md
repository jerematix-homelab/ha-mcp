# Feature Comparison: ha-mcp vs. Official Home Assistant MCP Server

## Context

This document compares the features of `ha-mcp` (this project) with the official [Home Assistant MCP Server](https://www.home-assistant.io/integrations/mcp_server) integration.

---

## Architectural Differences

| Aspect | ha-mcp | Official HA MCP Server |
|--------|--------|------------------------|
| **Type** | Standalone Go binary (external server) | HA integration (built-in) |
| **Transport** | HTTP JSON-RPC | Streamable HTTP |
| **HA Communication** | WebSocket + REST API (Hybrid) | Direct Python API (internal) |
| **Tool Design** | 26 specialized tools with granular control | Dynamically generated tools from Assist API (~10 tools) |
| **Authentication** | Long-Lived Access Token | OAuth (IndieAuth) + Long-Lived Token |
| **Entity Access** | All entities (no filtering) | Only explicitly exposed entities (Voice Assistant Exposure) |

---

## Detailed Tool Comparison

### Entity Queries & Control

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| Query entity states | `query_entities` mode=current (Filter, Pagination, natural/json format) | `HassGetState`, `GetLiveContext` (all exposed entities) |
| Query single entity | `get_state` (natural/json format) | `HassGetState` |
| Entity on/off/toggle | `call_service` (any service) | `HassTurnOn`, `HassTurnOff`, `HassTurnToggle` |
| Query temperature | via `get_state` | `HassGetTemperature` (specialized) |
| Query weather | via `get_state` | `HassGetWeather` (specialized) |
| List domains | `query_entities` mode=domains | -- |
| Entity history | `query_entities` mode=history (time range, filter, pagination, natural/json) | -- |
| Entity statistics | `query_entities` mode=statistics (long-term data, pagination, natural/json) | -- |
| Cover control | `call_service` (domain=cover) | `HassOpenCover`, `HassCloseCover` |
| Date/Time | `get_datetime` | `GetDateTime` |

### Automations

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| List automations | `manage_automation` action=list | -- |
| Automation details | `manage_automation` action=get | -- |
| Create automation | `manage_automation` action=create | -- |
| Edit automation | `manage_automation` action=update | -- |
| Delete automation | `manage_automation` action=delete | -- |
| Enable/disable automation | `manage_automation` action=toggle | -- |

### Scripts

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| List scripts | `manage_script` action=list | -- |
| Script details | `manage_script` action=get | -- |
| Create script | `manage_script` action=create | -- |
| Edit script | `manage_script` action=update | -- |
| Delete script | `manage_script` action=delete | -- |
| Execute script | `manage_script` action=execute | `ScriptTool` (exposed scripts as individual tools) |

### Scenes

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| List scenes | `manage_scene` action=list | -- |
| Scene details | `manage_scene` action=get | -- |
| Create scene | `manage_scene` action=create | -- |
| Edit scene | `manage_scene` action=update | -- |
| Delete scene | `manage_scene` action=delete | -- |
| Activate scene | `manage_scene` action=activate | via Intent/Service |

### Helpers (Input Helpers)

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| List helpers | `list_helpers`, `manage_helper` action=list | -- |
| Create helper | `manage_helper` action=create (15 types) | -- |
| Delete helper | `manage_helper` action=delete | -- |
| Helper actions | `helper_action` (toggle, set, increment, etc.) | via `call_service` Intents (limited) |
| Timer management | `helper_action` (start/pause/cancel/finish) | Timer Intents (HassStartTimer, etc.) |

### Registry & System

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| Entity registry | `get_registry` type=entities | -- |
| Device registry | `get_registry` type=devices | -- |
| Area registry | `get_registry` type=areas | Area context in prompts |
| List services | `list_services` | -- |
| System info | `get_system_info` | -- |
| Validate config | `validate_config` | -- |
| Config entries | `list_config_entries`, `get_config_entry` | -- |

### Analysis & Advanced

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| Entity analysis | `analyze_entity` (references in automations/scripts/scenes) | -- |
| Dependency analysis | `get_entity_dependencies` | -- |
| Target analysis | `analyze_target` (triggers/conditions/services) | -- |
| Render Jinja2 templates | `render_template` | -- |
| Logbook | `get_logbook` | -- |
| Statistics (long-term) | `get_statistics` | -- |
| Lovelace dashboard | `get_lovelace_config` | -- |

### Media

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| Browse media | `browse_media` | -- |
| Camera stream URL | `get_camera_stream` | -- |
| Sign media path | `sign_media_path` | -- |

### Calendar & Todo

| Function | ha-mcp | Official HA MCP |
|----------|--------|-----------------|
| Calendar events | -- | `CalendarGetEvents` |
| Todo lists | -- | `TodoGetItems` |

---

## Summary

### ha-mcp Strengths:
- **CRUD Operations**: Complete create/edit/delete for automations, scripts, scenes, and helpers
- **Registry Access**: Detailed access to entity, device, and area registries
- **Analysis**: Entity dependencies, automation targets, cross-references
- **Historical Data**: Entity history, logbook, long-term statistics
- **System Administration**: Config validation, config entries, service listing
- **Media**: Media browser, camera streams, signed URLs
- **Dashboard**: Read Lovelace configuration
- **Templates**: Jinja2 template rendering
- **Flexibility**: `call_service` can invoke *any* HA service
- **Output Formats**: Natural Language (LLM-optimized) and JSON
- **Pagination**: Comprehensive pagination for large datasets

### Official HA MCP Strengths:
- **Calendar & Todos**: Dedicated tools (`CalendarGetEvents`, `TodoGetItems`) - not available in ha-mcp
- **Simplicity**: Fewer tools, intent-based, easier for basic scenarios
- **Security**: Entity exposure control (only whitelisted entities visible)
- **No Infrastructure**: Runs inside HA itself, no external server needed
- **OAuth Support**: Standards-compliant authentication

### Feature Gaps in ha-mcp:
1. **Calendar Events** retrieval (`CalendarGetEvents` equivalent)
2. **Todo Lists** management (`TodoGetItems` equivalent)
3. **Entity Exposure Filter** (security feature)

### Feature Gaps in Official HA MCP:
1. No CRUD for automations/scripts/scenes/helpers
2. No registry queries
3. No history/statistics/logbook
4. No analysis tools (dependencies, targets)
5. No template rendering
6. No media tools
7. No dashboard insights
8. No config validation
9. No pagination
10. No `call_service` for arbitrary services
