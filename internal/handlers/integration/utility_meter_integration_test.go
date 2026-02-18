//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type UtilityMeterIntegrationTestSuite struct {
	HelperTestSuite
}

func TestUtilityMeterIntegration(t *testing.T) {
	suite.Run(t, new(UtilityMeterIntegrationTestSuite))
}

// createSourceCounter creates an input_number and a template sensor that wraps it.
// Utility meter requires a sensor entity as source, not input_number.
func (s *UtilityMeterIntegrationTestSuite) createSourceCounter(prefix string, initialValue float64) (string, string) {
	inputName := GenerateTestID(prefix + "_input")
	inputEntityID := BuildEntityID("input_number", inputName)

	inputConfig := homeassistant.HelperConfig{
		Platform: "input_number",
		Config: map[string]any{
			"name":    inputName,
			"min":     0.0,
			"max":     10000.0,
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

func (s *UtilityMeterIntegrationTestSuite) TestUtilityMeterLifecycle() {
	// Create source counter (input_number + template sensor wrapper)
	inputEntityID, sensorEntityID := s.createSourceCounter("umeter_src", 100.0)
	meterName := GenerateTestID("utility_meter")
	meterEntityID := BuildEntityID("sensor", meterName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), meterEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	})

	// Create utility meter with daily cycle (cycle is required)
	meterConfig := homeassistant.HelperConfig{
		Platform: "utility_meter",
		Config: map[string]any{
			"name":   meterName,
			"source": sensorEntityID, // Use sensor entity as source
			"cycle":  "daily",        // Required field
		},
	}

	err := s.Client().CreateHelper(s.Context(), meterConfig)
	s.Require().NoError(err, "Failed to create utility_meter")

	entity, err := s.WaitForEntity(meterEntityID, 5*time.Second)
	s.Require().NoError(err, "Utility meter did not appear")
	s.NotEmpty(entity.State, "Utility meter should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), meterEntityID)
	s.Require().NoError(err, "Failed to delete utility_meter")

	err = s.WaitForEntityGone(meterEntityID, 5*time.Second)
	s.Require().NoError(err, "Utility meter should be deleted")

	// Cleanup sources
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
}

func (s *UtilityMeterIntegrationTestSuite) TestUtilityMeterWithCycle() {
	// Create source counter (input_number + template sensor wrapper)
	inputEntityID, sensorEntityID := s.createSourceCounter("umeter_daily", 50.0)
	meterName := GenerateTestID("meter_daily")
	meterEntityID := BuildEntityID("sensor", meterName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), meterEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	})

	// Create utility meter with monthly cycle
	meterConfig := homeassistant.HelperConfig{
		Platform: "utility_meter",
		Config: map[string]any{
			"name":   meterName,
			"source": sensorEntityID,
			"cycle":  "monthly",
		},
	}

	err := s.Client().CreateHelper(s.Context(), meterConfig)
	s.Require().NoError(err, "Failed to create utility_meter with cycle")

	entity, err := s.WaitForEntity(meterEntityID, 5*time.Second)
	s.Require().NoError(err, "Utility meter did not appear")
	s.NotEmpty(entity.State, "Utility meter should have a state")

	// Cleanup
	_ = s.Client().DeleteHelper(s.Context(), meterEntityID)
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
}
