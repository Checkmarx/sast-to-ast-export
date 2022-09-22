package querymapping

type (
	QueryMap struct {
		AstID  string `json:"astId"`
		SastID string `json:"sastId"`
	}

	MapSource struct {
		Mappings []QueryMap `json:"mappings"`
	}
)
