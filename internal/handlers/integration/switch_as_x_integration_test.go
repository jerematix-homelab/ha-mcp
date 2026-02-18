//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type SwitchAsXIntegrationTestSuite struct {
	HelperTestSuite
}

func TestSwitchAsXIntegration(t *testing.T) {
	suite.Run(t, new(SwitchAsXIntegrationTestSuite))
}

// createSourceSwitch creates an input_boolean and wraps it as a switch via template.
// switch_as_x requires a switch entity, not input_boolean.
func (s *SwitchAsXIntegrationTestSuite) createSourceSwitch(prefix string) (string, string) {
	// Create input_boolean as base
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
	switchName := GenerateTestID(prefix + "_switch")
	switchEntityID := BuildEntityID("switch", switchName)

	// Template switch config (minimal config for switch type)
	templateConfig := homeassistant.HelperConfig{
		Platform: "template",
		Config: map[string]any{
			"name":     switchName,
			"turn_on":  map[string]any{"service": "input_boolean.turn_on", "data": map[string]any{"entity_id": boolEntityID}},
			"turn_off": map[string]any{"service": "input_boolean.turn_off", "data": map[string]any{"entity_id": boolEntityID}},
			"type":     "switch", // Menu selection for template platform
		},
	}

	err = s.Client().CreateHelper(s.Context(), templateConfig)
	s.Require().NoError(err, "Failed to create template switch")

	_, err = s.WaitForEntity(switchEntityID, 5*time.Second)
	s.Require().NoError(err, "Template switch did not appear")

	return boolEntityID, switchEntityID
}

func (s *SwitchAsXIntegrationTestSuite) TestSwitchAsXLight() {
	// Create source switch (input_boolean + template switch wrapper)
	boolEntityID, switchEntityID := s.createSourceSwitch("swx_light_src")
	switchAsXName := GenerateTestID("switch_as_light")
	switchAsXEntityID := BuildEntityID("light", switchAsXName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), switchAsXEntityID)
		_ = s.Client().DeleteHelper(s.Context(), switchEntityID)
		_ = s.Client().DeleteHelper(s.Context(), boolEntityID)
	})

	// Create switch_as_x as light
	switchAsXConfig := homeassistant.HelperConfig{
		Platform: "switch_as_x",
		Config: map[string]any{
			"name":          switchAsXName,
			"entity_id":     switchEntityID, // Use the switch entity
			"target_domain": "light",
		},
	}

	err := s.Client().CreateHelper(s.Context(), switchAsXConfig)
	s.Require().NoError(err, "Failed to create switch_as_x as light")

	entity, err := s.WaitForEntity(switchAsXEntityID, 5*time.Second)
	s.Require().NoError(err, "Switch_as_x light did not appear")
	s.NotEmpty(entity.State, "Switch_as_x should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), switchAsXEntityID)
	s.Require().NoError(err, "Failed to delete switch_as_x")

	err = s.WaitForEntityGone(switchAsXEntityID, 5*time.Second)
	s.Require().NoError(err, "Switch_as_x should be deleted")

	// Cleanup sources
	_ = s.Client().DeleteHelper(s.Context(), switchEntityID)
	_ = s.Client().DeleteHelper(s.Context(), boolEntityID)
}

func (s *SwitchAsXIntegrationTestSuite) TestSwitchAsXCover() {
	// Create source switch (input_boolean + template switch wrapper)
	boolEntityID, switchEntityID := s.createSourceSwitch("swx_cover_src")
	switchAsXName := GenerateTestID("switch_as_cover")
	switchAsXEntityID := BuildEntityID("cover", switchAsXName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), switchAsXEntityID)
		_ = s.Client().DeleteHelper(s.Context(), switchEntityID)
		_ = s.Client().DeleteHelper(s.Context(), boolEntityID)
	})

	// Create switch_as_x as cover
	switchAsXConfig := homeassistant.HelperConfig{
		Platform: "switch_as_x",
		Config: map[string]any{
			"name":          switchAsXName,
			"entity_id":     switchEntityID,
			"target_domain": "cover",
		},
	}

	err := s.Client().CreateHelper(s.Context(), switchAsXConfig)
	s.Require().NoError(err, "Failed to create switch_as_x as cover")

	entity, err := s.WaitForEntity(switchAsXEntityID, 5*time.Second)
	s.Require().NoError(err, "Switch_as_x cover did not appear")
	s.NotEmpty(entity.State, "Switch_as_x should have a state")

	// Cleanup
	_ = s.Client().DeleteHelper(s.Context(), switchAsXEntityID)
	_ = s.Client().DeleteHelper(s.Context(), switchEntityID)
	_ = s.Client().DeleteHelper(s.Context(), boolEntityID)
}
