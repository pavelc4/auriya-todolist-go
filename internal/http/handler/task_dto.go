package handler

import "time"

// CreateTaskRequest defines the request body for creating a new task.
type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required,max=255"`
	Description string     `json:"description"`
	Status      string     `json:"status" binding:"omitempty,oneof=pending in-progress completed"`
	Priority    int32      `json:"priority" binding:"omitempty,min=1,max=5"`
	DueDate     *time.Time `json:"due_date"`
	ProjectID   *int64     `json:"project_id" binding:"omitempty,min=1"`
}

// UpdateTaskRequest defines the request body for updating a task.
type UpdateTaskRequest struct {
	Title       *string    `json:"title" binding:"omitempty,max=255"`
	Description *string    `json:"description"`
	Status      *string    `json:"status" binding:"omitempty,oneof=pending in-progress completed"`
	Priority    *int32     `json:"priority" binding:"omitempty,min=1,max=5"`
	DueDate     *time.Time `json:"due_date"`
	ProjectID   *int64     `json:"project_id" binding:"omitempty,min=1"`
}

// TaskResponse defines the standard response for a task.
type TaskResponse struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    int32      `json:"priority"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ListTasksQuery defines the query parameters for listing tasks.
type ListTasksQuery struct {
	Page      int32      `form:"page,default=1"`
	Limit     int32      `form:"limit,default=10"`
	Status    string     `form:"status"`
	DueBefore *time.Time `form:"due_before"`
}
