package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zorak1103/ha-mcp/internal/homeassistant"
	"github.com/zorak1103/ha-mcp/internal/mcp"
)

// mockSceneClient implements homeassistant.Client for testing.
type mockSceneClient struct {
	homeassistant.Client
	listScenesFn  func(ctx context.Context) ([]homeassistant.Entity, error)
	createSceneFn func(ctx context.Context, sceneID string, config homeassistant.SceneConfig) error
	updateSceneFn func(ctx context.Context, sceneID string, config homeassistant.SceneConfig) error
	deleteSceneFn func(ctx context.Context, sceneID string) error
	callServiceFn func(ctx context.Context, domain, service string, data map[string]any) ([]homeassistant.Entity, error)
	getStateFn    func(ctx context.Context, entityID string) (*homeassistant.Entity, error)

	// Track IDs passed to methods for verification
	lastUpdateSceneID string
	lastDeleteSceneID string
	lastGetStateID    string
}

func (m *mockSceneClient) ListScenes(ctx context.Context) ([]homeassistant.Entity, error) {
	if m.listScenesFn != nil {
		return m.listScenesFn(ctx)
	}
	return []homeassistant.Entity{}, nil
}

func (m *mockSceneClient) CreateScene(ctx context.Context, sceneID string, config homeassistant.SceneConfig) error {
	if m.createSceneFn != nil {
		return m.createSceneFn(ctx, sceneID, config)
	}
	return nil
}

func (m *mockSceneClient) UpdateScene(ctx context.Context, sceneID string, config homeassistant.SceneConfig) error {
	m.lastUpdateSceneID = sceneID
	if m.updateSceneFn != nil {
		return m.updateSceneFn(ctx, sceneID, config)
	}
	return nil
}

func (m *mockSceneClient) DeleteScene(ctx context.Context, sceneID string) error {
	m.lastDeleteSceneID = sceneID
	if m.deleteSceneFn != nil {
		return m.deleteSceneFn(ctx, sceneID)
	}
	return nil
}

func (m *mockSceneClient) CallService(ctx context.Context, domain, service string, data map[string]any) ([]homeassistant.Entity, error) {
	if m.callServiceFn != nil {
		return m.callServiceFn(ctx, domain, service, data)
	}
	return nil, nil
}

func (m *mockSceneClient) GetState(ctx context.Context, entityID string) (*homeassistant.Entity, error) {
	m.lastGetStateID = entityID
	if m.getStateFn != nil {
		return m.getStateFn(ctx, entityID)
	}
	return &homeassistant.Entity{
		EntityID:   entityID,
		State:      "scening",
		Attributes: map[string]any{"friendly_name": "Test Scene"},
	}, nil
}

func TestNewSceneHandlers(t *testing.T) {
	t.Parallel()

	h := NewSceneHandlers()
	if h == nil {
		t.Error("NewSceneHandlers() returned nil")
	}
}

func TestSceneHandlers_RegisterTools(t *testing.T) {
	t.Parallel()

	h := NewSceneHandlers()
	registry := mcp.NewRegistry()

	h.RegisterTools(registry)

	tools := registry.ListTools()
	const expectedToolCount = 1 // manage_scene
	if len(tools) != expectedToolCount {
		t.Errorf("RegisterTools() registered %d tools, want %d", len(tools), expectedToolCount)
	}

	if tools[0].Name != "manage_scene" {
		t.Errorf("Expected tool name 'manage_scene', got %q", tools[0].Name)
	}
}

func TestSceneHandlers_ManageScene_List(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          map[string]any
		listScenesErr error
		listScenes    []homeassistant.Entity
		wantError     bool
		wantContains  string
	}{
		{
			name:       "success empty",
			args:       map[string]any{"action": "list"},
			listScenes: []homeassistant.Entity{},
			wantError:  false,
		},
		{
			name: "success with scenes (json)",
			args: map[string]any{"action": "list", "format": "json"},
			listScenes: []homeassistant.Entity{
				{
					EntityID: "scene.movie_time",
					State:    "scening",
					Attributes: map[string]any{
						"friendly_name": "Movie Time",
						"entity_id":     []any{"light.living_room", "media_player.tv"},
					},
				},
				{
					EntityID:   "scene.night_mode",
					State:      "scening",
					Attributes: map[string]any{"friendly_name": "Night Mode"},
				},
			},
			wantError:    false,
			wantContains: "movie_time",
		},
		{
			name: "success with scenes (natural)",
			args: map[string]any{"action": "list"},
			listScenes: []homeassistant.Entity{
				{
					EntityID: "scene.movie_time",
					State:    "scening",
					Attributes: map[string]any{
						"friendly_name": "Movie Time",
						"entity_id":     []any{"light.living_room", "media_player.tv"},
					},
				},
			},
			wantError:    false,
			wantContains: "Movie Time",
		},
		{
			name: "success with name filter (json)",
			args: map[string]any{
				"action":        "list",
				"name_contains": "movie",
				"format":        "json",
			},
			listScenes: []homeassistant.Entity{
				{
					EntityID:   "scene.movie_time",
					State:      "scening",
					Attributes: map[string]any{"friendly_name": "Movie Time"},
				},
				{
					EntityID:   "scene.night_mode",
					State:      "scening",
					Attributes: map[string]any{"friendly_name": "Night Mode"},
				},
			},
			wantError:    false,
			wantContains: "movie_time",
		},
		{
			name: "success with entity filter (json)",
			args: map[string]any{
				"action":          "list",
				"entity_contains": "light",
				"format":          "json",
			},
			listScenes: []homeassistant.Entity{
				{
					EntityID: "scene.movie_time",
					State:    "scening",
					Attributes: map[string]any{
						"friendly_name": "Movie Time",
						"entity_id":     []any{"light.living_room"},
					},
				},
				{
					EntityID:   "scene.night_mode",
					State:      "scening",
					Attributes: map[string]any{"friendly_name": "Night Mode"},
				},
			},
			wantError:    false,
			wantContains: "movie_time",
		},
		{
			name:          "client error",
			args:          map[string]any{"action": "list"},
			listScenesErr: errors.New("connection failed"),
			wantError:     true,
			wantContains:  "Error listing scenes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockSceneClient{
				listScenesFn: func(_ context.Context) ([]homeassistant.Entity, error) {
					if tt.listScenesErr != nil {
						return nil, tt.listScenesErr
					}
					return tt.listScenes, nil
				},
			}

			h := NewSceneHandlers()
			result, err := h.handleManageScene(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("handleManageScene() returned error: %v", err)
				return
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			if len(result.Content) == 0 {
				t.Error("Content is empty")
				return
			}

			content := result.Content[0].Text
			if tt.wantContains != "" && !strings.Contains(content, tt.wantContains) {
				t.Errorf("Content = %q, want to contain %q", content, tt.wantContains)
			}
		})
	}
}

func TestSceneHandlers_ManageScene_Get(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         map[string]any
		getStateErr  error
		wantError    bool
		wantContains string
	}{
		{
			name: "success",
			args: map[string]any{
				"action":   "get",
				"scene_id": "movie_time",
			},
			wantError: false,
		},
		{
			name: "missing scene_id",
			args: map[string]any{
				"action": "get",
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "empty scene_id",
			args: map[string]any{
				"action":   "get",
				"scene_id": "",
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "client error",
			args: map[string]any{
				"action":   "get",
				"scene_id": "movie_time",
			},
			getStateErr:  errors.New("not found"),
			wantError:    true,
			wantContains: "Error getting scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockSceneClient{
				getStateFn: func(_ context.Context, entityID string) (*homeassistant.Entity, error) {
					if tt.getStateErr != nil {
						return nil, tt.getStateErr
					}
					return &homeassistant.Entity{
						EntityID:   entityID,
						State:      "scening",
						Attributes: map[string]any{"friendly_name": "Movie Time"},
					}, nil
				},
			}

			h := NewSceneHandlers()
			result, err := h.handleManageScene(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("handleManageScene() returned error: %v", err)
				return
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			if len(result.Content) == 0 {
				t.Error("Content is empty")
				return
			}

			content := result.Content[0].Text
			if tt.wantContains != "" && !strings.Contains(content, tt.wantContains) {
				t.Errorf("Content = %q, want to contain %q", content, tt.wantContains)
			}
		})
	}
}

func TestSceneHandlers_ManageScene_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           map[string]any
		createSceneErr error
		wantError      bool
		wantContains   string
	}{
		{
			name: "success",
			args: map[string]any{
				"action":   "create",
				"scene_id": "movie_time",
				"name":     "Movie Time",
				"entities": map[string]any{
					"light.living_room": "off",
				},
			},
			wantError:    false,
			wantContains: "created successfully",
		},
		{
			name: "success with detailed entities",
			args: map[string]any{
				"action":   "create",
				"scene_id": "movie_time",
				"name":     "Movie Time",
				"icon":     "mdi:movie",
				"entities": map[string]any{
					"light.living_room": map[string]any{
						"state": "on",
						"attributes": map[string]any{
							"brightness": 50,
							"color_temp": 400,
						},
					},
					"media_player.tv": "on",
				},
			},
			wantError:    false,
			wantContains: "created successfully",
		},
		{
			name: "missing scene_id",
			args: map[string]any{
				"action": "create",
				"name":   "Movie Time",
				"entities": map[string]any{
					"light.living_room": "off",
				},
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "empty scene_id",
			args: map[string]any{
				"action":   "create",
				"scene_id": "",
				"name":     "Movie Time",
				"entities": map[string]any{
					"light.living_room": "off",
				},
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "missing name",
			args: map[string]any{
				"action":   "create",
				"scene_id": "movie_time",
				"entities": map[string]any{
					"light.living_room": "off",
				},
			},
			wantError:    true,
			wantContains: "name is required",
		},
		{
			name: "empty name",
			args: map[string]any{
				"action":   "create",
				"scene_id": "movie_time",
				"name":     "",
				"entities": map[string]any{
					"light.living_room": "off",
				},
			},
			wantError:    true,
			wantContains: "name is required",
		},
		{
			name: "missing entities",
			args: map[string]any{
				"action":   "create",
				"scene_id": "movie_time",
				"name":     "Movie Time",
			},
			wantError:    true,
			wantContains: "entities is required",
		},
		{
			name: "empty entities",
			args: map[string]any{
				"action":   "create",
				"scene_id": "movie_time",
				"name":     "Movie Time",
				"entities": map[string]any{},
			},
			wantError:    true,
			wantContains: "entities is required",
		},
		{
			name: "invalid entity state format",
			args: map[string]any{
				"action":   "create",
				"scene_id": "movie_time",
				"name":     "Movie Time",
				"entities": map[string]any{
					"light.living_room": 123,
				},
			},
			wantError:    true,
			wantContains: "Invalid state format",
		},
		{
			name: "client error",
			args: map[string]any{
				"action":   "create",
				"scene_id": "movie_time",
				"name":     "Movie Time",
				"entities": map[string]any{
					"light.living_room": "off",
				},
			},
			createSceneErr: errors.New("creation failed"),
			wantError:      true,
			wantContains:   "Error creating scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockSceneClient{
				createSceneFn: func(_ context.Context, _ string, _ homeassistant.SceneConfig) error {
					return tt.createSceneErr
				},
			}

			h := NewSceneHandlers()
			result, err := h.handleManageScene(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("handleManageScene() returned error: %v", err)
				return
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			if len(result.Content) == 0 {
				t.Error("Content is empty")
				return
			}

			content := result.Content[0].Text
			if tt.wantContains != "" && !strings.Contains(content, tt.wantContains) {
				t.Errorf("Content = %q, want to contain %q", content, tt.wantContains)
			}
		})
	}
}

func TestSceneHandlers_ManageScene_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           map[string]any
		getStateErr    error
		updateSceneErr error
		wantError      bool
		wantContains   string
	}{
		{
			name: "success",
			args: map[string]any{
				"action":   "update",
				"scene_id": "movie_time",
				"name":     "Updated Movie Time",
			},
			wantError:    false,
			wantContains: "updated successfully",
		},
		{
			name: "success with entities",
			args: map[string]any{
				"action":   "update",
				"scene_id": "movie_time",
				"name":     "Updated Movie Time",
				"icon":     "mdi:movie-open",
				"entities": map[string]any{
					"light.living_room": "on",
				},
			},
			wantError:    false,
			wantContains: "updated successfully",
		},
		{
			name: "missing scene_id",
			args: map[string]any{
				"action": "update",
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "empty scene_id",
			args: map[string]any{
				"action":   "update",
				"scene_id": "",
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "get state error",
			args: map[string]any{
				"action":   "update",
				"scene_id": "movie_time",
			},
			getStateErr:  errors.New("not found"),
			wantError:    true,
			wantContains: "Error getting current scene",
		},
		{
			name: "update error",
			args: map[string]any{
				"action":   "update",
				"scene_id": "movie_time",
				"name":     "Updated",
			},
			updateSceneErr: errors.New("update failed"),
			wantError:      true,
			wantContains:   "Error updating scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockSceneClient{
				getStateFn: func(_ context.Context, entityID string) (*homeassistant.Entity, error) {
					if tt.getStateErr != nil {
						return nil, tt.getStateErr
					}
					return &homeassistant.Entity{
						EntityID:   entityID,
						State:      "scening",
						Attributes: map[string]any{"friendly_name": "Movie Time"},
					}, nil
				},
				updateSceneFn: func(_ context.Context, _ string, _ homeassistant.SceneConfig) error {
					return tt.updateSceneErr
				},
			}

			h := NewSceneHandlers()
			result, err := h.handleManageScene(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("handleManageScene() returned error: %v", err)
				return
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			if len(result.Content) == 0 {
				t.Error("Content is empty")
				return
			}

			content := result.Content[0].Text
			if tt.wantContains != "" && !strings.Contains(content, tt.wantContains) {
				t.Errorf("Content = %q, want to contain %q", content, tt.wantContains)
			}
		})
	}
}

func TestSceneHandlers_ManageScene_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           map[string]any
		deleteSceneErr error
		wantError      bool
		wantContains   string
	}{
		{
			name: "success",
			args: map[string]any{
				"action":   "delete",
				"scene_id": "movie_time",
			},
			wantError:    false,
			wantContains: "deleted successfully",
		},
		{
			name: "missing scene_id",
			args: map[string]any{
				"action": "delete",
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "empty scene_id",
			args: map[string]any{
				"action":   "delete",
				"scene_id": "",
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "client error",
			args: map[string]any{
				"action":   "delete",
				"scene_id": "movie_time",
			},
			deleteSceneErr: errors.New("deletion failed"),
			wantError:      true,
			wantContains:   "Error deleting scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockSceneClient{
				deleteSceneFn: func(_ context.Context, _ string) error {
					return tt.deleteSceneErr
				},
			}

			h := NewSceneHandlers()
			result, err := h.handleManageScene(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("handleManageScene() returned error: %v", err)
				return
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			if len(result.Content) == 0 {
				t.Error("Content is empty")
				return
			}

			content := result.Content[0].Text
			if tt.wantContains != "" && !strings.Contains(content, tt.wantContains) {
				t.Errorf("Content = %q, want to contain %q", content, tt.wantContains)
			}
		})
	}
}

func TestSceneHandlers_ManageScene_Activate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		args           map[string]any
		callServiceErr error
		wantError      bool
		wantContains   string
	}{
		{
			name: "success",
			args: map[string]any{
				"action":   "activate",
				"scene_id": "movie_time",
			},
			wantError:    false,
			wantContains: "activated successfully",
		},
		{
			name: "success with transition",
			args: map[string]any{
				"action":     "activate",
				"scene_id":   "movie_time",
				"transition": 2.5,
			},
			wantError:    false,
			wantContains: "activated successfully",
		},
		{
			name: "missing scene_id",
			args: map[string]any{
				"action": "activate",
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "empty scene_id",
			args: map[string]any{
				"action":   "activate",
				"scene_id": "",
			},
			wantError:    true,
			wantContains: "scene_id is required",
		},
		{
			name: "client error",
			args: map[string]any{
				"action":   "activate",
				"scene_id": "movie_time",
			},
			callServiceErr: errors.New("activation failed"),
			wantError:      true,
			wantContains:   "Error activating scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockSceneClient{
				callServiceFn: func(_ context.Context, domain, service string, _ map[string]any) ([]homeassistant.Entity, error) {
					if domain != "scene" {
						t.Errorf("wrong domain: %s", domain)
					}
					if service != "turn_on" {
						t.Errorf("wrong service: %s", service)
					}
					return nil, tt.callServiceErr
				},
			}

			h := NewSceneHandlers()
			result, err := h.handleManageScene(context.Background(), client, tt.args)

			if err != nil {
				t.Errorf("handleManageScene() returned error: %v", err)
				return
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			if len(result.Content) == 0 {
				t.Error("Content is empty")
				return
			}

			content := result.Content[0].Text
			if tt.wantContains != "" && !strings.Contains(content, tt.wantContains) {
				t.Errorf("Content = %q, want to contain %q", content, tt.wantContains)
			}
		})
	}
}

func TestSceneHandlers_ManageScene_InvalidAction(t *testing.T) {
	t.Parallel()

	h := NewSceneHandlers()
	client := &mockSceneClient{}

	// Test missing action
	result, err := h.handleManageScene(context.Background(), client, map[string]any{})
	if err != nil {
		t.Errorf("handleManageScene() returned error: %v", err)
		return
	}
	if !result.IsError {
		t.Error("Expected error for missing action")
	}
	if !strings.Contains(result.Content[0].Text, "action is required") {
		t.Errorf("Expected 'action is required' error, got: %s", result.Content[0].Text)
	}

	// Test invalid action
	result, err = h.handleManageScene(context.Background(), client, map[string]any{"action": "invalid"})
	if err != nil {
		t.Errorf("handleManageScene() returned error: %v", err)
		return
	}
	if !result.IsError {
		t.Error("Expected error for invalid action")
	}
	if !strings.Contains(result.Content[0].Text, "invalid action") {
		t.Errorf("Expected 'invalid action' error, got: %s", result.Content[0].Text)
	}
}

func TestFindSceneByID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		searchID     string
		scenes       []homeassistant.Entity
		stateMap     map[string]*homeassistant.Entity
		wantFound    bool
		wantEntityID string
	}{
		{
			name:     "find by entity_id",
			searchID: "scene.movie_time",
			scenes: []homeassistant.Entity{
				{EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			stateMap: map[string]*homeassistant.Entity{
				"scene.movie_time": {EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			wantFound:    true,
			wantEntityID: "scene.movie_time",
		},
		{
			name:     "find by friendly_name - partial match",
			searchID: "movie time",
			scenes: []homeassistant.Entity{
				{EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			stateMap: map[string]*homeassistant.Entity{
				"scene.movie_time": {EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			wantFound:    true,
			wantEntityID: "scene.movie_time",
		},
		{
			name:     "find by friendly_name - case insensitive",
			searchID: "MOVIE TIME",
			scenes: []homeassistant.Entity{
				{EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			stateMap: map[string]*homeassistant.Entity{
				"scene.movie_time": {EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			wantFound:    true,
			wantEntityID: "scene.movie_time",
		},
		{
			name:     "find by friendly_name - partial match with 'Scene' suffix",
			searchID: "Scene",
			scenes: []homeassistant.Entity{
				{EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			stateMap: map[string]*homeassistant.Entity{
				"scene.movie_time": {EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			wantFound:    true,
			wantEntityID: "scene.movie_time",
		},
		{
			name:     "not found - no matching friendly_name",
			searchID: "nonexistent",
			scenes: []homeassistant.Entity{
				{EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			stateMap: map[string]*homeassistant.Entity{
				"scene.movie_time": {EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
			},
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockSceneClient{
				listScenesFn: func(_ context.Context) ([]homeassistant.Entity, error) {
					return tt.scenes, nil
				},
				getStateFn: func(_ context.Context, entityID string) (*homeassistant.Entity, error) {
					if e, ok := tt.stateMap[entityID]; ok {
						return e, nil
					}
					return nil, errors.New("not found")
				},
			}

			h := &SceneHandlers{}
			result, err := h.findSceneByID(context.Background(), client, tt.searchID)

			if tt.wantFound {
				if err != nil {
					t.Errorf("findSceneByID() unexpected error = %v", err)
					return
				}
				if result == nil {
					t.Error("findSceneByID() returned nil, want scene")
					return
				}
				if result.EntityID != tt.wantEntityID {
					t.Errorf("findSceneByID() EntityID = %q, want %q", result.EntityID, tt.wantEntityID)
				}
			} else {
				if err == nil {
					t.Error("findSceneByID() expected error, got nil")
				}
			}
		})
	}
}

func TestManageScene_GetByFriendlyName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		args         map[string]any
		setupClient  func() *mockSceneClient
		wantError    bool
		wantContains string
	}{
		{
			name: "get by friendly_name - partial match",
			args: map[string]any{"action": "get", "scene_id": "movie time"},
			setupClient: func() *mockSceneClient {
				return &mockSceneClient{
					getStateFn: func(_ context.Context, entityID string) (*homeassistant.Entity, error) {
						if entityID == "scene.movie time" {
							return nil, errors.New("not found")
						}
						if entityID == "scene.movie_time" {
							return &homeassistant.Entity{
								EntityID: "scene.movie_time",
								State:    "scening",
								Attributes: map[string]any{
									"friendly_name": "Movie Time Scene",
								},
							}, nil
						}
						return nil, errors.New("not found")
					},
					listScenesFn: func(_ context.Context) ([]homeassistant.Entity, error) {
						return []homeassistant.Entity{
							{EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
						}, nil
					},
				}
			},
			wantContains: "Movie Time Scene",
		},
		{
			name: "get by friendly_name - case insensitive",
			args: map[string]any{"action": "get", "scene_id": "MOVIE"},
			setupClient: func() *mockSceneClient {
				return &mockSceneClient{
					getStateFn: func(_ context.Context, entityID string) (*homeassistant.Entity, error) {
						if entityID == "scene.MOVIE" {
							return nil, errors.New("not found")
						}
						if entityID == "scene.movie_time" {
							return &homeassistant.Entity{
								EntityID: "scene.movie_time",
								State:    "scening",
								Attributes: map[string]any{
									"friendly_name": "Movie Time Scene",
								},
							}, nil
						}
						return nil, errors.New("not found")
					},
					listScenesFn: func(_ context.Context) ([]homeassistant.Entity, error) {
						return []homeassistant.Entity{
							{EntityID: "scene.movie_time", State: "scening", Attributes: map[string]any{"friendly_name": "Movie Time Scene"}},
						}, nil
					},
				}
			},
			wantContains: "Movie Time Scene",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := NewSceneHandlers()
			result, err := h.handleManageScene(context.Background(), tt.setupClient(), tt.args)
			if err != nil {
				t.Fatalf("handleManageScene() unexpected error = %v", err)
			}

			if result.IsError != tt.wantError {
				t.Errorf("IsError = %v, want %v", result.IsError, tt.wantError)
			}

			content := result.Content[0].Text
			if tt.wantContains != "" && !strings.Contains(content, tt.wantContains) {
				t.Errorf("Expected content to contain %q, got: %s", tt.wantContains, content)
			}
		})
	}
}

// TestSceneHandlers_IDNormalization tests that scene_id inputs are properly normalized
// to avoid double-prefix bugs (e.g., scene.scene.movie_night) and to resolve config IDs.
func TestSceneHandlers_IDNormalization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		action            string
		inputID           string
		wantGetStateID    string // For update action
		wantUpdateSceneID string
		wantDeleteSceneID string
		additionalArgs    map[string]any
	}{
		{
			name:              "update - with scene. prefix",
			action:            "update",
			inputID:           "scene.movie_night",
			wantGetStateID:    "scene.movie_night", // Should NOT be scene.scene.movie_night
			wantUpdateSceneID: "movie_night",       // Should strip prefix for REST API
			additionalArgs: map[string]any{
				"name": "Updated Movie Night",
			},
		},
		{
			name:              "update - without prefix",
			action:            "update",
			inputID:           "movie_night",
			wantGetStateID:    "scene.movie_night", // Should add prefix for GetState
			wantUpdateSceneID: "movie_night",
			additionalArgs: map[string]any{
				"name": "Updated Movie Night",
			},
		},
		{
			name:              "delete - with scene. prefix",
			action:            "delete",
			inputID:           "scene.movie_night",
			wantDeleteSceneID: "movie_night", // Should strip prefix for REST API
		},
		{
			name:              "delete - without prefix",
			action:            "delete",
			inputID:           "movie_night",
			wantDeleteSceneID: "movie_night",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			client := &mockSceneClient{}
			h := &SceneHandlers{}

			args := map[string]any{
				"action":   tt.action,
				"scene_id": tt.inputID,
			}
			for k, v := range tt.additionalArgs {
				args[k] = v
			}

			_, err := h.handleManageScene(context.Background(), client, args)
			if err != nil {
				t.Fatalf("handleManageScene() unexpected error = %v", err)
			}

			// Verify correct IDs were used
			if tt.wantGetStateID != "" && client.lastGetStateID != tt.wantGetStateID {
				t.Errorf("GetState called with ID %q, want %q", client.lastGetStateID, tt.wantGetStateID)
			}
			if tt.wantUpdateSceneID != "" && client.lastUpdateSceneID != tt.wantUpdateSceneID {
				t.Errorf("UpdateScene called with ID %q, want %q", client.lastUpdateSceneID, tt.wantUpdateSceneID)
			}
			if tt.wantDeleteSceneID != "" && client.lastDeleteSceneID != tt.wantDeleteSceneID {
				t.Errorf("DeleteScene called with ID %q, want %q", client.lastDeleteSceneID, tt.wantDeleteSceneID)
			}
		})
	}
}
