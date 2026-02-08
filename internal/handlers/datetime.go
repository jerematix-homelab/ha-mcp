package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// DatetimeHandlers provides date/time related tools
type DatetimeHandlers struct {
	nowFunc func() time.Time
}

// NewDatetimeHandlers creates a new DatetimeHandlers instance
func NewDatetimeHandlers() *DatetimeHandlers {
	return &DatetimeHandlers{
		nowFunc: time.Now,
	}
}

// RegisterTools registers all datetime tools with the MCP registry
func (h *DatetimeHandlers) RegisterTools(registry *mcp.Registry) {
	registry.RegisterTool(h.getDatetimeTool(), h.HandleGetDatetime)
}

func (h *DatetimeHandlers) getDatetimeTool() mcp.Tool {
	return mcp.Tool{
		Name:        "get_datetime",
		Description: "Get the current date and time in Home Assistant's configured timezone (or a specified timezone). Returns formatted date, time, timezone info, ISO 8601 timestamp, Unix timestamp, and day-of-week/year info. Useful for time-based automations and providing temporal context to the LLM.",
		InputSchema: mcp.JSONSchema{
			Type: "object",
			Properties: map[string]mcp.JSONSchema{
				"timezone": {
					Type:        "string",
					Description: "Optional timezone override (IANA timezone name, e.g., 'America/New_York', 'Europe/Berlin', 'UTC'). If not provided, uses Home Assistant's configured timezone.",
				},
			},
		},
	}
}

// HandleGetDatetime handles requests to get current date/time.
func (h *DatetimeHandlers) HandleGetDatetime(ctx context.Context, client homeassistant.Client, args map[string]any) (*mcp.ToolsCallResult, error) {
	// Check for optional timezone override
	timezone, _ := GetOptionalString(args, "timezone")

	var tzName string
	var loc *time.Location
	var err error

	if timezone != "" {
		// Use provided timezone
		tzName = timezone
		loc, err = time.LoadLocation(tzName)
		if err != nil {
			return &mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{
					mcp.NewTextContent(fmt.Sprintf("Invalid timezone '%s': %v", tzName, err)),
				},
				IsError: true,
			}, nil
		}
	} else {
		// Get timezone from Home Assistant config
		config, err := client.GetConfig(ctx)
		if err != nil {
			return &mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{
					mcp.NewTextContent(fmt.Sprintf("Failed to get Home Assistant config: %v", err)),
				},
				IsError: true,
			}, nil
		}

		tzName = config.TimeZone
		loc, err = time.LoadLocation(tzName)
		if err != nil {
			return &mcp.ToolsCallResult{
				Content: []mcp.ContentBlock{
					mcp.NewTextContent(fmt.Sprintf("Invalid timezone in Home Assistant config '%s': %v", tzName, err)),
				},
				IsError: true,
			}, nil
		}
	}

	// Get current time in the specified timezone
	now := h.nowFunc().In(loc)

	// Format output
	output := formatDatetimeOutput(now, tzName)

	return &mcp.ToolsCallResult{
		Content: []mcp.ContentBlock{
			mcp.NewTextContent(output),
		},
	}, nil
}

func formatDatetimeOutput(t time.Time, tzName string) string {
	// Get timezone offset
	_, offset := t.Zone()
	offsetHours := offset / 3600
	offsetMins := (offset % 3600) / 60
	offsetStr := fmt.Sprintf("UTC%+03d:%02d", offsetHours, offsetMins)

	// Calculate week number (ISO 8601)
	_, week := t.ISOWeek()

	return fmt.Sprintf(`Current Date and Time Information:

Date: %s
Time: %s
Timezone: %s (%s)

Day of week: %s
Day of year: %d
Week number: %d

ISO 8601: %s
Unix timestamp: %d`,
		t.Format("Monday, January 2, 2006"),
		t.Format("15:04:05"),
		tzName,
		offsetStr,
		t.Weekday().String(),
		t.YearDay(),
		week,
		t.Format(time.RFC3339),
		t.Unix(),
	)
}
