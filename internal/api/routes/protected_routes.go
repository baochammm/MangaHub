package routes

import (
	"github.com/baochammm/mangahub/internal/api/middleware"
	"github.com/baochammm/mangahub/internal/udp"
	"github.com/baochammm/mangahub/internal/user"
	"github.com/baochammm/mangahub/package/database"
	"github.com/gin-gonic/gin"
)

func RegisterProtectedRoutes(router *gin.Engine) {

	router.Use(middleware.AuthMiddleware())
	userRepo := user.NewRepository(database.DB)
	userHandler := user.NewHandler(userRepo)
	udpRepo := udp.NewUDPRepository(database.DB)
	udpHandler := udp.NewUDPHandler(udpRepo)
	router.POST("/users/library", userHandler.AddReadingEntry)                // mangahub library add
	router.PATCH("/users/library", userHandler.UpdateReadingStatus)           // mangahub library update --manga-id <id> --status <new-status>
	router.PATCH("/users/progress", userHandler.UpdateReadingProgress)        // mangahub progress update --manga-id <id> --current-chapter <chapter>
	router.DELETE("/users/library", userHandler.DeleteReadingEntry)           // mangahub library remove --manga-id <id>
	router.GET("/users/library", userHandler.GetUserLibrary)                  // mangahub library list
	router.GET("/users/library/:status", userHandler.GetUserLibraryViaStatus) // mangahub library list --status=<status>
	router.POST("/users/notifications/subscribe", udpHandler.ProcessUDPAddress)
	router.POST("/users/notifications/subscribe/:manga", udpHandler.SubscribeToManga)

}
