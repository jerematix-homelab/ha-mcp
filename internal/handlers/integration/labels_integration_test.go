//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type LabelIntegrationTestSuite struct {
	LabelTestSuite
}

func TestLabelIntegration(t *testing.T) {
	suite.Run(t, new(LabelIntegrationTestSuite))
}

func (s *LabelIntegrationTestSuite) TestLabelLifecycle() {
	labelName := GenerateTestID("label")

	s.RegisterCleanup(func() {
		labels, _ := s.Client().GetLabelRegistry(s.Context())
		for _, label := range labels {
			if label.Name == labelName {
				_ = s.Client().DeleteLabel(s.Context(), label.LabelID)
			}
		}
	})

	// Create label
	labelConfig := homeassistant.LabelConfig{
		Name:  labelName,
		Icon:  "mdi:label",
		Color: "#FF5722",
	}

	created, err := s.Client().CreateLabel(s.Context(), labelConfig)
	s.Require().NoError(err, "Failed to create label")
	s.Require().NotNil(created)
	s.Equal(labelName, created.Name)
	s.Equal("mdi:label", created.Icon)
	s.Equal("#FF5722", created.Color)

	labelID := created.LabelID

	// Allow time for registry to update
	time.Sleep(500 * time.Millisecond)

	// Verify label appears in registry
	label, err := s.FindLabelByID(labelID)
	s.Require().NoError(err, "Label should appear in registry")
	s.Equal(labelName, label.Name)
	s.Equal("mdi:label", label.Icon)
	s.Equal("#FF5722", label.Color)

	// Update label (name + icon + color)
	updatedName := GenerateTestID("label_updated")
	updateConfig := homeassistant.LabelConfig{
		Name:  updatedName,
		Icon:  "mdi:tag",
		Color: "#2196F3",
	}

	updated, err := s.Client().UpdateLabel(s.Context(), labelID, updateConfig)
	s.Require().NoError(err, "Failed to update label")
	s.Equal(updatedName, updated.Name)
	s.Equal("mdi:tag", updated.Icon)
	s.Equal("#2196F3", updated.Color)

	time.Sleep(500 * time.Millisecond)

	// Verify update
	label, err = s.FindLabelByID(labelID)
	s.Require().NoError(err)
	s.Equal(updatedName, label.Name)
	s.Equal("mdi:tag", label.Icon)
	s.Equal("#2196F3", label.Color)

	// Delete label
	err = s.Client().DeleteLabel(s.Context(), labelID)
	s.Require().NoError(err, "Failed to delete label")

	time.Sleep(500 * time.Millisecond)

	// Verify label is gone
	_, err = s.FindLabelByID(labelID)
	s.Error(err, "Label should be deleted from registry")
}

func (s *LabelIntegrationTestSuite) TestLabelWithAllFields() {
	labelName := GenerateTestID("label_full")

	s.RegisterCleanup(func() {
		labels, _ := s.Client().GetLabelRegistry(s.Context())
		for _, label := range labels {
			if label.Name == labelName {
				_ = s.Client().DeleteLabel(s.Context(), label.LabelID)
			}
		}
	})

	// Create label with all fields
	labelConfig := homeassistant.LabelConfig{
		Name:        labelName,
		Icon:        "mdi:test-tube",
		Color:       "#9C27B0",
		Description: "Test label description",
	}

	created, err := s.Client().CreateLabel(s.Context(), labelConfig)
	s.Require().NoError(err, "Failed to create label with all fields")
	s.Require().NotNil(created)

	labelID := created.LabelID

	time.Sleep(500 * time.Millisecond)

	// Verify all fields
	label, err := s.FindLabelByID(labelID)
	s.Require().NoError(err)
	s.Equal(labelName, label.Name)
	s.Equal("mdi:test-tube", label.Icon)
	s.Equal("#9C27B0", label.Color)
	s.Equal("Test label description", label.Description)

	// Cleanup
	_ = s.Client().DeleteLabel(s.Context(), labelID)
}

func (s *LabelIntegrationTestSuite) TestLabelUpdatePartial() {
	labelName := GenerateTestID("label_partial")

	s.RegisterCleanup(func() {
		labels, _ := s.Client().GetLabelRegistry(s.Context())
		for _, label := range labels {
			if label.Name == labelName {
				_ = s.Client().DeleteLabel(s.Context(), label.LabelID)
			}
		}
	})

	// Create label with icon and color
	labelConfig := homeassistant.LabelConfig{
		Name:  labelName,
		Icon:  "mdi:home",
		Color: "#4CAF50",
	}

	created, err := s.Client().CreateLabel(s.Context(), labelConfig)
	s.Require().NoError(err)

	labelID := created.LabelID

	time.Sleep(500 * time.Millisecond)

	// Update only description (icon and color should remain)
	updateConfig := homeassistant.LabelConfig{
		Description: "New description",
	}

	_, err = s.Client().UpdateLabel(s.Context(), labelID, updateConfig)
	s.Require().NoError(err, "Failed to update label with partial config")

	time.Sleep(500 * time.Millisecond)

	// Verify icon and color unchanged, description updated
	label, err := s.FindLabelByID(labelID)
	s.Require().NoError(err)
	s.Equal("mdi:home", label.Icon, "Icon should remain unchanged")
	s.Equal("#4CAF50", label.Color, "Color should remain unchanged")
	s.Equal("New description", label.Description, "Description should be updated")

	// Cleanup
	_ = s.Client().DeleteLabel(s.Context(), labelID)
}

func (s *LabelIntegrationTestSuite) TestMultipleLabels() {
	label1Name := GenerateTestID("label_1")
	label2Name := GenerateTestID("label_2")

	var label1ID, label2ID string

	s.RegisterCleanup(func() {
		if label1ID != "" {
			_ = s.Client().DeleteLabel(s.Context(), label1ID)
		}
		if label2ID != "" {
			_ = s.Client().DeleteLabel(s.Context(), label2ID)
		}
	})

	// Create first label
	config1 := homeassistant.LabelConfig{
		Name:  label1Name,
		Icon:  "mdi:numeric-1",
		Color: "#F44336",
	}

	created1, err := s.Client().CreateLabel(s.Context(), config1)
	s.Require().NoError(err, "Failed to create label 1")
	label1ID = created1.LabelID

	// Create second label
	config2 := homeassistant.LabelConfig{
		Name:  label2Name,
		Icon:  "mdi:numeric-2",
		Color: "#03A9F4",
	}

	created2, err := s.Client().CreateLabel(s.Context(), config2)
	s.Require().NoError(err, "Failed to create label 2")
	label2ID = created2.LabelID

	time.Sleep(500 * time.Millisecond)

	// Verify both labels exist in registry
	label1, err := s.FindLabelByID(label1ID)
	s.Require().NoError(err, "Label 1 should exist")
	s.Equal(label1Name, label1.Name)

	label2, err := s.FindLabelByID(label2ID)
	s.Require().NoError(err, "Label 2 should exist")
	s.Equal(label2Name, label2.Name)

	// Delete both labels
	err = s.Client().DeleteLabel(s.Context(), label1ID)
	s.Require().NoError(err, "Failed to delete label 1")

	err = s.Client().DeleteLabel(s.Context(), label2ID)
	s.Require().NoError(err, "Failed to delete label 2")

	time.Sleep(500 * time.Millisecond)

	// Verify both are gone
	_, err = s.FindLabelByID(label1ID)
	s.Error(err, "Label 1 should be deleted")

	_, err = s.FindLabelByID(label2ID)
	s.Error(err, "Label 2 should be deleted")
}
