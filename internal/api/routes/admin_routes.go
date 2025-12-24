package routes

import (
	"github.com/baochammm/mangahub/internal/api/middleware"
	"github.com/baochammm/mangahub/internal/manga"
	"github.com/baochammm/mangahub/internal/udp"
	"github.com/baochammm/mangahub/package/database"
	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(router *gin.Engine) {
	// Apply both auth and admin middleware
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware())
	adminGroup.Use(middleware.AdminMiddleware())

	mangaRepo := manga.NewRepository(database.DB)
	udpRepo := udp.NewUDPRepository(database.DB)
	udpHandler := udp.NewUDPHandler(udpRepo)
	mangaHandler := manga.NewHandler(mangaRepo, udpHandler)

	// Admin-only routes
	adminGroup.PUT("/manga", mangaHandler.UpdateManga) // create manga endpoint /localhost:8080/admin/manga
	adminGroup.PUT("/manga/chapter-release", mangaHandler.UpdateMangaChapterRelease)

}
