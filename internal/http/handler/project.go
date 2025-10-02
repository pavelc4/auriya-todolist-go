package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/pavelc4/auriya-todolist-go/internal/db/sqlc"
	"github.com/pavelc4/auriya-todolist-go/internal/http/repository"
)

type ProjectHandler struct {
	Store *repository.Store
}

func NewProjectHandler(store *repository.Store) *ProjectHandler {
	return &ProjectHandler{Store: store}
}

// newProjectResponse converts a database project model to a JSON response model.
func newProjectResponse(project db.Project) ProjectResponse {
	return ProjectResponse{
		ID:        project.ID,
		UserID:    project.UserID,
		Name:      project.Name,
		CreatedAt: project.CreatedAt.Time,
		UpdatedAt: project.UpdatedAt.Time,
	}
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	arg := db.CreateProjectParams{
		UserID: userID.(int64),
		Name:   req.Name,
	}

	project, err := h.Store.Queries.CreateProject(c.Request.Context(), arg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newProjectResponse(project))
}

func (h *ProjectHandler) Get(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "detail": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	arg := db.GetProjectParams{
		ID:     uri.ID,
		UserID: userID.(int64),
	}

	project, err := h.Store.Queries.GetProject(c.Request.Context(), arg)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, newProjectResponse(project))
}

func (h *ProjectHandler) List(c *gin.Context) {
	userID, _ := c.Get("userID")

	projects, err := h.Store.Queries.ListProjects(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}

	var projectResponses []ProjectResponse
	for _, p := range projects {
		projectResponses = append(projectResponses, newProjectResponse(p))
	}

	c.JSON(http.StatusOK, projectResponses)
}

func (h *ProjectHandler) Update(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "detail": err.Error()})
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	arg := db.UpdateProjectParams{
		ID:     uri.ID,
		UserID: userID.(int64),
		Name:   req.Name,
	}

	project, err := h.Store.Queries.UpdateProject(c.Request.Context(), arg)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, newProjectResponse(project))
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required,min=1"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_id", "detail": err.Error()})
		return
	}

	userID, _ := c.Get("userID")

	arg := db.DeleteProjectParams{
		ID:     uri.ID,
		UserID: userID.(int64),
	}

	if err := h.Store.Queries.DeleteProject(c.Request.Context(), arg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db_error", "detail": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
