package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pavelc4/auriya-todolist-go/internal/http/handler"
)

func New(db *pgxpool.Pool) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.Default())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello World",
			"status":  "OK",
		})
	})

	health := handler.NewHealthHandler(db)
	r.GET("/health", health.Health)

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"message": "Ngapain Bego",
			"error":   "Route not found",
		})
	})

	return r
}
