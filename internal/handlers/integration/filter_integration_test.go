//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type FilterIntegrationTestSuite struct {
	HelperTestSuite
}

func TestFilterIntegration(t *testing.T) {
	suite.Run(t, new(FilterIntegrationTestSuite))
}

// createSourceNumber creates an input_number and wraps it with a template sensor.
// Filter requires a sensor entity, not input_number.
func (s *FilterIntegrationTestSuite) createSourceNumber(prefix string, initialValue float64) (string, string) {
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

func (s *FilterIntegrationTestSuite) TestFilterLifecycle() {
	// Create source number (input_number + template sensor wrapper)
	inputEntityID, sensorEntityID := s.createSourceNumber("filter_src", 42.0)
	filterName := GenerateTestID("filter")
	filterEntityID := BuildEntityID("sensor", filterName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), filterEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	})

	// Create filter sensor with required filter field
	filterConfig := homeassistant.HelperConfig{
		Platform: "filter",
		Config: map[string]any{
			"name":      filterName,
			"entity_id": sensorEntityID, // Use sensor entity
			"filter":    "outlier",      // Required field - filter type
		},
	}

	err := s.Client().CreateHelper(s.Context(), filterConfig)
	s.Require().NoError(err, "Failed to create filter")

	entity, err := s.WaitForEntity(filterEntityID, 5*time.Second)
	s.Require().NoError(err, "Filter did not appear")
	s.NotEmpty(entity.State, "Filter should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), filterEntityID)
	s.Require().NoError(err, "Failed to delete filter")

	err = s.WaitForEntityGone(filterEntityID, 5*time.Second)
	s.Require().NoError(err, "Filter should be deleted")

	// Cleanup sources
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
}
