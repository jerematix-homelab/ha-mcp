//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type ZoneIntegrationTestSuite struct {
	ZoneTestSuite
}

func TestZoneIntegration(t *testing.T) {
	suite.Run(t, new(ZoneIntegrationTestSuite))
}

func (s *ZoneIntegrationTestSuite) TestZoneLifecycle() {
	zoneName := GenerateTestID("zone")
	lat := 51.5074
	lon := -0.1278
	rad := 100.0

	s.RegisterCleanup(func() {
		zones, _ := s.Client().GetZones(s.Context())
		for _, zone := range zones {
			if zone.Name == zoneName {
				_ = s.Client().DeleteZone(s.Context(), zone.ID)
			}
		}
	})

	// Create zone
	zoneConfig := homeassistant.ZoneConfig{
		Name:      zoneName,
		Latitude:  &lat,
		Longitude: &lon,
		Radius:    &rad,
		Icon:      "mdi:map-marker",
	}

	created, err := s.Client().CreateZone(s.Context(), zoneConfig)
	s.Require().NoError(err, "Failed to create zone")
	s.Require().NotNil(created)
	s.Equal(zoneName, created.Name)
	s.Equal(51.5074, created.Latitude)
	s.Equal(-0.1278, created.Longitude)
	s.Equal(100.0, created.Radius)
	s.Equal("mdi:map-marker", created.Icon)

	zoneID := created.ID

	// Allow time for registry to update
	time.Sleep(500 * time.Millisecond)

	// Verify zone appears in registry
	zone, err := s.FindZoneByID(zoneID)
	s.Require().NoError(err, "Zone should appear in registry")
	s.Equal(zoneName, zone.Name)
	s.Equal(51.5074, zone.Latitude)
	s.Equal(-0.1278, zone.Longitude)
	s.Equal(100.0, zone.Radius)

	// Update zone (name + coordinates + radius)
	updatedName := GenerateTestID("zone_updated")
	newLat := 48.8566
	newLon := 2.3522
	newRad := 150.0
	updateConfig := homeassistant.ZoneConfig{
		Name:      updatedName,
		Latitude:  &newLat,
		Longitude: &newLon,
		Radius:    &newRad,
		Icon:      "mdi:home-map-marker",
	}

	updated, err := s.Client().UpdateZone(s.Context(), zoneID, updateConfig)
	s.Require().NoError(err, "Failed to update zone")
	s.Equal(updatedName, updated.Name)
	s.Equal(48.8566, updated.Latitude)
	s.Equal(2.3522, updated.Longitude)
	s.Equal(150.0, updated.Radius)
	s.Equal("mdi:home-map-marker", updated.Icon)

	time.Sleep(500 * time.Millisecond)

	// Verify update
	zone, err = s.FindZoneByID(zoneID)
	s.Require().NoError(err)
	s.Equal(updatedName, zone.Name)
	s.Equal(48.8566, zone.Latitude)
	s.Equal(2.3522, zone.Longitude)
	s.Equal(150.0, zone.Radius)

	// Delete zone
	err = s.Client().DeleteZone(s.Context(), zoneID)
	s.Require().NoError(err, "Failed to delete zone")

	time.Sleep(500 * time.Millisecond)

	// Verify zone is gone
	_, err = s.FindZoneByID(zoneID)
	s.Error(err, "Zone should be deleted from registry")
}

func (s *ZoneIntegrationTestSuite) TestZoneWithAllFields() {
	zoneName := GenerateTestID("zone_full")
	lat := 40.7128
	lon := -74.0060
	rad := 200.0
	passive := true

	s.RegisterCleanup(func() {
		zones, _ := s.Client().GetZones(s.Context())
		for _, zone := range zones {
			if zone.Name == zoneName {
				_ = s.Client().DeleteZone(s.Context(), zone.ID)
			}
		}
	})

	// Create zone with all fields including passive flag
	zoneConfig := homeassistant.ZoneConfig{
		Name:      zoneName,
		Latitude:  &lat,
		Longitude: &lon,
		Radius:    &rad,
		Icon:      "mdi:city",
		Passive:   &passive,
	}

	created, err := s.Client().CreateZone(s.Context(), zoneConfig)
	s.Require().NoError(err, "Failed to create zone with all fields")
	s.Require().NotNil(created)

	zoneID := created.ID

	time.Sleep(500 * time.Millisecond)

	// Verify all fields
	zone, err := s.FindZoneByID(zoneID)
	s.Require().NoError(err)
	s.Equal(zoneName, zone.Name)
	s.Equal(40.7128, zone.Latitude)
	s.Equal(-74.0060, zone.Longitude)
	s.Equal(200.0, zone.Radius)
	s.Equal("mdi:city", zone.Icon)
	s.Equal(true, zone.Passive)

	// Cleanup
	_ = s.Client().DeleteZone(s.Context(), zoneID)
}

func (s *ZoneIntegrationTestSuite) TestZoneUpdatePartial() {
	zoneName := GenerateTestID("zone_partial")
	lat := 37.7749
	lon := -122.4194
	rad := 100.0

	s.RegisterCleanup(func() {
		zones, _ := s.Client().GetZones(s.Context())
		for _, zone := range zones {
			if zone.Name == zoneName {
				_ = s.Client().DeleteZone(s.Context(), zone.ID)
			}
		}
	})

	// Create zone with basic fields
	zoneConfig := homeassistant.ZoneConfig{
		Name:      zoneName,
		Latitude:  &lat,
		Longitude: &lon,
		Radius:    &rad,
		Icon:      "mdi:home",
	}

	created, err := s.Client().CreateZone(s.Context(), zoneConfig)
	s.Require().NoError(err)

	zoneID := created.ID

	time.Sleep(500 * time.Millisecond)

	// Update only radius (coordinates and icon should remain)
	newRad := 250.0
	updateConfig := homeassistant.ZoneConfig{
		Radius: &newRad,
	}

	_, err = s.Client().UpdateZone(s.Context(), zoneID, updateConfig)
	s.Require().NoError(err, "Failed to update zone with partial config")

	time.Sleep(500 * time.Millisecond)

	// Verify coordinates and icon unchanged, radius updated
	zone, err := s.FindZoneByID(zoneID)
	s.Require().NoError(err)
	s.Equal(37.7749, zone.Latitude, "Latitude should remain unchanged")
	s.Equal(-122.4194, zone.Longitude, "Longitude should remain unchanged")
	s.Equal("mdi:home", zone.Icon, "Icon should remain unchanged")
	s.Equal(250.0, zone.Radius, "Radius should be updated")

	// Cleanup
	_ = s.Client().DeleteZone(s.Context(), zoneID)
}

func (s *ZoneIntegrationTestSuite) TestMultipleZones() {
	zone1Name := GenerateTestID("zone_1")
	zone2Name := GenerateTestID("zone_2")
	lat1, lon1, rad1 := 34.0522, -118.2437, 100.0
	lat2, lon2, rad2 := 41.8781, -87.6298, 150.0

	var zone1ID, zone2ID string

	s.RegisterCleanup(func() {
		if zone1ID != "" {
			_ = s.Client().DeleteZone(s.Context(), zone1ID)
		}
		if zone2ID != "" {
			_ = s.Client().DeleteZone(s.Context(), zone2ID)
		}
	})

	// Create first zone
	config1 := homeassistant.ZoneConfig{
		Name:      zone1Name,
		Latitude:  &lat1,
		Longitude: &lon1,
		Radius:    &rad1,
		Icon:      "mdi:numeric-1",
	}

	created1, err := s.Client().CreateZone(s.Context(), config1)
	s.Require().NoError(err, "Failed to create zone 1")
	zone1ID = created1.ID

	// Create second zone
	config2 := homeassistant.ZoneConfig{
		Name:      zone2Name,
		Latitude:  &lat2,
		Longitude: &lon2,
		Radius:    &rad2,
		Icon:      "mdi:numeric-2",
	}

	created2, err := s.Client().CreateZone(s.Context(), config2)
	s.Require().NoError(err, "Failed to create zone 2")
	zone2ID = created2.ID

	time.Sleep(500 * time.Millisecond)

	// Verify both zones exist in registry
	zone1, err := s.FindZoneByID(zone1ID)
	s.Require().NoError(err, "Zone 1 should exist")
	s.Equal(zone1Name, zone1.Name)
	s.Equal(34.0522, zone1.Latitude)

	zone2, err := s.FindZoneByID(zone2ID)
	s.Require().NoError(err, "Zone 2 should exist")
	s.Equal(zone2Name, zone2.Name)
	s.Equal(41.8781, zone2.Latitude)

	// Delete both zones
	err = s.Client().DeleteZone(s.Context(), zone1ID)
	s.Require().NoError(err, "Failed to delete zone 1")

	err = s.Client().DeleteZone(s.Context(), zone2ID)
	s.Require().NoError(err, "Failed to delete zone 2")

	time.Sleep(500 * time.Millisecond)

	// Verify both are gone
	_, err = s.FindZoneByID(zone1ID)
	s.Error(err, "Zone 1 should be deleted")

	_, err = s.FindZoneByID(zone2ID)
	s.Error(err, "Zone 2 should be deleted")
}
