package manga

import (
	"net/http"
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
	mangas, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mangas)
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

	mangas, err := h.repo.SearchByTitle(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mangas)
}

// filter manga by genre
func (h *Handler) FilterByGenre(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "need query parameter"})
		return
	}

	genres := strings.Split(query, ",")
	for i := range genres {
		genres[i] = strings.TrimSpace(strings.ToLower(genres[i]))
	}

	mangas, err := h.repo.FilterByGenre(genres)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mangas)
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
