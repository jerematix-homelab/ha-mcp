// Package handlers provides MCP tool handlers for Home Assistant operations.
package handlers

import (
	"context"
	"sync"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

// AnalysisSnapshot holds pre-fetched data used during entity analysis.
// By fetching all required data upfront in parallel, we reduce redundant API calls
// and improve analysis performance.
type AnalysisSnapshot struct {
	// AllStates contains all entity states from Home Assistant.
	AllStates []homeassistant.Entity

	// EntityRegistry contains all registered entities.
	EntityRegistry []homeassistant.EntityRegistryEntry

	// DeviceRegistry contains all registered devices.
	DeviceRegistry []homeassistant.DeviceRegistryEntry

	// AreaRegistry contains all registered areas.
	AreaRegistry []homeassistant.AreaRegistryEntry
}

// CreateAnalysisSnapshot fetches all data needed for entity analysis in parallel.
// This optimizes analysis by making 4 parallel API calls instead of multiple sequential calls.
// Returns a snapshot even if some fetches fail - the caller should check individual fields.
//
//nolint:funlen // Parallel fetch structure is clear and readable despite statement count.
func CreateAnalysisSnapshot(ctx context.Context, client homeassistant.Client) *AnalysisSnapshot {
	snapshot := &AnalysisSnapshot{}

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Fetch all states
	wg.Add(1)
	go func() {
		defer wg.Done()
		states, _ := client.GetStates(ctx)
		mu.Lock()
		snapshot.AllStates = states
		mu.Unlock()
	}()

	// Fetch entity registry
	wg.Add(1)
	go func() {
		defer wg.Done()
		entities, _ := client.GetEntityRegistry(ctx)
		mu.Lock()
		snapshot.EntityRegistry = entities
		mu.Unlock()
	}()

	// Fetch device registry
	wg.Add(1)
	go func() {
		defer wg.Done()
		devices, _ := client.GetDeviceRegistry(ctx)
		mu.Lock()
		snapshot.DeviceRegistry = devices
		mu.Unlock()
	}()

	// Fetch area registry
	wg.Add(1)
	go func() {
		defer wg.Done()
		areas, _ := client.GetAreaRegistry(ctx)
		mu.Lock()
		snapshot.AreaRegistry = areas
		mu.Unlock()
	}()

	wg.Wait()
	return snapshot
}

// FindEntityByID finds an entity state by its ID in the snapshot.
// Returns nil if not found.
func (s *AnalysisSnapshot) FindEntityByID(entityID string) *homeassistant.Entity {
	for i := range s.AllStates {
		if s.AllStates[i].EntityID == entityID {
			return &s.AllStates[i]
		}
	}
	return nil
}

// GetEntityArea returns the area_id for an entity, either directly or via its device.
// Returns empty string if the entity is not assigned to any area.
func (s *AnalysisSnapshot) GetEntityArea(entityID string) string {
	// Find entity in registry
	var entityEntry *homeassistant.EntityRegistryEntry
	for i := range s.EntityRegistry {
		if s.EntityRegistry[i].EntityID == entityID {
			entityEntry = &s.EntityRegistry[i]
			break
		}
	}

	if entityEntry == nil {
		return ""
	}

	// If entity has a direct area_id, return it
	if entityEntry.AreaID != "" {
		return entityEntry.AreaID
	}

	// Otherwise, check the device's area
	if entityEntry.DeviceID != "" {
		for i := range s.DeviceRegistry {
			if s.DeviceRegistry[i].ID == entityEntry.DeviceID {
				return s.DeviceRegistry[i].AreaID
			}
		}
	}

	return ""
}

// FindDeviceByID finds a device registry entry by device ID.
// Returns nil if not found.
func (s *AnalysisSnapshot) FindDeviceByID(deviceID string) *homeassistant.DeviceRegistryEntry {
	for i := range s.DeviceRegistry {
		if s.DeviceRegistry[i].ID == deviceID {
			return &s.DeviceRegistry[i]
		}
	}
	return nil
}

// FindEntityRegistryEntry finds an entity registry entry by entity ID.
// Returns nil if not found.
func (s *AnalysisSnapshot) FindEntityRegistryEntry(entityID string) *homeassistant.EntityRegistryEntry {
	for i := range s.EntityRegistry {
		if s.EntityRegistry[i].EntityID == entityID {
			return &s.EntityRegistry[i]
		}
	}
	return nil
}

// FindAreaByID finds an area registry entry by area ID.
// Returns nil if not found.
func (s *AnalysisSnapshot) FindAreaByID(areaID string) *homeassistant.AreaRegistryEntry {
	for i := range s.AreaRegistry {
		if s.AreaRegistry[i].AreaID == areaID {
			return &s.AreaRegistry[i]
		}
	}
	return nil
}
