package links

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/shihaohong/bitly/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	var req struct {
		URL string `json:"url" binding:"required,url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	link, err := h.svc.Create(c.Request.Context(), userID, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create link"})
		return
	}

	c.JSON(http.StatusCreated, link)
}

func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")

	url, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
		return
	}

	c.Redirect(http.StatusMovedPermanently, url)
}

func (h *Handler) List(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)

	links, err := h.svc.List(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list links"})
		return
	}
	if links == nil {
		links = []Link{}
	}

	c.JSON(http.StatusOK, links)
}

func (h *Handler) Delete(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	code := c.Param("code")

	if err := h.svc.Delete(c.Request.Context(), userID, code); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete link"})
		return
	}

	c.Status(http.StatusNoContent)
}
