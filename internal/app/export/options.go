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
	// PresetsOption Represent presets
	PresetsOption = "presets"
)

func GetOptions() []string {
	return []string{UsersOption, TeamsOption, ResultsOption, ProjectsOption, QueriesOption, PresetsOption}
}
