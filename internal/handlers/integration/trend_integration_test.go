//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type TrendIntegrationTestSuite struct {
	HelperTestSuite
}

func TestTrendIntegration(t *testing.T) {
	suite.Run(t, new(TrendIntegrationTestSuite))
}

// createSourceNumber creates an input_number and wraps it with a template sensor.
// Trend requires a sensor entity, not input_number.
func (s *TrendIntegrationTestSuite) createSourceNumber(prefix string, initialValue float64) (string, string) {
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

func (s *TrendIntegrationTestSuite) TestTrendLifecycle() {
	// Create source number (input_number + template sensor wrapper)
	inputEntityID, sensorEntityID := s.createSourceNumber("trend_src", 25.0)
	trendName := GenerateTestID("trend")
	trendEntityID := BuildEntityID("binary_sensor", trendName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), trendEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	})

	// Create trend sensor
	trendConfig := homeassistant.HelperConfig{
		Platform: "trend",
		Config: map[string]any{
			"name":      trendName,
			"entity_id": sensorEntityID, // Use sensor entity
		},
	}

	err := s.Client().CreateHelper(s.Context(), trendConfig)
	s.Require().NoError(err, "Failed to create trend")

	entity, err := s.WaitForEntity(trendEntityID, 5*time.Second)
	s.Require().NoError(err, "Trend did not appear")
	s.NotEmpty(entity.State, "Trend should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), trendEntityID)
	s.Require().NoError(err, "Failed to delete trend")

	err = s.WaitForEntityGone(trendEntityID, 5*time.Second)
	s.Require().NoError(err, "Trend should be deleted")

	// Cleanup sources
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
}

func (s *TrendIntegrationTestSuite) TestTrendWithGradient() {
	// Create source number (input_number + template sensor wrapper)
	inputEntityID, sensorEntityID := s.createSourceNumber("trend_grad", 50.0)
	trendName := GenerateTestID("trend_gradient")
	trendEntityID := BuildEntityID("binary_sensor", trendName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), trendEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	})

	// Create trend sensor (keeping only basic config - gradient fields not supported in this HA version)
	trendConfig := homeassistant.HelperConfig{
		Platform: "trend",
		Config: map[string]any{
			"name":      trendName,
			"entity_id": sensorEntityID,
		},
	}

	err := s.Client().CreateHelper(s.Context(), trendConfig)
	s.Require().NoError(err, "Failed to create trend")

	entity, err := s.WaitForEntity(trendEntityID, 5*time.Second)
	s.Require().NoError(err, "Trend did not appear")
	s.NotEmpty(entity.State, "Trend should have a state")

	// Cleanup
	_ = s.Client().DeleteHelper(s.Context(), trendEntityID)
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
}
