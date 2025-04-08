package export

const (
	// UsersOption represent users
	UsersOption = "users"
	// TeamsOption represent teams (groups)
	TeamsOption = "teams"
	// ResultsOption represent results (projects & data)
	ResultsOption = "triage"
	// ProjectsOption represent projects
	ProjectsOption = "projects"
	// QueriesOption represent custom queries
	QueriesOption = "queries"
	// PresetsOption represent presets
	PresetsOption = "presets"
	// EngineConfigurations represent the configurations engine scan can have
	EngineConfigurationsOption = "configs"
	// Filters and exclude settings
	FiltersOption = "filters"
)

func GetOptions() []string {
	return []string{UsersOption, TeamsOption, ResultsOption, ProjectsOption, QueriesOption, PresetsOption, EngineConfigurationsOption, FiltersOption}
}
