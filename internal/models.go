package internal

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type Project struct {
	Id       int    `json:"id"`
	TeamId   int    `json:"teamId"`
	Name     string `json:"name"`
	IsPublic bool   `json:"isPublic"`
}
