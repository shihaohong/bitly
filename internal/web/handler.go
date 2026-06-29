package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed dist
var distFS embed.FS

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

func (h *Handler) Register(r *gin.Engine) {
	sub, _ := fs.Sub(distFS, "dist")
	fileServer := http.FileServer(http.FS(sub))

	r.GET("/assets/*filepath", func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	})

	spa := func(c *gin.Context) {
		data, _ := distFS.ReadFile("dist/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", data)
	}
	r.GET("/login", spa)
	r.GET("/register", spa)
	r.GET("/dashboard", spa)
}
