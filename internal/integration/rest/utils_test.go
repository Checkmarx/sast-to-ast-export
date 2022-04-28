package rest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtendProjects(t *testing.T) {
	t.Run("test same order", func(t *testing.T) {
		projects := []*Project{{ID: 1, Name: "test_name", IsPublic: true, TeamID: 1}}
		projectsOData := []*ProjectOData{{ID: 1, CreatedDate: "2022-04-21T20:30:59.39+03:00",
			CustomFields: []*CustomField{{FieldName: "Creator_custom_field", FieldValue: "test"}}}}

		expectedProjects := []*Project{{ID: 1, Name: "test_name", IsPublic: true, TeamID: 1,
			CreatedDate: "2022-04-21T20:30:59.39+03:00", Configuration: &Configuration{
				CustomFields: []*CustomField{{FieldName: "Creator_custom_field", FieldValue: "test"}},
			}}}

		resultProjects := ExtendProjects(projects, projectsOData)

		assert.Equal(t, expectedProjects, resultProjects)
	})

	t.Run("test wrong order", func(t *testing.T) {
		projects := []*Project{{ID: 1, Name: "test_name", IsPublic: true, TeamID: 1},
			{ID: 2, Name: "test_name_2", IsPublic: true, TeamID: 2}}
		projectsOData := []*ProjectOData{
			{ID: 2, CreatedDate: "2022-05-25T20:30:59.39+03:00",
				CustomFields: []*CustomField{}},
			{ID: 1, CreatedDate: "2022-04-21T20:30:59.39+03:00",
				CustomFields: []*CustomField{{FieldName: "Creator_custom_field", FieldValue: "test"}}}}

		expectedProjects := []*Project{{ID: 1, Name: "test_name", IsPublic: true, TeamID: 1,
			CreatedDate: "2022-04-21T20:30:59.39+03:00", Configuration: &Configuration{
				CustomFields: []*CustomField{{FieldName: "Creator_custom_field", FieldValue: "test"}},
			}},
			{ID: 2, Name: "test_name_2", IsPublic: true, TeamID: 2,
				CreatedDate: "2022-05-25T20:30:59.39+03:00", Configuration: &Configuration{
					CustomFields: []*CustomField{},
				}},
		}

		resultProjects := ExtendProjects(projects, projectsOData)

		assert.Equal(t, expectedProjects, resultProjects)
	})
}
