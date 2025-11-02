package main

import (
	"log"

	"github.com/baochammm/mangahub/internal/manga"
	"github.com/baochammm/mangahub/internal/user"
	"github.com/baochammm/mangahub/package/database"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitSQLite("./data/mangahub.db")
	defer database.Close()

	mangaRepo := manga.NewRepository(database.DB)
	mangaHandler := manga.NewHandler(mangaRepo)
	userRepo := user.NewRepository(database.DB)
	userHandler := user.NewHandler(userRepo)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to MangaHub API! Available endpoints:  /manga [GET] - List all mangas,  /users [POST] - Create a new user,  /users/:user_id/reading-list [POST] - Add a reading entry for a user, /users/:user_id [GET] - Get user details with reading lists"})
	})
	// Manga routes
	r.GET("/manga", mangaHandler.GetAll)

	// User routes
	r.POST("/users", userHandler.CreateUser)
	r.POST("/users/:user_id/reading-list", userHandler.AddReadingEntry)
	r.GET("/users/:user_id", userHandler.GetUser)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("server failed:", err)
	}
}
