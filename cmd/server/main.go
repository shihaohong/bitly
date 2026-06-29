package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/shihaohong/bitly/internal/auth"
	"github.com/shihaohong/bitly/internal/db"
	"github.com/shihaohong/bitly/internal/links"
	"github.com/shihaohong/bitly/internal/middleware"
	"github.com/shihaohong/bitly/internal/web"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	pool, err := db.New(ctx)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	authSvc := auth.NewService(pool)
	authHandler := auth.NewHandler(authSvc)

	linksSvc := links.NewService(pool)
	linksHandler := links.NewHandler(linksSvc)

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/dashboard")
	})
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	web.NewHandler().Register(r)

	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	r.GET("/:code", linksHandler.Redirect)

	api := r.Group("/api", middleware.Auth())
	{
		api.POST("/links", linksHandler.Create)
		api.GET("/links", linksHandler.List)
		api.DELETE("/links/:code", linksHandler.Delete)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
