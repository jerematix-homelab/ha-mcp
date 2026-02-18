//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type FloorIntegrationTestSuite struct {
	FloorTestSuite
}

func TestFloorIntegration(t *testing.T) {
	suite.Run(t, new(FloorIntegrationTestSuite))
}

func (s *FloorIntegrationTestSuite) TestFloorLifecycle() {
	floorName := GenerateTestID("floor")
	level1 := 1

	s.RegisterCleanup(func() {
		floors, _ := s.Client().GetFloorRegistry(s.Context())
		for _, floor := range floors {
			if floor.Name == floorName {
				_ = s.Client().DeleteFloor(s.Context(), floor.FloorID)
			}
		}
	})

	// Create floor
	floorConfig := homeassistant.FloorConfig{
		Name:  floorName,
		Icon:  "mdi:home-floor-1",
		Level: &level1,
	}

	created, err := s.Client().CreateFloor(s.Context(), floorConfig)
	s.Require().NoError(err, "Failed to create floor")
	s.Require().NotNil(created)
	s.Equal(floorName, created.Name)
	s.Equal("mdi:home-floor-1", created.Icon)
	s.Require().NotNil(created.Level)
	s.Equal(1, *created.Level)

	floorID := created.FloorID

	// Allow time for registry to update
	time.Sleep(500 * time.Millisecond)

	// Verify floor appears in registry
	floor, err := s.FindFloorByID(floorID)
	s.Require().NoError(err, "Floor should appear in registry")
	s.Equal(floorName, floor.Name)
	s.Equal("mdi:home-floor-1", floor.Icon)
	s.Require().NotNil(floor.Level)
	s.Equal(1, *floor.Level)

	// Update floor (name + icon + level)
	updatedName := GenerateTestID("floor_updated")
	level2 := 2
	updateConfig := homeassistant.FloorConfig{
		Name:  updatedName,
		Icon:  "mdi:home-floor-2",
		Level: &level2,
	}

	updated, err := s.Client().UpdateFloor(s.Context(), floorID, updateConfig)
	s.Require().NoError(err, "Failed to update floor")
	s.Equal(updatedName, updated.Name)
	s.Equal("mdi:home-floor-2", updated.Icon)
	s.Require().NotNil(updated.Level)
	s.Equal(2, *updated.Level)

	time.Sleep(500 * time.Millisecond)

	// Verify update
	floor, err = s.FindFloorByID(floorID)
	s.Require().NoError(err)
	s.Equal(updatedName, floor.Name)
	s.Equal("mdi:home-floor-2", floor.Icon)
	s.Require().NotNil(floor.Level)
	s.Equal(2, *floor.Level)

	// Delete floor
	err = s.Client().DeleteFloor(s.Context(), floorID)
	s.Require().NoError(err, "Failed to delete floor")

	time.Sleep(500 * time.Millisecond)

	// Verify floor is gone
	_, err = s.FindFloorByID(floorID)
	s.Error(err, "Floor should be deleted from registry")
}

func (s *FloorIntegrationTestSuite) TestFloorWithAllFields() {
	floorName := GenerateTestID("floor_full")
	level3 := 3
	aliases := []string{"third_floor", "top_floor"}

	s.RegisterCleanup(func() {
		floors, _ := s.Client().GetFloorRegistry(s.Context())
		for _, floor := range floors {
			if floor.Name == floorName {
				_ = s.Client().DeleteFloor(s.Context(), floor.FloorID)
			}
		}
	})

	// Create floor with all fields
	floorConfig := homeassistant.FloorConfig{
		Name:    floorName,
		Icon:    "mdi:home-floor-3",
		Level:   &level3,
		Aliases: aliases,
	}

	created, err := s.Client().CreateFloor(s.Context(), floorConfig)
	s.Require().NoError(err, "Failed to create floor with all fields")
	s.Require().NotNil(created)

	floorID := created.FloorID

	time.Sleep(500 * time.Millisecond)

	// Verify all fields
	floor, err := s.FindFloorByID(floorID)
	s.Require().NoError(err)
	s.Equal(floorName, floor.Name)
	s.Equal("mdi:home-floor-3", floor.Icon)
	s.Require().NotNil(floor.Level)
	s.Equal(3, *floor.Level)
	s.ElementsMatch(aliases, floor.Aliases)

	// Cleanup
	_ = s.Client().DeleteFloor(s.Context(), floorID)
}

func (s *FloorIntegrationTestSuite) TestFloorUpdatePartial() {
	floorName := GenerateTestID("floor_partial")
	level0 := 0

	s.RegisterCleanup(func() {
		floors, _ := s.Client().GetFloorRegistry(s.Context())
		for _, floor := range floors {
			if floor.Name == floorName {
				_ = s.Client().DeleteFloor(s.Context(), floor.FloorID)
			}
		}
	})

	// Create floor with icon and level
	floorConfig := homeassistant.FloorConfig{
		Name:  floorName,
		Icon:  "mdi:home-floor-0",
		Level: &level0,
	}

	created, err := s.Client().CreateFloor(s.Context(), floorConfig)
	s.Require().NoError(err)

	floorID := created.FloorID

	time.Sleep(500 * time.Millisecond)

	// Update only aliases (icon and level should remain)
	updateConfig := homeassistant.FloorConfig{
		Aliases: []string{"ground_floor", "main_floor"},
	}

	_, err = s.Client().UpdateFloor(s.Context(), floorID, updateConfig)
	s.Require().NoError(err, "Failed to update floor with partial config")

	time.Sleep(500 * time.Millisecond)

	// Verify icon and level unchanged, aliases updated
	floor, err := s.FindFloorByID(floorID)
	s.Require().NoError(err)
	s.Equal("mdi:home-floor-0", floor.Icon, "Icon should remain unchanged")
	s.Require().NotNil(floor.Level, "Level should remain set")
	s.Equal(0, *floor.Level, "Level should remain unchanged")
	s.ElementsMatch([]string{"ground_floor", "main_floor"}, floor.Aliases, "Aliases should be updated")

	// Cleanup
	_ = s.Client().DeleteFloor(s.Context(), floorID)
}

func (s *FloorIntegrationTestSuite) TestFloorAliasUpdate() {
	floorName := GenerateTestID("floor_alias")
	level0 := 0

	s.RegisterCleanup(func() {
		floors, _ := s.Client().GetFloorRegistry(s.Context())
		for _, floor := range floors {
			if floor.Name == floorName {
				_ = s.Client().DeleteFloor(s.Context(), floor.FloorID)
			}
		}
	})

	// Create floor without aliases
	created, err := s.Client().CreateFloor(s.Context(), homeassistant.FloorConfig{
		Name:  floorName,
		Level: &level0,
	})
	s.Require().NoError(err, "failed to create floor")
	floorID := created.FloorID

	time.Sleep(500 * time.Millisecond)

	// Add initial aliases
	_, err = s.Client().UpdateFloor(s.Context(), floorID, homeassistant.FloorConfig{
		Aliases: []string{"ground level", "entry floor"},
	})
	s.Require().NoError(err, "failed to set aliases")

	time.Sleep(500 * time.Millisecond)
	floor, err := s.FindFloorByID(floorID)
	s.Require().NoError(err)
	s.ElementsMatch([]string{"ground level", "entry floor"}, floor.Aliases, "aliases should be set")
	s.Require().NotNil(floor.Level, "level should remain set")
	s.Equal(0, *floor.Level, "level should remain unchanged")

	// Replace aliases with a single new one
	_, err = s.Client().UpdateFloor(s.Context(), floorID, homeassistant.FloorConfig{
		Aliases: []string{"lobby"},
	})
	s.Require().NoError(err, "failed to replace aliases")

	time.Sleep(500 * time.Millisecond)
	floor, err = s.FindFloorByID(floorID)
	s.Require().NoError(err)
	s.ElementsMatch([]string{"lobby"}, floor.Aliases, "aliases should be replaced")
	s.Require().NotNil(floor.Level, "level should remain set")
	s.Equal(0, *floor.Level, "level should remain unchanged after alias update")

	// Cleanup
	_ = s.Client().DeleteFloor(s.Context(), floorID)
}

func (s *FloorIntegrationTestSuite) TestMultipleFloors() {
	floor1Name := GenerateTestID("floor_1")
	floor2Name := GenerateTestID("floor_2")
	level1 := 1
	level2 := 2

	var floor1ID, floor2ID string

	s.RegisterCleanup(func() {
		if floor1ID != "" {
			_ = s.Client().DeleteFloor(s.Context(), floor1ID)
		}
		if floor2ID != "" {
			_ = s.Client().DeleteFloor(s.Context(), floor2ID)
		}
	})

	// Create first floor
	config1 := homeassistant.FloorConfig{
		Name:  floor1Name,
		Icon:  "mdi:numeric-1",
		Level: &level1,
	}

	created1, err := s.Client().CreateFloor(s.Context(), config1)
	s.Require().NoError(err, "Failed to create floor 1")
	floor1ID = created1.FloorID

	// Create second floor
	config2 := homeassistant.FloorConfig{
		Name:  floor2Name,
		Icon:  "mdi:numeric-2",
		Level: &level2,
	}

	created2, err := s.Client().CreateFloor(s.Context(), config2)
	s.Require().NoError(err, "Failed to create floor 2")
	floor2ID = created2.FloorID

	time.Sleep(500 * time.Millisecond)

	// Verify both floors exist in registry
	floor1, err := s.FindFloorByID(floor1ID)
	s.Require().NoError(err, "Floor 1 should exist")
	s.Equal(floor1Name, floor1.Name)
	s.Require().NotNil(floor1.Level)
	s.Equal(1, *floor1.Level)

	floor2, err := s.FindFloorByID(floor2ID)
	s.Require().NoError(err, "Floor 2 should exist")
	s.Equal(floor2Name, floor2.Name)
	s.Require().NotNil(floor2.Level)
	s.Equal(2, *floor2.Level)

	// Delete both floors
	err = s.Client().DeleteFloor(s.Context(), floor1ID)
	s.Require().NoError(err, "Failed to delete floor 1")

	err = s.Client().DeleteFloor(s.Context(), floor2ID)
	s.Require().NoError(err, "Failed to delete floor 2")

	time.Sleep(500 * time.Millisecond)

	// Verify both are gone
	_, err = s.FindFloorByID(floor1ID)
	s.Error(err, "Floor 1 should be deleted")

	_, err = s.FindFloorByID(floor2ID)
	s.Error(err, "Floor 2 should be deleted")
}
