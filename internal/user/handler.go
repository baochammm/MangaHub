package user

import (
	"net/http"
	"time"

	"github.com/baochammm/mangahub/package/models"
	"github.com/baochammm/mangahub/utils"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) AddReadingEntry(c *gin.Context) {
	userID, err := utils.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var entry models.ReadingEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if entry.MangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga_id is required"})
		return
	}

	// check if manga exists
	exists, err := h.repo.MangaExists(entry.MangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga_id does not exist"})
		return
	}

	// check duplicate entry
	existsRL, err := h.repo.ReadingEntryExists(userID, entry.MangaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "database error"})
		return
	}
	if existsRL {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga already added to library"})
		return
	}

	entry.LastUpdated = time.Now()

	if err := h.repo.AddReadingEntry(userID, entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "reading entry added"})
}
func (h *Handler) UpdateReadingStatus(c *gin.Context) {
	userID, err := utils.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var entry models.ReadingEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if entry.MangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga_id is required"})
		return
	}
	entry.LastUpdated = time.Now()

	var mangaExists bool
	mangaExists, err = h.repo.IsMangaInUserLibrary(userID, entry.MangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !mangaExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "manga not found in user library"})
		return
	}
	if entry.Status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be provided"})
		return
	}

	if err := h.repo.UpdateReadingStatus(userID, entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reading entry updated"})
}
func (h *Handler) UpdateReadingProgress(c *gin.Context) {
	userID, err := utils.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var entry models.ReadingEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if entry.CurrentChapter == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current_chapter is required"})
		return
	}
	if entry.MangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga_id is required"})
		return
	}
	entry.LastUpdated = time.Now()

	var mangaExists bool
	mangaExists, err = h.repo.IsMangaInUserLibrary(userID, entry.MangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !mangaExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "manga not found in user library"})
		return
	}
	if err := h.repo.UpdateReadingProgress(userID, entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reading entry for manga " + entry.MangaID + " updated"})
}
func (h *Handler) DeleteReadingEntry(c *gin.Context) {
	userID, err := utils.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var entry models.ReadingEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if entry.MangaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "manga_id is required"})
		return
	}
	entry.LastUpdated = time.Now()

	var mangaExists bool
	mangaExists, err = h.repo.IsMangaInUserLibrary(userID, entry.MangaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !mangaExists {
		c.JSON(http.StatusNotFound, gin.H{"error": "manga not found in user library"})
		return
	}
	if err := h.repo.DeleteReadingEntry(userID, entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "manga entry " + entry.MangaID + " deleted"})
}
func (h *Handler) GetUserLibrary(c *gin.Context) {
	userID, err := utils.GetUserIdFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	user, err := h.repo.GetUserReadingLists(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}
func (h *Handler) GetUserLibraryViaStatus(c *gin.Context) {
	userID, err := utils.GetUserIdFromContext(c)
	status := c.Param("status")

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	readingLists, err := h.repo.GetUserReadingListsViaStatus(userID, status)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "reading lists not found"})
		return
	}
	c.JSON(http.StatusOK, readingLists)
}

//TODO: Update reading entry -> change status
