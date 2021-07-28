package internal

type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type User struct {
	ID            int    `json:"id"`
	UserName      string `json:"userName"`
	LastLoginDate string `json:"lastLoginDate"`
	RoleIds       []int  `json:"roleIds"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Email         string `json:"email"`
}

type Role struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type Project struct {
	ID       int    `json:"id"`
	TeamID   int    `json:"teamId"`
	Name     string `json:"name"`
	IsPublic bool   `json:"isPublic"`
}
