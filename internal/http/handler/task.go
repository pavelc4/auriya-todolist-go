package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pavelc4/auriya-todolist-go/internal/cache"
	db "github.com/pavelc4/auriya-todolist-go/internal/db/sqlc"
	"github.com/pavelc4/auriya-todolist-go/internal/http/repository"
)

type TaskHandler struct {
	Store *repository.Store
	cache *cache.Service
}

func NewTaskHandler(store *repository.Store, cache *cache.Service) *TaskHandler {
	return &TaskHandler{Store: store, cache: cache}
}

// newTaskResponse converts a database task model to a JSON response model.
func newTaskResponse(task db.Task) TaskResponse {
	var dueDatePtr *time.Time
	if task.DueDate.Valid {
		dueDatePtr = &task.DueDate.Time
	}

	var desc string
	if task.Description != nil {
		desc = *task.Description
	}

	var status string
	if task.Status != "" {
		status = task.Status
	}

	return TaskResponse{
		ID:          task.ID,
		UserID:      task.UserID,
		Title:       task.Title,
		Description: desc,
		Status:      status,
		Priority:    task.Priority,
		DueDate:     dueDatePtr,
		CreatedAt:   task.CreatedAt.Time,
		UpdatedAt:   task.UpdatedAt.Time,
	}
}

// @Summary      Create a new task
// @Description  Adds a new task to the user's todolist
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body      CreateTaskRequest  true  "Task to create"
// @Success      201   {object}  TaskResponse
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Security     BearerAuth
// @Router       /tasks [post]
func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	var dueDate pgtype.Timestamptz
	if req.DueDate != nil {
		dueDate = pgtype.Timestamptz{Time: *req.DueDate, Valid: true}
	}

	var projectID pgtype.Int8
	if req.ProjectID != nil {
		projectID = pgtype.Int8{Int64: *req.ProjectID, Valid: true}
	}

	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	var status *string
	if req.Status != "" {
		status = &req.Status
	}

	arg := db.CreateTaskParams{
		Title:       req.Title,
		Description: description,
		Status:      status,
		Priority:    req.Priority,
		DueDate:     dueDate,
		UserID:      userID.(int64),
		ProjectID:   projectID,
	}

	task, err := h.Store.Queries.CreateTask(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	resp := newTaskResponse(task)

	// Set cache
	cacheKey := fmt.Sprintf("task:%d", task.ID)
	h.cache.Set(cacheKey, task, 5*time.Minute)

	c.JSON(http.StatusCreated, resp)
}

func (h *TaskHandler) Get(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "detail": err.Error()})
		return
	}

	// Check cache first
	cacheKey := fmt.Sprintf("task:%d", uri.ID)
	if cached, found := h.cache.Get(cacheKey); found {
		if task, ok := cached.(db.Task); ok {
			// Verify user ID just in case
			userID, _ := c.Get("userID")
			if task.UserID == userID.(int64) {
				c.JSON(http.StatusOK, newTaskResponse(task))
				return
			}
		}
	}

	userID, _ := c.Get("userID")

	arg := db.GetTaskParams{
		ID:     uri.ID,
		UserID: userID.(int64),
	}

	task, err := h.Store.Queries.GetTask(c.Request.Context(), arg)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	// Set cache
	h.cache.Set(cacheKey, task, 5*time.Minute)

	c.JSON(http.StatusOK, newTaskResponse(task))
}

func (h *TaskHandler) List(c *gin.Context) {
	var q ListTasksQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_query", "detail": err.Error()})
		return
	}
	offset := (q.Page - 1) * q.Limit
	userID, _ := c.Get("userID")

	var dueBefore pgtype.Timestamptz
	if q.DueBefore != nil {
		dueBefore = pgtype.Timestamptz{Time: *q.DueBefore, Valid: true}
	}

	var status *string
	if q.Status != "" {
		status = &q.Status
	}

	// Caching for list endpoints is more complex, skipping for now.
	items, err := h.Store.Queries.ListTasks(c.Request.Context(), db.ListTasksParams{
		UserID:    userID.(int64),
		Status:    status,
		DueBefore: dueBefore,
		Limit:     q.Limit,
		Offset:    offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	var taskResponses []TaskResponse
	for _, item := range items {
		taskResponses = append(taskResponses, newTaskResponse(item))
	}

	c.JSON(http.StatusOK, gin.H{
		"items": taskResponses,
		"page":  q.Page,
		"limit": q.Limit,
	})
}

func (h *TaskHandler) ListByProject(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "detail": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	arg := db.ListTasksByProjectParams{
		UserID:    userID.(int64),
		ProjectID: pgtype.Int8{Int64: uri.ID, Valid: true},
	}

	tasks, err := h.Store.Queries.ListTasksByProject(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	var taskResponses []TaskResponse
	for _, task := range tasks {
		taskResponses = append(taskResponses, newTaskResponse(task))
	}

	c.JSON(http.StatusOK, taskResponses)
}

func (h *TaskHandler) Update(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "detail": err.Error()})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	var priority pgtype.Int4
	if req.Priority != nil {
		priority = pgtype.Int4{Int32: *req.Priority, Valid: true}
	}
	var dueDate pgtype.Timestamptz
	if req.DueDate != nil {
		dueDate = pgtype.Timestamptz{Time: *req.DueDate, Valid: true}
	}
	var projectID pgtype.Int8
	if req.ProjectID != nil {
		projectID = pgtype.Int8{Int64: *req.ProjectID, Valid: true}
	}

	arg := db.UpdateTaskParams{
		ID:          uri.ID,
		UserID:      userID.(int64),
		Title:       toPgText(req.Title),
		Description: req.Description,
		Status:      toPgText(req.Status),
		Priority:    priority,
		DueDate:     dueDate,
		ProjectID:   projectID,
	}
	task, err := h.Store.Queries.UpdateTask(c.Request.Context(), arg)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("task:%d", uri.ID)
	h.cache.Delete(cacheKey)

	c.JSON(http.StatusOK, newTaskResponse(task))
}

func (h *TaskHandler) Delete(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "detail": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	arg := db.DeleteTaskParams{
		ID:     uri.ID,
		UserID: userID.(int64),
	}

	if err := h.Store.Queries.DeleteTask(c.Request.Context(), arg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	// Invalidate cache
	cacheKey := fmt.Sprintf("task:%d", uri.ID)
	h.cache.Delete(cacheKey)

	c.Status(http.StatusNoContent)
}

func toPgText(s *string) pgtype.Text {
	if s != nil {
		return pgtype.Text{String: *s, Valid: true}
	}
	return pgtype.Text{Valid: false}
}
