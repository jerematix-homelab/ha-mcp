//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type StatisticsIntegrationTestSuite struct {
	HelperTestSuite
}

func TestStatisticsIntegration(t *testing.T) {
	suite.Run(t, new(StatisticsIntegrationTestSuite))
}

// createSourceNumber creates an input_number and wraps it with a template sensor.
// Statistics requires a sensor entity, not input_number.
func (s *StatisticsIntegrationTestSuite) createSourceNumber(prefix string, initialValue float64) (string, string) {
	inputName := GenerateTestID(prefix + "_input")
	inputEntityID := BuildEntityID("input_number", inputName)

	inputConfig := homeassistant.HelperConfig{
		Platform: "input_number",
		Config: map[string]any{
			"name":    inputName,
			"min":     0.0,
			"max":     100.0,
			"initial": initialValue,
		},
	}

	err := s.Client().CreateHelper(s.Context(), inputConfig)
	s.Require().NoError(err, "Failed to create input_number")

	_, err = s.WaitForEntity(inputEntityID, 5*time.Second)
	s.Require().NoError(err, "Input_number did not appear")

	// Create template sensor that wraps the input_number
	sensorName := GenerateTestID(prefix + "_sensor")
	sensorEntityID := BuildEntityID("sensor", sensorName)

	templateConfig := homeassistant.HelperConfig{
		Platform: "template",
		Config: map[string]any{
			"name":  sensorName,
			"state": "{{ states('" + inputEntityID + "') | float }}",
		},
	}

	err = s.Client().CreateHelper(s.Context(), templateConfig)
	s.Require().NoError(err, "Failed to create template sensor")

	_, err = s.WaitForEntity(sensorEntityID, 5*time.Second)
	s.Require().NoError(err, "Template sensor did not appear")

	return inputEntityID, sensorEntityID
}

func (s *StatisticsIntegrationTestSuite) TestStatisticsLifecycle() {
	// Create source number (input_number + template sensor wrapper)
	inputEntityID, sensorEntityID := s.createSourceNumber("stats_src", 50.0)
	statsName := GenerateTestID("statistics")
	statsEntityID := BuildEntityID("sensor", statsName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), statsEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	})

	// Create statistics sensor (requires either max_age or sampling_size)
	statsConfig := homeassistant.HelperConfig{
		Platform: "statistics",
		Config: map[string]any{
			"name":          statsName,
			"entity_id":     sensorEntityID, // Use sensor entity
			"sampling_size": 20,             // Required: either this or max_age
		},
	}

	err := s.Client().CreateHelper(s.Context(), statsConfig)
	s.Require().NoError(err, "Failed to create statistics")

	entity, err := s.WaitForEntity(statsEntityID, 5*time.Second)
	s.Require().NoError(err, "Statistics did not appear")
	s.NotEmpty(entity.State, "Statistics should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), statsEntityID)
	s.Require().NoError(err, "Failed to delete statistics")

	err = s.WaitForEntityGone(statsEntityID, 5*time.Second)
	s.Require().NoError(err, "Statistics should be deleted")

	// Cleanup sources
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
}

func (s *StatisticsIntegrationTestSuite) TestStatisticsWithSamplingSize() {
	// Create source number (input_number + template sensor wrapper)
	inputEntityID, sensorEntityID := s.createSourceNumber("stats_samp", 30.0)
	statsName := GenerateTestID("stats_sampling")
	statsEntityID := BuildEntityID("sensor", statsName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), statsEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	})

	// Create statistics sensor (requires either max_age or sampling_size)
	statsConfig := homeassistant.HelperConfig{
		Platform: "statistics",
		Config: map[string]any{
			"name":          statsName,
			"entity_id":     sensorEntityID,
			"sampling_size": 20, // Required: either this or max_age
		},
	}

	err := s.Client().CreateHelper(s.Context(), statsConfig)
	s.Require().NoError(err, "Failed to create statistics")

	entity, err := s.WaitForEntity(statsEntityID, 5*time.Second)
	s.Require().NoError(err, "Statistics did not appear")
	s.NotEmpty(entity.State, "Statistics should have a state")

	// Cleanup
	_ = s.Client().DeleteHelper(s.Context(), statsEntityID)
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
}
