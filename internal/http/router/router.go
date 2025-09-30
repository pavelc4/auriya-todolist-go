package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pavelc4/auriya-todolist-go/internal/http/handler"
	"github.com/pavelc4/auriya-todolist-go/internal/http/repository"
	"github.com/pavelc4/auriya-todolist-go/internal/http/service"
	"golang.org/x/oauth2"
)

func New(db *pgxpool.Pool, googleConf, githubConf *oauth2.Config, userRepo *repository.UserRepository, jwtService *service.JWTService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
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

	store := repository.NewStore(db)
	task := handler.NewTaskHandler(store)
	auth := handler.NewAuthHandler(googleConf, githubConf, userRepo, jwtService)

	// auth routes
	// Google
	r.GET("/auth/google/login", auth.GoogleLogin)
	r.GET("/auth/google/callback", auth.GoogleCallback)
	// Github
	r.GET("/auth/github/login", auth.GitHubLogin)
	r.GET("/auth/github/callback", auth.GitHubCallback)

	api := r.Group("/api")
	{
		api.POST("/register", auth.Register)
		api.POST("/login", auth.Login)

		api.POST("/tasks", task.Create)
		api.GET("/tasks/:id", task.Get)
		api.GET("/tasks", task.List)
		api.PATCH("/tasks/:id", task.Update)
		api.DELETE("/tasks/:id", task.Delete)
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"message": "Ngapain Bego",
			"error":   "Route not found",
		})
	})

	return r
}
