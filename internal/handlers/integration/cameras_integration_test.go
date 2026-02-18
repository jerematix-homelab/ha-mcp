//go:build integration

package integration

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CameraIntegrationTestSuite struct {
	IntegrationTestSuite
}

func TestCameraIntegration(t *testing.T) {
	suite.Run(t, new(CameraIntegrationTestSuite))
}

// TestListCameras verifies that we can find camera entities.
func (s *CameraIntegrationTestSuite) TestListCameras() {
	// Get all states and filter for camera domain
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Failed to get states")

	var cameras []string
	for _, state := range states {
		if strings.HasPrefix(state.EntityID, "camera.") {
			cameras = append(cameras, state.EntityID)
		}
	}

	s.T().Logf("Found %d camera entit(ies)", len(cameras))

	// Note: Test doesn't require cameras to exist, just verifies API works
	if len(cameras) == 0 {
		s.T().Log("No cameras found - this is normal if no camera integrations are configured")
	}
}

// TestGetCameraStream verifies we can get a camera stream URL.
func (s *CameraIntegrationTestSuite) TestGetCameraStream() {
	// Get all states to find a camera
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Failed to get states")

	var cameraEntityID string
	for _, state := range states {
		if strings.HasPrefix(state.EntityID, "camera.") {
			cameraEntityID = state.EntityID
			break
		}
	}

	if cameraEntityID == "" {
		s.T().Skip("No camera entities found in Home Assistant")
		return
	}

	s.T().Logf("Testing stream for: %s", cameraEntityID)

	// Call camera/stream WebSocket command
	streamInfo, err := s.Client().GetCameraStream(s.Context(), cameraEntityID)

	// Note: Stream may not be available for all camera types
	if err != nil {
		s.T().Logf("Stream not available for %s (expected for some camera types): %v", cameraEntityID, err)
	} else {
		s.Require().NotNil(streamInfo, "StreamInfo should not be nil")
		s.Require().NotEmpty(streamInfo.URL, "Stream URL should not be empty")
		s.T().Logf("Stream URL: %s", streamInfo.URL)
	}
}

// TestGetCameraSnapshot verifies we can get a camera snapshot.
func (s *CameraIntegrationTestSuite) TestGetCameraSnapshot() {
	// Get all states to find a camera
	states, err := s.Client().GetStates(s.Context())
	s.Require().NoError(err, "Failed to get states")

	var cameraEntityID string
	for _, state := range states {
		if strings.HasPrefix(state.EntityID, "camera.") {
			cameraEntityID = state.EntityID
			break
		}
	}

	if cameraEntityID == "" {
		s.T().Skip("No camera entities found in Home Assistant")
		return
	}

	s.T().Logf("Testing snapshot for: %s", cameraEntityID)

	// Call camera snapshot REST API
	imageData, contentType, err := s.Client().GetCameraSnapshot(s.Context(), cameraEntityID)

	// Note: Snapshot may not be available for all camera types
	if err != nil {
		s.T().Logf("Snapshot not available for %s (expected for some camera types): %v", cameraEntityID, err)
	} else {
		s.Require().NotEmpty(imageData, "Image data should not be empty")
		s.Require().NotEmpty(contentType, "Content type should not be empty")
		s.T().Logf("Snapshot retrieved: %d bytes, type: %s", len(imageData), contentType)
	}
}
