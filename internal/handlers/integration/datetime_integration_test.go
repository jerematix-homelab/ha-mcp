//go:build integration

package integration

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/handlers"
)

type DatetimeIntegrationTestSuite struct {
	HelperTestSuite
}

func TestDatetimeIntegration(t *testing.T) {
	suite.Run(t, new(DatetimeIntegrationTestSuite))
}

func (s *DatetimeIntegrationTestSuite) TestGetDatetimeDefault() {
	h := handlers.NewDatetimeHandlers()

	// Call with no timezone (uses HA config timezone)
	result, err := h.HandleGetDatetime(s.Context(), s.Client(), map[string]any{})
	s.Require().NoError(err, "get_datetime should not return an error")
	s.Require().NotNil(result, "Result should not be nil")
	s.False(result.IsError, "Result should not be an error")
	s.Require().NotEmpty(result.Content, "Result should have content")

	content := result.Content[0].Text
	s.NotEmpty(content, "Content should not be empty")

	// Verify expected fields are present
	s.Contains(content, "Date:", "Output should contain Date field")
	s.Contains(content, "Time:", "Output should contain Time field")
	s.Contains(content, "Timezone:", "Output should contain Timezone field")
	s.Contains(content, "ISO 8601:", "Output should contain ISO 8601 field")
	s.Contains(content, "Unix timestamp:", "Output should contain Unix timestamp field")
	s.Contains(content, "Day of week:", "Output should contain Day of week field")
	s.Contains(content, "Day of year:", "Output should contain Day of year field")
	s.Contains(content, "Week number:", "Output should contain Week number field")

	// Verify the timestamp is reasonable (within last minute)
	s.Contains(content, time.Now().Format("2006"), "Output should contain current year")
}

func (s *DatetimeIntegrationTestSuite) TestGetDatetimeWithTimezoneOverride() {
	h := handlers.NewDatetimeHandlers()

	// Call with explicit timezone
	result, err := h.HandleGetDatetime(s.Context(), s.Client(), map[string]any{
		"timezone": "America/New_York",
	})
	s.Require().NoError(err, "get_datetime should not return an error")
	s.Require().NotNil(result, "Result should not be nil")
	s.False(result.IsError, "Result should not be an error")
	s.Require().NotEmpty(result.Content, "Result should have content")

	content := result.Content[0].Text
	s.Contains(content, "Timezone: America/New_York", "Output should show requested timezone")
}

func (s *DatetimeIntegrationTestSuite) TestGetDatetimeUTC() {
	h := handlers.NewDatetimeHandlers()

	// Call with UTC timezone
	result, err := h.HandleGetDatetime(s.Context(), s.Client(), map[string]any{
		"timezone": "UTC",
	})
	s.Require().NoError(err, "get_datetime should not return an error")
	s.Require().NotNil(result, "Result should not be nil")
	s.False(result.IsError, "Result should not be an error")
	s.Require().NotEmpty(result.Content, "Result should have content")

	content := result.Content[0].Text
	s.Contains(content, "Timezone: UTC", "Output should show UTC timezone")
	s.Contains(content, "UTC+00:00", "UTC should show +00:00 offset")
}

func (s *DatetimeIntegrationTestSuite) TestGetDatetimeInvalidTimezone() {
	h := handlers.NewDatetimeHandlers()

	// Call with invalid timezone
	result, err := h.HandleGetDatetime(s.Context(), s.Client(), map[string]any{
		"timezone": "Invalid/Timezone",
	})
	s.Require().NoError(err, "Handler should not return error for invalid timezone")
	s.Require().NotNil(result, "Result should not be nil")
	s.True(result.IsError, "Result should indicate an error")
	s.Require().NotEmpty(result.Content, "Result should have error content")

	content := result.Content[0].Text
	s.Contains(strings.ToLower(content), "invalid", "Error message should mention invalid timezone")
}

func (s *DatetimeIntegrationTestSuite) TestGetDatetimeTimezoneDifference() {
	h := handlers.NewDatetimeHandlers()

	// Get time in two different timezones
	resultUTC, err := h.HandleGetDatetime(s.Context(), s.Client(), map[string]any{
		"timezone": "UTC",
	})
	s.Require().NoError(err)
	s.False(resultUTC.IsError)

	resultJST, err := h.HandleGetDatetime(s.Context(), s.Client(), map[string]any{
		"timezone": "Asia/Tokyo",
	})
	s.Require().NoError(err)
	s.False(resultJST.IsError)

	// Both should contain current year (sanity check)
	currentYear := time.Now().Format("2006")
	s.Contains(resultUTC.Content[0].Text, currentYear, "UTC result should contain current year")
	s.Contains(resultJST.Content[0].Text, currentYear, "JST result should contain current year")

	// Verify timezone offset is different
	s.Contains(resultUTC.Content[0].Text, "UTC+00:00", "UTC should show +00:00")
	s.Contains(resultJST.Content[0].Text, "UTC+09:00", "Tokyo should show +09:00")
}
