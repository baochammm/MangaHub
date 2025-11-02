package main

import (
	"log"

	"github.com/baochammm/mangahub/internal/manga"
	"github.com/baochammm/mangahub/package/database"
	"github.com/gin-gonic/gin"
)

func main() {
	database.InitSQLite("./data/manga.db")
	defer database.Close()

	repo := manga.NewRepository(database.DB)
	handler := manga.NewHandler(repo)

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to MangaHub API!"})
	})
	// Manga routes
	r.GET("/manga", handler.GetAll)

	if err := r.Run(":8080"); err != nil {
		log.Fatal("server failed:", err)
	}
}
