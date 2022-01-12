package export

const (
	// UsersOption represent users
	UsersOption = "users"
	// TeamsOption represent teams (groups)
	TeamsOption = "teams"
	// ResultsOption represent results (projects & data)
	ResultsOption = "triaged"
)

func GetOptions() []string {
	return []string{UsersOption, TeamsOption, ResultsOption}
}
