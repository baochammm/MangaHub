package manga

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/baochammm/mangahub/internal/udp"
	"github.com/baochammm/mangahub/package/models"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo       *Repository
	udpHandler *udp.UDPHandler
}

func NewHandler(
	repo *Repository,
	udpHandler *udp.UDPHandler,
) *Handler {
	return &Handler{
		repo:       repo,
		udpHandler: udpHandler,
	}
}

func (h *Handler) GetAll(c *gin.Context) {
	page := c.Query("page")
	pageSize := c.Query("page_size")

	if page == "" || page == "0" {
		page = "1"
	}
	if pageSize == "" || pageSize == "0" {
		pageSize = "20"
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})
		return
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size parameter"})
		return
	}
	result, err := h.repo.GetAll(pageInt, pageSizeInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"page":        pageInt,
		"page_size":   pageSizeInt,
		"total_pages": result.TotalPages,
		"total_items": result.TotalItems,
		"items":       result.Items,
	})
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	manga, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, manga)
}

// search manga by title
func (h *Handler) SearchByTitle(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "need query parameter"})
		return
	}

	mangas, err := h.repo.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mangas)
}

// filter manga by genre
func (h *Handler) FilterByGenre(c *gin.Context) {
	page := c.Query("page")
	pageSize := c.Query("page_size")

	if page == "" || page == "0" {
		page = "1"
	}
	if pageSize == "" || pageSize == "0" {
		pageSize = "20"
	}
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "need query parameter"})
		return
	}

	genres := strings.Split(query, ",")
	for i := range genres {
		genres[i] = strings.TrimSpace(strings.ToLower(genres[i]))
	}
	pageInt, err := strconv.Atoi(page)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})
		return
	}
	pageSizeInt, err := strconv.Atoi(pageSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_size parameter"})
		return
	}
	result, err := h.repo.FilterByGenre(genres, pageInt, pageSizeInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"page":        pageInt,
		"page_size":   pageSizeInt,
		"total_pages": result.TotalPages,
		"total_items": result.TotalItems,
		"items":       result.Items,
	})
}

// special admin handler to update manga database
func (h *Handler) UpdateManga(c *gin.Context) {
	var m models.Manga
	if err := c.ShouldBindJSON(&m); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	success, err := h.repo.UpdateManga(m)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": success,
		"message": "Manga database updated successfully"})
	h.udpHandler.NotifyNewChapter(m.ID, int64(m.ChapterCount))
}
