package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/pavelc4/auriya-todolist-go/internal/db/sqlc"
	"github.com/pavelc4/auriya-todolist-go/internal/http/repository"
)

type TaskHandler struct {
	Store *repository.Store
}

func NewTaskHandler(store *repository.Store) *TaskHandler {
	return &TaskHandler{Store: store}
}

func (h *TaskHandler) Create(c *gin.Context) {
	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}

	// nilai default jika kosong
	status := "pending"
	if req.Status != nil {
		status = *req.Status
	}
	priority := int32(1)
	if req.Priority != nil {
		priority = *req.Priority
	}
	var due time.Time
	if req.DueDate != nil {
		due = *req.DueDate
	}

	arg := db.CreateTaskParams{
		Title:       req.Title,                 // string
		Description: sval(req.Description, ""), // string
		Status:      status,                    // string
		Priority:    priority,                  // int32
		DueDate:     due,                       // time.Time
	}

	task, err := h.Store.Queries.CreateTask(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, task)
}

// helper konversi pointer -> nilai
func sval(p *string, def string) string {
	if p == nil {
		return def
	}
	return *p
}
func ival(p *int32, def int32) int32 {
	if p == nil {
		return def
	}
	return *p
}
func tval(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}

func (h *TaskHandler) Get(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "detail": err.Error()})
		return
	}
	task, err := h.Store.Queries.GetTask(c.Request.Context(), uri.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
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

	// ubah pointer -> nilai untuk Params yang butuh value
	var status string
	if q.Status != nil {
		status = *q.Status
	}
	var dueBefore time.Time
	if q.DueBefore != nil {
		dueBefore = *q.DueBefore
	}

	items, err := h.Store.Queries.ListTasks(c.Request.Context(), db.ListTasksParams{
		Status:    status,    // string
		DueBefore: dueBefore, // time.Time
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

	arg := db.UpdateTaskParams{
		ID:          uri.ID,
		Title:       sval(req.Title, ""),       // string
		Description: sval(req.Description, ""), // string
		Status:      sval(req.Status, ""),      // string
		Priority:    ival(req.Priority, 0),     // int32
		DueDate:     tval(req.DueDate),         // time.Time
	}
	task, err := h.Store.Queries.UpdateTask(c.Request.Context(), arg)
	if err != nil {
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
	if err := h.Store.Queries.DeleteTask(c.Request.Context(), uri.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
