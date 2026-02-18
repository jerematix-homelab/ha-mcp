//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type MinMaxIntegrationTestSuite struct {
	HelperTestSuite
}

func TestMinMaxIntegration(t *testing.T) {
	suite.Run(t, new(MinMaxIntegrationTestSuite))
}

// createSourceNumbers creates multiple input_number entities for min_max testing.
func (s *MinMaxIntegrationTestSuite) createSourceNumbers(prefix string, count int) []string {
	entityIDs := make([]string, count)

	for i := 0; i < count; i++ {
		name := GenerateTestID(prefix + "_num")
		entityID := BuildEntityID("input_number", name)

		config := homeassistant.HelperConfig{
			Platform: "input_number",
			Config: map[string]any{
				"name":    name,
				"min":     0.0,
				"max":     100.0,
				"initial": float64((i + 1) * 10),
			},
		}

		err := s.Client().CreateHelper(s.Context(), config)
		s.Require().NoError(err, "Failed to create input_number %d", i)

		_, err = s.WaitForEntity(entityID, 5*time.Second)
		s.Require().NoError(err, "Input_number %d did not appear", i)

		entityIDs[i] = entityID
	}

	return entityIDs
}

func (s *MinMaxIntegrationTestSuite) TestMinMaxLifecycle() {
	// Create source numbers
	sourceIDs := s.createSourceNumbers("minmax_src", 3)
	minMaxName := GenerateTestID("min_max")
	minMaxEntityID := BuildEntityID("sensor", minMaxName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), minMaxEntityID)
		for _, sourceID := range sourceIDs {
			_ = s.Client().DeleteHelper(s.Context(), sourceID)
		}
	})

	// Create min_max sensor (type is required)
	minMaxConfig := homeassistant.HelperConfig{
		Platform: "min_max",
		Config: map[string]any{
			"name":       minMaxName,
			"entity_ids": sourceIDs,
			"type":       "max", // Required: min, max, mean, etc.
		},
	}

	err := s.Client().CreateHelper(s.Context(), minMaxConfig)
	s.Require().NoError(err, "Failed to create min_max")

	entity, err := s.WaitForEntity(minMaxEntityID, 5*time.Second)
	s.Require().NoError(err, "Min_max did not appear")
	s.NotEmpty(entity.State, "Min_max should have a state")

	// Test delete
	err = s.Client().DeleteHelper(s.Context(), minMaxEntityID)
	s.Require().NoError(err, "Failed to delete min_max")

	err = s.WaitForEntityGone(minMaxEntityID, 5*time.Second)
	s.Require().NoError(err, "Min_max should be deleted")

	// Cleanup sources
	for _, sourceID := range sourceIDs {
		_ = s.Client().DeleteHelper(s.Context(), sourceID)
	}
}

func (s *MinMaxIntegrationTestSuite) TestMinMaxWithType() {
	// Create source numbers
	sourceIDs := s.createSourceNumbers("minmax_mean", 2)
	minMaxName := GenerateTestID("minmax_mean")
	minMaxEntityID := BuildEntityID("sensor", minMaxName)

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteHelper(s.Context(), minMaxEntityID)
		for _, sourceID := range sourceIDs {
			_ = s.Client().DeleteHelper(s.Context(), sourceID)
		}
	})

	// Create min_max sensor with type
	minMaxConfig := homeassistant.HelperConfig{
		Platform: "min_max",
		Config: map[string]any{
			"name":       minMaxName,
			"entity_ids": sourceIDs,
			"type":       "mean",
		},
	}

	err := s.Client().CreateHelper(s.Context(), minMaxConfig)
	s.Require().NoError(err, "Failed to create min_max with type")

	entity, err := s.WaitForEntity(minMaxEntityID, 5*time.Second)
	s.Require().NoError(err, "Min_max did not appear")
	s.NotEmpty(entity.State, "Min_max should have a state")

	// Cleanup
	_ = s.Client().DeleteHelper(s.Context(), minMaxEntityID)
	for _, sourceID := range sourceIDs {
		_ = s.Client().DeleteHelper(s.Context(), sourceID)
	}
}
