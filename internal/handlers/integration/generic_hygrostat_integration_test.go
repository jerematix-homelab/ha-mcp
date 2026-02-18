//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type GenericHygrostatIntegrationTestSuite struct {
	HelperTestSuite
}

func TestGenericHygrostatIntegration(t *testing.T) {
	suite.Run(t, new(GenericHygrostatIntegrationTestSuite))
}

// createHygrostatSources creates template switch (humidifier) and template sensor (humidity sensor).
// generic_hygrostat requires switch entity (not input_boolean) and sensor entity (not input_number).
func (s *GenericHygrostatIntegrationTestSuite) createHygrostatSources(prefix string, initialHumidity float64) (string, string, string, string) {
	// Create input_boolean for humidifier base
	boolName := GenerateTestID(prefix + "_bool")
	boolEntityID := BuildEntityID("input_boolean", boolName)

	boolConfig := homeassistant.HelperConfig{
		Platform: "input_boolean",
		Config: map[string]any{
			"name":    boolName,
			"initial": false,
		},
	}

	err := s.Client().CreateHelper(s.Context(), boolConfig)
	s.Require().NoError(err, "Failed to create input_boolean")

	_, err = s.WaitForEntity(boolEntityID, 5*time.Second)
	s.Require().NoError(err, "Input_boolean did not appear")

	// Create template switch that wraps the input_boolean
	humidifierName := GenerateTestID(prefix + "_humidifier")
	humidifierEntityID := BuildEntityID("switch", humidifierName)

	switchConfig := homeassistant.HelperConfig{
		Platform: "template",
		Config: map[string]any{
			"name":     humidifierName,
			"turn_on":  map[string]any{"service": "input_boolean.turn_on", "data": map[string]any{"entity_id": boolEntityID}},
			"turn_off": map[string]any{"service": "input_boolean.turn_off", "data": map[string]any{"entity_id": boolEntityID}},
			"type":     "switch", // Menu selection
		},
	}

	err = s.Client().CreateHelper(s.Context(), switchConfig)
	s.Require().NoError(err, "Failed to create template switch")

	_, err = s.WaitForEntity(humidifierEntityID, 5*time.Second)
	s.Require().NoError(err, "Template switch did not appear")

	// Create input_number for humidity base
	inputName := GenerateTestID(prefix + "_input")
	inputEntityID := BuildEntityID("input_number", inputName)

	inputConfig := homeassistant.HelperConfig{
		Platform: "input_number",
		Config: map[string]any{
			"name":    inputName,
			"min":     0.0,
			"max":     100.0,
			"initial": initialHumidity,
		},
	}

	err = s.Client().CreateHelper(s.Context(), inputConfig)
	s.Require().NoError(err, "Failed to create input_number")

	_, err = s.WaitForEntity(inputEntityID, 5*time.Second)
	s.Require().NoError(err, "Input_number did not appear")

	// Create template sensor that wraps the input_number
	sensorName := GenerateTestID(prefix + "_sensor")
	sensorEntityID := BuildEntityID("sensor", sensorName)

	templateSensorConfig := homeassistant.HelperConfig{
		Platform: "template",
		Config: map[string]any{
			"name":  sensorName,
			"state": "{{ states('" + inputEntityID + "') | float }}",
		},
	}

	err = s.Client().CreateHelper(s.Context(), templateSensorConfig)
	s.Require().NoError(err, "Failed to create template sensor")

	_, err = s.WaitForEntity(sensorEntityID, 5*time.Second)
	s.Require().NoError(err, "Template sensor did not appear")

	return boolEntityID, humidifierEntityID, inputEntityID, sensorEntityID
}

func (s *GenericHygrostatIntegrationTestSuite) TestGenericHygrostatLifecycle() {
	// Create source entities (input_boolean + template switch, input_number + template sensor)
	boolEntityID, humidifierEntityID, inputEntityID, sensorEntityID := s.createHygrostatSources("hygro", 50.0)
	hygroName := GenerateTestID("hygrostat")
	hygroEntityID := BuildEntityID("humidifier", hygroName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), hygroEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
		_ = s.Client().DeleteHelper(s.Context(), humidifierEntityID)
		_ = s.Client().DeleteHelper(s.Context(), boolEntityID)
	})

	// Create generic hygrostat (using API field names)
	hygroConfig := homeassistant.HelperConfig{
		Platform: "generic_hygrostat",
		Config: map[string]any{
			"name":          hygroName,
			"humidifier":    humidifierEntityID, // API field name
			"target_sensor": sensorEntityID,     // API field name
			"device_class":  "humidifier",       // Required field
		},
	}

	err := s.Client().CreateHelper(s.Context(), hygroConfig)
	s.Require().NoError(err, "Failed to create generic_hygrostat")

	entity, err := s.WaitForEntity(hygroEntityID, 5*time.Second)
	s.Require().NoError(err, "Generic hygrostat did not appear")
	s.NotEmpty(entity.State, "Hygrostat should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), hygroEntityID)
	s.Require().NoError(err, "Failed to delete generic_hygrostat")

	err = s.WaitForEntityGone(hygroEntityID, 5*time.Second)
	s.Require().NoError(err, "Generic hygrostat should be deleted")

	// Cleanup sources
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	_ = s.Client().DeleteHelper(s.Context(), humidifierEntityID)
	_ = s.Client().DeleteHelper(s.Context(), boolEntityID)
}

func (s *GenericHygrostatIntegrationTestSuite) TestGenericHygrostatWithTolerances() {
	// Create source entities (input_boolean + template switch, input_number + template sensor)
	boolEntityID, humidifierEntityID, inputEntityID, sensorEntityID := s.createHygrostatSources("hygro_tol", 45.0)
	hygroName := GenerateTestID("hygro_tolerance")
	hygroEntityID := BuildEntityID("humidifier", hygroName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), hygroEntityID)
		_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
		_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
		_ = s.Client().DeleteHelper(s.Context(), humidifierEntityID)
		_ = s.Client().DeleteHelper(s.Context(), boolEntityID)
	})

	// Create generic hygrostat with tolerances (using API field names)
	hygroConfig := homeassistant.HelperConfig{
		Platform: "generic_hygrostat",
		Config: map[string]any{
			"name":          hygroName,
			"humidifier":    humidifierEntityID, // API field name
			"target_sensor": sensorEntityID,     // API field name
			"device_class":  "humidifier",       // Required field
			"dry_tolerance": 2.0,
			"wet_tolerance": 2.0,
		},
	}

	err := s.Client().CreateHelper(s.Context(), hygroConfig)
	s.Require().NoError(err, "Failed to create generic_hygrostat with tolerances")

	entity, err := s.WaitForEntity(hygroEntityID, 5*time.Second)
	s.Require().NoError(err, "Generic hygrostat did not appear")
	s.NotEmpty(entity.State, "Hygrostat should have a state")

	// Cleanup
	_ = s.Client().DeleteHelper(s.Context(), hygroEntityID)
	_ = s.Client().DeleteHelper(s.Context(), sensorEntityID)
	_ = s.Client().DeleteHelper(s.Context(), inputEntityID)
	_ = s.Client().DeleteHelper(s.Context(), humidifierEntityID)
	_ = s.Client().DeleteHelper(s.Context(), boolEntityID)
}
