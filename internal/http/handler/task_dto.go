package handler

import "time"

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description" binding:"omitempty"`
	Status      *string    `json:"status" binding:"omitempty,oneof=pending in_progress done"`
	Priority    *int32     `json:"priority" binding:"omitempty,min=1,max=5"`
	DueDate     *time.Time `json:"due_date" binding:"omitempty"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title" binding:"omitempty"`
	Description *string    `json:"description" binding:"omitempty"`
	Status      *string    `json:"status" binding:"omitempty,oneof=pending in_progress done"`
	Priority    *int32     `json:"priority" binding:"omitempty,min=1,max=5"`
	DueDate     *time.Time `json:"due_date" binding:"omitempty"`
}

type ListTasksRequest struct {
	Status    *string    `form:"status" binding:"omitempty,oneof=pending in_progress done"`
	DueBefore *time.Time `form:"due_before" binding:"omitempty"`
	Limit     int32      `form:"limit,default=20" binding:"min=1,max=100"`
	Page      int32      `form:"page,default=1" binding:"min=1"`
}
