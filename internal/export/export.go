package export

const (
	UsersOption   = "users"
	TeamsOption   = "teams"
	ResultsOption = "results"
)

func GetOptions() []string {
	return []string{UsersOption, TeamsOption, ResultsOption}
}
