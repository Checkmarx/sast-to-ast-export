package rest

import "time"

type Retry struct {
	Attempts int
	MinSleep,
	MaxSleep time.Duration
}

// ExtendProjects adds custom fields and created date to projects
func ExtendProjects(projects []*Project, projectsOData []*ProjectOData) []*Project {
	for indexProject, project := range projects {
		// arrays have to have the same order
		if len(projectsOData) > indexProject && project.ID == projectsOData[indexProject].ID {
			addDateAndCustomFields(project, projectsOData[indexProject])
			continue
		}

		// if we have wrong order in OData array
		for _, projectOData := range projectsOData {
			if project.ID == projectOData.ID {
				addDateAndCustomFields(project, projectOData)
				break
			}
		}
	}

	return projects
}

// addDateAndCustomFields adds created date and custom fields in project
func addDateAndCustomFields(project *Project, projectOData *ProjectOData) {
	project.Configuration = &Configuration{
		CustomFields: projectOData.CustomFields,
	}
	project.CreatedDate = projectOData.CreatedDate
}
