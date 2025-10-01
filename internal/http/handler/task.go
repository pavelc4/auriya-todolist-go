package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/pavelc4/auriya-todolist-go/internal/db/sqlc"
	"github.com/pavelc4/auriya-todolist-go/internal/http/repository"
)

type TaskHandler struct {
	Store *repository.Store
}

func NewTaskHandler(store *repository.Store) *TaskHandler {
	return &TaskHandler{Store: store}
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

	arg := db.CreateTaskParams{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		DueDate:     dueDate,
		UserID:      userID.(int64),
	}

	task, err := h.Store.Queries.CreateTask(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	var dueDatePtr *time.Time
	if task.DueDate.Valid {
		dueDatePtr = &task.DueDate.Time
	}

	resp := TaskResponse{
		ID:          task.ID,
		UserID:      task.UserID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Priority:    task.Priority,
		DueDate:     dueDatePtr,
		CreatedAt:   task.CreatedAt.Time,
		UpdatedAt:   task.UpdatedAt.Time,
	}

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
	c.JSON(http.StatusOK, task)
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

	items, err := h.Store.Queries.ListTasks(c.Request.Context(), db.ListTasksParams{
		UserID:    userID.(int64),
		Status:    q.Status,
		DueBefore: dueBefore,
		Limit:     q.Limit,
		Offset:    offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"page":  q.Page,
		"limit": q.Limit,
	})
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

	var title pgtype.Text
	if req.Title != nil {
		title = pgtype.Text{String: *req.Title, Valid: true}
	}
	var status pgtype.Text
	if req.Status != nil {
		status = pgtype.Text{String: *req.Status, Valid: true}
	}
	var priority pgtype.Int4
	if req.Priority != nil {
		priority = pgtype.Int4{Int32: *req.Priority, Valid: true}
	}
	var dueDate pgtype.Timestamptz
	if req.DueDate != nil {
		dueDate = pgtype.Timestamptz{Time: *req.DueDate, Valid: true}
	}

	arg := db.UpdateTaskParams{
		ID:          uri.ID,
		UserID:      userID.(int64),
		Title:       title,
		Description: req.Description,
		Status:      status,
		Priority:    priority,
		DueDate:     dueDate,
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
	c.JSON(http.StatusOK, task)
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
	c.Status(http.StatusNoContent)
}
