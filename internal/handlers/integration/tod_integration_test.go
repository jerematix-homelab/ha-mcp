//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type TodIntegrationTestSuite struct {
	HelperTestSuite
}

func TestTodIntegration(t *testing.T) {
	suite.Run(t, new(TodIntegrationTestSuite))
}

func (s *TodIntegrationTestSuite) TestTodLifecycle() {
	todName := GenerateTestID("tod")
	todEntityID := BuildEntityID("binary_sensor", todName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), todEntityID)
	})

	// Create time of day sensor (6am to 10pm)
	todConfig := homeassistant.HelperConfig{
		Platform: "tod",
		Config: map[string]any{
			"name":        todName,
			"after_time":  "06:00:00",
			"before_time": "22:00:00",
		},
	}

	err := s.Client().CreateHelper(s.Context(), todConfig)
	s.Require().NoError(err, "Failed to create tod")

	entity, err := s.WaitForEntity(todEntityID, 5*time.Second)
	s.Require().NoError(err, "Tod did not appear")
	s.NotEmpty(entity.State, "Tod should have a state (on or off)")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), todEntityID)
	s.Require().NoError(err, "Failed to delete tod")

	err = s.WaitForEntityGone(todEntityID, 5*time.Second)
	s.Require().NoError(err, "Tod should be deleted")
}

func (s *TodIntegrationTestSuite) TestTodWithOffsets() {
	// Skip this test - offset fields are not supported in this Home Assistant version
	s.T().Skip("after_offset and before_offset are not supported in this HA version")
}
