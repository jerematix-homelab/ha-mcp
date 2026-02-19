> [README](../README.md) | [Configuration](configuration.md) | [Tools](tools.md) | [Access Control](access-control.md) | [Architecture](architecture.md) | [Troubleshooting](troubleshooting.md) | [Feature Comparison](feature-comparison.md) | [Integration Tests](integration-tests.md)

# Troubleshooting

## WebSocket Connection Issues

1. **Verify Home Assistant URL**: Ensure the URL is accessible from where ha-mcp runs
2. **Check Token**: Verify the token is valid and not expired
3. **WebSocket Support**: Ensure Home Assistant allows WebSocket connections (default enabled)
4. **Proxy Configuration**: If using a reverse proxy, ensure WebSocket upgrade is allowed
5. **Firewall**: Ensure port 8123 (HA) and 8080 (MCP) are accessible

## Connection States

ha-mcp includes automatic reconnection with exponential backoff:

- **Initial connection**: Establishes WebSocket and authenticates
- **Disconnection**: Automatic reconnect attempts (1s, 2s, 4s, ... up to 60s)
- **Health monitoring**: Periodic ping to detect connection issues

## Debug Mode

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

## Common Errors

| Error                      | Solution                                          |
| -------------------------- | ------------------------------------------------- |
| `connection refused`       | Check if Home Assistant is running and accessible |
| `401 unauthorized`         | Token is invalid or expired, create a new one     |
| `websocket: bad handshake` | Check URL format and proxy WebSocket support      |
| `auth_invalid`             | Token authentication failed, verify token         |
| `entity not found`         | Verify the entity_id exists in Home Assistant     |
| `connection closed`        | Network issue, ha-mcp will auto-reconnect         |
