//go:build integration

package integration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zorak1103/ha-mcp/internal/homeassistant"
)

type PersonIntegrationTestSuite struct {
	PersonTestSuite
}

func TestPersonIntegration(t *testing.T) {
	suite.Run(t, new(PersonIntegrationTestSuite))
}

func (s *PersonIntegrationTestSuite) TestPersonLifecycle() {
	personName := GenerateTestID("person")

	s.RegisterCleanup(func() {
		persons, _ := s.Client().GetPersons(s.Context())
		for _, person := range persons {
			if person.Name == personName {
				_ = s.Client().DeletePerson(s.Context(), person.ID)
			}
		}
	})

	// Create person (minimal fields - no user_id or device_trackers)
	personConfig := homeassistant.PersonConfig{
		Name: personName,
	}

	created, err := s.Client().CreatePerson(s.Context(), personConfig)
	s.Require().NoError(err, "Failed to create person")
	s.Require().NotNil(created)
	s.Equal(personName, created.Name)

	personID := created.ID

	// Allow time for registry to update
	time.Sleep(500 * time.Millisecond)

	// Verify person appears in registry
	person, err := s.FindPersonByID(personID)
	s.Require().NoError(err, "Person should appear in registry")
	s.Equal(personName, person.Name)

	// Update person (name only)
	updatedName := GenerateTestID("person_updated")
	updateConfig := homeassistant.PersonConfig{
		Name: updatedName,
	}

	updated, err := s.Client().UpdatePerson(s.Context(), personID, updateConfig)
	s.Require().NoError(err, "Failed to update person")
	s.Equal(updatedName, updated.Name)

	time.Sleep(500 * time.Millisecond)

	// Verify update
	person, err = s.FindPersonByID(personID)
	s.Require().NoError(err)
	s.Equal(updatedName, person.Name)

	// Delete person
	err = s.Client().DeletePerson(s.Context(), personID)
	s.Require().NoError(err, "Failed to delete person")

	time.Sleep(500 * time.Millisecond)

	// Verify person is gone
	_, err = s.FindPersonByID(personID)
	s.Error(err, "Person should be deleted from registry")
}

func (s *PersonIntegrationTestSuite) TestPersonWithAllFields() {
	personName := GenerateTestID("person_full")
	// Note: device_trackers array would require actual device tracker entities
	// For integration test, we test with empty array to verify the field works

	s.RegisterCleanup(func() {
		persons, _ := s.Client().GetPersons(s.Context())
		for _, person := range persons {
			if person.Name == personName {
				_ = s.Client().DeletePerson(s.Context(), person.ID)
			}
		}
	})

	// Create person with all fields (except user_id which requires actual user)
	personConfig := homeassistant.PersonConfig{
		Name:           personName,
		Picture:        "/local/test_picture.png",
		DeviceTrackers: []string{}, // Empty array to test field works
	}

	created, err := s.Client().CreatePerson(s.Context(), personConfig)
	s.Require().NoError(err, "Failed to create person with all fields")
	s.Require().NotNil(created)

	personID := created.ID

	time.Sleep(500 * time.Millisecond)

	// Verify all fields
	person, err := s.FindPersonByID(personID)
	s.Require().NoError(err)
	s.Equal(personName, person.Name)
	s.Equal("/local/test_picture.png", person.Picture)

	// Cleanup
	_ = s.Client().DeletePerson(s.Context(), personID)
}

func (s *PersonIntegrationTestSuite) TestPersonUpdatePartial() {
	personName := GenerateTestID("person_partial")

	s.RegisterCleanup(func() {
		persons, _ := s.Client().GetPersons(s.Context())
		for _, person := range persons {
			if person.Name == personName {
				_ = s.Client().DeletePerson(s.Context(), person.ID)
			}
		}
	})

	// Create person with name only
	personConfig := homeassistant.PersonConfig{
		Name: personName,
	}

	created, err := s.Client().CreatePerson(s.Context(), personConfig)
	s.Require().NoError(err)

	personID := created.ID

	time.Sleep(500 * time.Millisecond)

	// Update only picture (name should remain)
	updateConfig := homeassistant.PersonConfig{
		Picture: "/local/updated_picture.png",
	}

	_, err = s.Client().UpdatePerson(s.Context(), personID, updateConfig)
	s.Require().NoError(err, "Failed to update person with partial config")

	time.Sleep(500 * time.Millisecond)

	// Verify name unchanged, picture updated
	person, err := s.FindPersonByID(personID)
	s.Require().NoError(err)
	s.Equal(personName, person.Name, "Name should remain unchanged")
	s.Equal("/local/updated_picture.png", person.Picture, "Picture should be updated")

	// Cleanup
	_ = s.Client().DeletePerson(s.Context(), personID)
}

func (s *PersonIntegrationTestSuite) TestMultiplePersons() {
	person1Name := GenerateTestID("person_1")
	person2Name := GenerateTestID("person_2")

	var person1ID, person2ID string

	s.RegisterCleanup(func() {
		if person1ID != "" {
			_ = s.Client().DeletePerson(s.Context(), person1ID)
		}
		if person2ID != "" {
			_ = s.Client().DeletePerson(s.Context(), person2ID)
		}
	})

	// Create first person
	config1 := homeassistant.PersonConfig{
		Name: person1Name,
	}

	created1, err := s.Client().CreatePerson(s.Context(), config1)
	s.Require().NoError(err, "Failed to create person 1")
	person1ID = created1.ID

	// Create second person
	config2 := homeassistant.PersonConfig{
		Name: person2Name,
	}

	created2, err := s.Client().CreatePerson(s.Context(), config2)
	s.Require().NoError(err, "Failed to create person 2")
	person2ID = created2.ID

	time.Sleep(500 * time.Millisecond)

	// Verify both persons exist in registry
	person1, err := s.FindPersonByID(person1ID)
	s.Require().NoError(err, "Person 1 should exist")
	s.Equal(person1Name, person1.Name)

	person2, err := s.FindPersonByID(person2ID)
	s.Require().NoError(err, "Person 2 should exist")
	s.Equal(person2Name, person2.Name)

	// Delete both persons
	err = s.Client().DeletePerson(s.Context(), person1ID)
	s.Require().NoError(err, "Failed to delete person 1")

	err = s.Client().DeletePerson(s.Context(), person2ID)
	s.Require().NoError(err, "Failed to delete person 2")

	time.Sleep(500 * time.Millisecond)

	// Verify both are gone
	_, err = s.FindPersonByID(person1ID)
	s.Error(err, "Person 1 should be deleted")

	_, err = s.FindPersonByID(person2ID)
	s.Error(err, "Person 2 should be deleted")
}
