package export

type (
	MetadataRecord struct {
		QueryID      string `json:"queryId"`
		ResultID     string `json:"resultId"`
		PathID       string `json:"pathId"`
		SimilarityID string `json:"similarityId"`
	}

	MetadataQuery struct {
		QueryID  string
		Language string
		Name     string
		Group    string
	}

	MetadataResult struct {
		PathID    string
		ResultID  string
		FirstNode MetadataNode
		LastNode  MetadataNode
	}

	MetadataNode struct {
		FileName string
		Name     string
		Line     string
		Column   string
	}
)
