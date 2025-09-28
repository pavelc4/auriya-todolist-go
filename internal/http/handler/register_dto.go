package handler

type RegisterUserRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
	FullName        string `json:"full_name" binding:"omitempty"`
	Age             int    `json:"age" binding:"required,min=13"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}