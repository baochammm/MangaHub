package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/baochammm/mangahub/package/database"
	"github.com/baochammm/mangahub/package/models"
)

func main() {
	dbPath := "data/mangahub.db"
	database.InitSQLite(dbPath)

	populateManga()
	populateUsers()
	fmt.Println("Database created and populated successfully.")
}

func populateManga() {
	file, err := os.ReadFile("data/json/manga.json")
	if err != nil {
		log.Fatalf("failed to read manga.json: %v", err)
	}

	var mangas []models.Manga
	if err := json.Unmarshal(file, &mangas); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	for _, m := range mangas {
		genresJSON, _ := json.Marshal(m.Genres) // store array as JSON string

		_, err := database.DB.Exec(`
			INSERT INTO mangas 
			(id, title, author, artist, genres, chapter_count, published_year, status, cover_url, description)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			m.ID, m.Title, m.Author, m.Artist, string(genresJSON), m.ChapterCount,
			m.PublishedYear, m.Status, m.CoverURL, m.Description,
		)
		if err != nil {
			log.Fatalf("failed to insert manga %s: %v", m.ID, err)
		}
	}
	log.Printf("Inserted %d manga records.", len(mangas))
}

func populateUsers() {
	file, err := os.ReadFile("data/json/user.json")
	if err != nil {
		log.Fatalf("failed to read user.json: %v", err)
	}

	var users []models.User
	if err := json.Unmarshal(file, &users); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}
	for _, m := range users {

		_, err := database.DB.Exec(`
			INSERT INTO users 
			(username, password_hash)
			VALUES (?, ?)`,
			m.Username, m.PasswordHash,
		)
		if err != nil {
			log.Fatalf("failed to insert user %s: %v", m.Username, err)
		}
	}
}
