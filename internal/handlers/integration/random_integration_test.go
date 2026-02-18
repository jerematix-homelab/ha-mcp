//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type RandomIntegrationTestSuite struct {
	HelperTestSuite
}

func TestRandomIntegration(t *testing.T) {
	suite.Run(t, new(RandomIntegrationTestSuite))
}

func (s *RandomIntegrationTestSuite) TestRandomSensorLifecycle() {
	randomName := GenerateTestID("random_sensor")
	randomEntityID := BuildEntityID("sensor", randomName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), randomEntityID)
	})

	// Create random sensor
	randomConfig := homeassistant.HelperConfig{
		Platform: "random",
		Config: map[string]any{
			"name": randomName,
			"type": "sensor", // Menu selection for random platform
		},
	}

	err := s.Client().CreateHelper(s.Context(), randomConfig)
	s.Require().NoError(err, "Failed to create random_sensor")

	entity, err := s.WaitForEntity(randomEntityID, 5*time.Second)
	s.Require().NoError(err, "Random sensor did not appear")
	s.NotEmpty(entity.State, "Random sensor should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), randomEntityID)
	s.Require().NoError(err, "Failed to delete random_sensor")

	err = s.WaitForEntityGone(randomEntityID, 5*time.Second)
	s.Require().NoError(err, "Random sensor should be deleted")
}

func (s *RandomIntegrationTestSuite) TestRandomBinarySensorLifecycle() {
	randomName := GenerateTestID("random_binary")
	randomEntityID := BuildEntityID("binary_sensor", randomName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), randomEntityID)
	})

	// Create random binary_sensor
	randomConfig := homeassistant.HelperConfig{
		Platform: "random",
		Config: map[string]any{
			"name": randomName,
			"type": "binary_sensor", // Menu selection for random platform
		},
	}

	err := s.Client().CreateHelper(s.Context(), randomConfig)
	s.Require().NoError(err, "Failed to create random_binary_sensor")

	entity, err := s.WaitForEntity(randomEntityID, 5*time.Second)
	s.Require().NoError(err, "Random binary_sensor did not appear")
	s.NotEmpty(entity.State, "Random binary_sensor should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), randomEntityID)
	s.Require().NoError(err, "Failed to delete random_binary_sensor")

	err = s.WaitForEntityGone(randomEntityID, 5*time.Second)
	s.Require().NoError(err, "Random binary_sensor should be deleted")
}

func (s *RandomIntegrationTestSuite) TestRandomSensorWithRange() {
	randomName := GenerateTestID("random_range")
	randomEntityID := BuildEntityID("sensor", randomName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), randomEntityID)
	})

	// Create random sensor with min/max range
	randomConfig := homeassistant.HelperConfig{
		Platform: "random",
		Config: map[string]any{
			"name":    randomName,
			"type":    "sensor",
			"minimum": 10.0,
			"maximum": 50.0,
		},
	}

	err := s.Client().CreateHelper(s.Context(), randomConfig)
	s.Require().NoError(err, "Failed to create random_sensor with range")

	entity, err := s.WaitForEntity(randomEntityID, 5*time.Second)
	s.Require().NoError(err, "Random sensor did not appear")
	s.NotEmpty(entity.State, "Random sensor should have a state")

	// Cleanup
	_ = s.Client().DeleteHelper(s.Context(), randomEntityID)
}
