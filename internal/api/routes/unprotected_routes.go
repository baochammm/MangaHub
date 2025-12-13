package routes

import (
	"github.com/baochammm/mangahub/internal/auth"
	"github.com/baochammm/mangahub/internal/manga"
	"github.com/baochammm/mangahub/internal/udp"
	"github.com/baochammm/mangahub/package/database"
	"github.com/gin-gonic/gin"
)

func RegisterUnProtectedRoutes(router *gin.Engine) {
	mangaRepo := manga.NewRepository(database.DB)
	udpHandler := udp.NewUDPHandler(udp.NewUDPRepository(database.DB))
	mangaHandler := manga.NewHandler(mangaRepo, udpHandler)

	authRepo := auth.NewAuthRepository(database.DB)
	authHandler := auth.NewAuthHandler(authRepo)
	router.GET("/manga", mangaHandler.GetAll)
	router.POST("/auth/signup", authHandler.Signup)
	router.POST("/auth/login", authHandler.Login)
	router.POST("/auth/logout", authHandler.Logout)
	router.GET("/manga/search", mangaHandler.SearchByTitle)       // search manga by title
	router.GET("/manga/filter/genre", mangaHandler.FilterByGenre) // filter manga by genre
	router.GET("/manga/:id", mangaHandler.GetByID)

	router.PUT("/manga", mangaHandler.UpdateManga) // create manga

}
