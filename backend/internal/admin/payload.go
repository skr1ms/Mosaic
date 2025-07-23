package admin

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CreateAdminRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}