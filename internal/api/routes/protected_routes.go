package routes

import (
	"github.com/baochammm/mangahub/internal/api/middleware"
	"github.com/baochammm/mangahub/internal/manga"
	"github.com/baochammm/mangahub/internal/user"
	"github.com/baochammm/mangahub/package/database"
	"github.com/gin-gonic/gin"
)

func RegisterProtectedRoutes(router *gin.Engine) {
	router.Use(middleware.AuthMiddleware())
	userRepo := user.NewRepository(database.DB)
	userHandler := user.NewHandler(userRepo)
	router.POST("/users/library", userHandler.AddReadingEntry)
	router.GET("/users/library", userHandler.GetUserLibrary)
	mangaRepo := manga.NewRepository(database.DB)
	mangaHandler := manga.NewHandler(mangaRepo)
	router.GET("/manga/:id", mangaHandler.GetByID)
}
