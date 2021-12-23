package metadata

type (
	Record struct {
		QueryID      string `json:"queryId"`
		ResultID     string `json:"resultId"`
		PathID       string `json:"pathId"`
		SimilarityID string `json:"similarityId"`
	}

	Query struct {
		QueryID  string
		Language string
		Name     string
		Group    string
		Results  []*Result
	}

	Result struct {
		PathID    string
		ResultID  string
		FirstNode Node
		LastNode  Node
	}

	Node struct {
		FileName string
		Name     string
		Line     string
		Column   string
	}
)
