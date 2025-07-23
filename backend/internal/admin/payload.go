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

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
