package handler

import "time"

type CreateProjectRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

type UpdateProjectRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

type ProjectResponse struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
