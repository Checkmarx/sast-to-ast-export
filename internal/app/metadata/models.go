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

	SimilarityCalculationJob struct {
		Filename1, Name1, Line1, Column1, MethodLine1,
		Filename2, Name2, Line2, Column2, MethodLine2,
		QueryID string
	}

	SimilarityCalculationResult struct {
		Err          error
		SimilarityID string
	}
)
