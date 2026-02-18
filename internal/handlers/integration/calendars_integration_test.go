//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CalendarIntegrationTestSuite struct {
	IntegrationTestSuite
}

func TestCalendarIntegration(t *testing.T) {
	suite.Run(t, new(CalendarIntegrationTestSuite))
}

// TestListCalendars verifies that we can list all calendars.
func (s *CalendarIntegrationTestSuite) TestListCalendars() {
	calendars, err := s.Client().GetCalendars(s.Context())
	s.Require().NoError(err, "Failed to list calendars")

	s.T().Logf("Found %d calendar(s)", len(calendars))

	// Verify calendar structure
	for _, cal := range calendars {
		s.NotEmpty(cal.EntityID, "Calendar should have entity_id")
		s.NotEmpty(cal.Name, "Calendar should have name")
		s.T().Logf("  - %s (%s)", cal.Name, cal.EntityID)
	}
}

// TestCalendarEventOperations verifies CRUD operations on calendar events.
// This test requires at least one writable calendar to exist in Home Assistant.
func (s *CalendarIntegrationTestSuite) TestCalendarEventOperations() {
	// Get calendars
	calendars, err := s.Client().GetCalendars(s.Context())
	s.Require().NoError(err, "Failed to list calendars")

	if len(calendars) == 0 {
		s.T().Skip("No calendars found in Home Assistant - create one manually to run this test")
		return
	}

	// Find a writable calendar by testing GetCalendarEvents
	// Read-only calendars (like holiday calendars) may not support event queries
	var calendarEntityID string
	now := time.Now()
	start := now.Format(time.RFC3339)
	end := now.Add(30 * 24 * time.Hour).Format(time.RFC3339)

	for _, cal := range calendars {
		_, err := s.Client().GetCalendarEvents(s.Context(), cal.EntityID, start, end)
		if err == nil {
			calendarEntityID = cal.EntityID
			break
		}
	}

	if calendarEntityID == "" {
		s.T().Skip("No writable calendars found - all calendars are read-only (e.g., holiday calendars)")
		return
	}

	s.T().Logf("Using writable calendar: %s", calendarEntityID)

	// Test 1: Get events (verify API works with the writable calendar)
	events, err := s.Client().GetCalendarEvents(s.Context(), calendarEntityID, start, end)
	s.Require().NoError(err, "Failed to get calendar events")

	initialCount := len(events)
	s.T().Logf("Initial event count: %d", initialCount)

	// Test 2: Create event (datetime format)
	eventSummary := "HA-MCP Integration Test Event"
	eventStart := now.Add(24 * time.Hour).Format(time.RFC3339)
	eventEnd := now.Add(25 * time.Hour).Format(time.RFC3339)

	_, err = s.Client().CallService(s.Context(), "calendar", "create_event", map[string]any{
		"entity_id":       calendarEntityID,
		"summary":         eventSummary,
		"start_date_time": eventStart,
		"end_date_time":   eventEnd,
		"description":     "Created by ha-mcp integration test",
		"location":        "Test Location",
	})
	s.Require().NoError(err, "Failed to create calendar event")

	s.T().Log("Created test event successfully")

	// Test 3: Get events again to verify creation
	time.Sleep(1 * time.Second) // Brief delay for HA to process

	events, err = s.Client().GetCalendarEvents(s.Context(), calendarEntityID, start, end)
	s.Require().NoError(err, "Failed to get events after creation")

	s.Require().Greater(len(events), initialCount, "Should have more events after creation")

	// Find the created event
	var testEventUID string
	for _, event := range events {
		if event.Summary == eventSummary {
			testEventUID = event.UID
			s.T().Logf("Found created event with UID: %s", testEventUID)
			break
		}
	}
	s.Require().NotEmpty(testEventUID, "Should find the created event")

	// Test 4: Delete event (cleanup)
	_, err = s.Client().CallService(s.Context(), "calendar", "delete_event", map[string]any{
		"entity_id": calendarEntityID,
		"uid":       testEventUID,
	})
	s.Require().NoError(err, "Failed to delete calendar event")

	s.T().Log("Deleted test event successfully")

	// Test 5: Verify deletion
	time.Sleep(1 * time.Second) // Brief delay for HA to process

	events, err = s.Client().GetCalendarEvents(s.Context(), calendarEntityID, start, end)
	s.Require().NoError(err, "Failed to get events after deletion")

	// Verify event is gone
	for _, event := range events {
		s.Require().NotEqual(testEventUID, event.UID, "Deleted event should not be in list")
	}

	s.T().Log("Verified event deletion")
}

// TestCalendarAllDayEvent verifies creating all-day events.
func (s *CalendarIntegrationTestSuite) TestCalendarAllDayEvent() {
	// Get calendars
	calendars, err := s.Client().GetCalendars(s.Context())
	s.Require().NoError(err, "Failed to list calendars")

	if len(calendars) == 0 {
		s.T().Skip("No calendars found")
		return
	}

	// Find a writable calendar
	var calendarEntityID string
	now := time.Now()
	start := now.Format(time.RFC3339)
	end := now.Add(30 * 24 * time.Hour).Format(time.RFC3339)

	for _, cal := range calendars {
		_, err := s.Client().GetCalendarEvents(s.Context(), cal.EntityID, start, end)
		if err == nil {
			calendarEntityID = cal.EntityID
			break
		}
	}

	if calendarEntityID == "" {
		s.T().Skip("No writable calendars found")
		return
	}

	// Create all-day event
	tomorrow := time.Now().Add(24 * time.Hour)
	startDate := tomorrow.Format("2006-01-02")
	endDate := tomorrow.Add(24 * time.Hour).Format("2006-01-02")

	_, err = s.Client().CallService(s.Context(), "calendar", "create_event", map[string]any{
		"entity_id":  calendarEntityID,
		"summary":    "HA-MCP All-Day Test Event",
		"start_date": startDate,
		"end_date":   endDate,
	})
	s.Require().NoError(err, "Failed to create all-day event")

	s.T().Log("Created all-day event successfully")

	// Cleanup: Get events to find UID, then delete (reuse start/end from above)
	events, err := s.Client().GetCalendarEvents(s.Context(), calendarEntityID, start, end)
	s.Require().NoError(err, "Failed to get events for cleanup")

	for _, event := range events {
		if event.Summary == "HA-MCP All-Day Test Event" {
			_, _ = s.Client().CallService(s.Context(), "calendar", "delete_event", map[string]any{
				"entity_id": calendarEntityID,
				"uid":       event.UID,
			})
			s.T().Logf("Cleaned up all-day event (UID: %s)", event.UID)
			break
		}
	}
}
