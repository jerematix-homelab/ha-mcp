//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type TagIntegrationTestSuite struct {
	TagTestSuite
}

func TestTagIntegration(t *testing.T) {
	suite.Run(t, new(TagIntegrationTestSuite))
}

func (s *TagIntegrationTestSuite) TestTagLifecycle() {
	tagName := GenerateTestID("tag")
	tagID := GenerateTestID("tag_id")

	s.RegisterCleanup(func() {
		tags, _ := s.Client().GetTags(s.Context())
		for _, tag := range tags {
			if tag.TagID == tagID {
				_ = s.Client().DeleteTag(s.Context(), tag.TagID)
			}
		}
	})

	// Create tag
	tagConfig := homeassistant.TagConfig{
		TagID: tagID,
		Name:  tagName,
	}

	created, err := s.Client().CreateTag(s.Context(), tagConfig)
	s.Require().NoError(err, "Failed to create tag")
	s.Require().NotNil(created)
	s.Equal(tagID, created.TagID)
	s.Equal(tagName, created.Name)

	// Allow time for registry to update
	time.Sleep(500 * time.Millisecond)

	// Verify tag appears in registry
	tag, err := s.FindTagByID(tagID)
	s.Require().NoError(err, "Tag should appear in registry")
	s.Equal(tagID, tag.TagID)
	s.Equal(tagName, tag.Name)

	// Update tag (name + description)
	updatedName := GenerateTestID("tag_updated")
	updateConfig := homeassistant.TagConfig{
		Name:        updatedName,
		Description: "Updated tag description",
	}

	updated, err := s.Client().UpdateTag(s.Context(), tagID, updateConfig)
	s.Require().NoError(err, "Failed to update tag")
	s.Equal(updatedName, updated.Name)
	s.Equal("Updated tag description", updated.Description)

	time.Sleep(500 * time.Millisecond)

	// Verify update
	tag, err = s.FindTagByID(tagID)
	s.Require().NoError(err)
	s.Equal(updatedName, tag.Name)
	s.Equal("Updated tag description", tag.Description)

	// Delete tag
	err = s.Client().DeleteTag(s.Context(), tagID)
	s.Require().NoError(err, "Failed to delete tag")

	time.Sleep(500 * time.Millisecond)

	// Verify tag is gone
	_, err = s.FindTagByID(tagID)
	s.Error(err, "Tag should be deleted from registry")
}

func (s *TagIntegrationTestSuite) TestTagWithAllFields() {
	tagName := GenerateTestID("tag_full")
	tagID := GenerateTestID("tag_full_id")

	s.RegisterCleanup(func() {
		tags, _ := s.Client().GetTags(s.Context())
		for _, tag := range tags {
			if tag.TagID == tagID {
				_ = s.Client().DeleteTag(s.Context(), tag.TagID)
			}
		}
	})

	// Create tag with all fields
	tagConfig := homeassistant.TagConfig{
		TagID:       tagID,
		Name:        tagName,
		Description: "Full tag description",
	}

	created, err := s.Client().CreateTag(s.Context(), tagConfig)
	s.Require().NoError(err, "Failed to create tag with all fields")
	s.Require().NotNil(created)

	time.Sleep(500 * time.Millisecond)

	// Verify all fields
	tag, err := s.FindTagByID(tagID)
	s.Require().NoError(err)
	s.Equal(tagID, tag.TagID)
	s.Equal(tagName, tag.Name)
	s.Equal("Full tag description", tag.Description)

	// Cleanup
	_ = s.Client().DeleteTag(s.Context(), tagID)
}

func (s *TagIntegrationTestSuite) TestTagUpdatePartial() {
	tagName := GenerateTestID("tag_partial")
	tagID := GenerateTestID("tag_partial_id")

	s.RegisterCleanup(func() {
		tags, _ := s.Client().GetTags(s.Context())
		for _, tag := range tags {
			if tag.TagID == tagID {
				_ = s.Client().DeleteTag(s.Context(), tag.TagID)
			}
		}
	})

	// Create tag with name only
	tagConfig := homeassistant.TagConfig{
		TagID: tagID,
		Name:  tagName,
	}

	_, err := s.Client().CreateTag(s.Context(), tagConfig)
	s.Require().NoError(err)

	time.Sleep(500 * time.Millisecond)

	// Update only description (name should remain)
	updateConfig := homeassistant.TagConfig{
		Description: "Partial update description",
	}

	_, err = s.Client().UpdateTag(s.Context(), tagID, updateConfig)
	s.Require().NoError(err, "Failed to update tag with partial config")

	time.Sleep(500 * time.Millisecond)

	// Verify name unchanged, description updated
	tag, err := s.FindTagByID(tagID)
	s.Require().NoError(err)
	s.Equal(tagName, tag.Name, "Name should remain unchanged")
	s.Equal("Partial update description", tag.Description, "Description should be updated")

	// Cleanup
	_ = s.Client().DeleteTag(s.Context(), tagID)
}

func (s *TagIntegrationTestSuite) TestMultipleTags() {
	tag1Name := GenerateTestID("tag_1")
	tag1ID := GenerateTestID("tag_1_id")
	tag2Name := GenerateTestID("tag_2")
	tag2ID := GenerateTestID("tag_2_id")

	s.RegisterCleanup(func() {
		_ = s.Client().DeleteTag(s.Context(), tag1ID)
		_ = s.Client().DeleteTag(s.Context(), tag2ID)
	})

	// Create first tag
	config1 := homeassistant.TagConfig{
		TagID: tag1ID,
		Name:  tag1Name,
	}

	created1, err := s.Client().CreateTag(s.Context(), config1)
	s.Require().NoError(err, "Failed to create tag 1")
	s.Equal(tag1ID, created1.TagID)

	// Create second tag
	config2 := homeassistant.TagConfig{
		TagID: tag2ID,
		Name:  tag2Name,
	}

	created2, err := s.Client().CreateTag(s.Context(), config2)
	s.Require().NoError(err, "Failed to create tag 2")
	s.Equal(tag2ID, created2.TagID)

	time.Sleep(500 * time.Millisecond)

	// Verify both tags exist in registry
	tag1, err := s.FindTagByID(tag1ID)
	s.Require().NoError(err, "Tag 1 should exist")
	s.Equal(tag1Name, tag1.Name)

	tag2, err := s.FindTagByID(tag2ID)
	s.Require().NoError(err, "Tag 2 should exist")
	s.Equal(tag2Name, tag2.Name)

	// Delete both tags
	err = s.Client().DeleteTag(s.Context(), tag1ID)
	s.Require().NoError(err, "Failed to delete tag 1")

	err = s.Client().DeleteTag(s.Context(), tag2ID)
	s.Require().NoError(err, "Failed to delete tag 2")

	time.Sleep(500 * time.Millisecond)

	// Verify both are gone
	_, err = s.FindTagByID(tag1ID)
	s.Error(err, "Tag 1 should be deleted")

	_, err = s.FindTagByID(tag2ID)
	s.Error(err, "Tag 2 should be deleted")
}
