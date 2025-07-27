package admin

type CreateAdminRequest struct {
	Login    string `json:"login" validate:"required,secure_login"`
	Password string `json:"password" validate:"required,secure_password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,secure_password"`
}
