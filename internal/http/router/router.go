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
	r.Use(cors.Default()) // untuk dev; produksi pakai config ketat

	health := handler.NewHealthHandler(db)
	r.GET("/health", health.Health)

	return r
}
