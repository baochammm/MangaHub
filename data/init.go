package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/baochammm/mangahub/package/database"
	"github.com/baochammm/mangahub/package/models"
)

func main() {
	dbPath := "data/mangahub.db"
	database.InitSQLite(dbPath)

	// populateManga()
	// // populateUsers()
	// fmt.Println("Database created and populated successfully.")
	err := ExportMangaToJSON(database.DB, "mangas_export.json")
	if err != nil {
		log.Fatal(err)
	}

}

type MangaInJSON struct {
	Name       string `json:"name"`
	Title      string `json:"title"`
	StartDate  string `json:"start_date"`
	Status     string `json:"status"`
	Synopsis   string `json:"synopsis"`
	Popularity *int   `json:"popularity"`
	Rank       *int   `json:"rank"`

	MainPic struct {
		Large string `json:"large"`
	} `json:"main_picture"`

	Genres []struct {
		Name string `json:"name"`
	} `json:"genres"`
	ChapterCount *int `json:"num_chapters"`
	VolumeCount  *int `json:"num_volumes"`

	Authors []Author `json:"authors"`
}
type MangaExport struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Author        string   `json:"author"`
	Artist        string   `json:"artist"`
	Genres        []string `json:"genres"`
	ChapterCount  int      `json:"chapter_count"`
	VolumeCount   int      `json:"volume_count"`
	PublishedYear int      `json:"published_year"`
	Status        string   `json:"status"`
	CoverURL      string   `json:"cover_url"`
	Description   string   `json:"description"`
	Popularity    int      `json:"popularity"`
	Ranking       int      `json:"ranking"`
}

type Author struct {
	Role string `json:"role"`
	Node struct {
		First string `json:"first_name"`
		Last  string `json:"last_name"`
	} `json:"node"`
}

func joinAuthors(authors []Author, keyword string) *string {
	var names []string

	for _, a := range authors {
		if strings.Contains(a.Role, keyword) {
			full := strings.TrimSpace(a.Node.First + " " + a.Node.Last)
			if full != "" {
				names = append(names, full)
			}
		}
	}

	if len(names) == 0 {
		return nil
	}

	result := strings.Join(names, ", ")
	return &result
}

func extractYear(date string) *int {
	if len(date) < 4 {
		return nil
	}
	y, err := strconv.Atoi(date[:4])
	if err != nil {
		return nil
	}
	return &y
}
func populateManga() {
	file, err := os.ReadFile("data/json/top50pop_manga.json")
	if err != nil {
		log.Fatalf("failed to read manga.json: %v", err)
	}

	var mangas []MangaInJSON
	if err := json.Unmarshal(file, &mangas); err != nil {
		log.Fatalf("failed to unmarshal JSON: %v", err)
	}

	// for _, m := range mangas {
	// 	genresJSON, _ := json.Marshal(m.Genres) // store array as JSON string
	stmt := `
INSERT INTO mangas (
  id,
  title,
  author,
  artist,
  genres,
  chapter_count,
  volume_count,
  published_year,
  status,
  cover_url,
  description,
  popularity,
  ranking
) VALUES (
  @name,
  @title,
  @author,
  @artist,
  @genres,
  @chapter_count,
  @volume_count,
  @published_year,
  @status,
  @cover_url,
  @description,
  @popularity,
  @ranking
)
ON CONFLICT(id) DO UPDATE SET
  title          = COALESCE(mangas.title, excluded.title),
  author         = COALESCE(mangas.author, excluded.author),
  artist         = COALESCE(mangas.artist, excluded.artist),
  genres         = COALESCE(mangas.genres, excluded.genres),
  chapter_count  = COALESCE(mangas.chapter_count, excluded.chapter_count),
  volume_count   = COALESCE(mangas.volume_count, excluded.volume_count),
  published_year = COALESCE(mangas.published_year, excluded.published_year),
  status         = COALESCE(mangas.status, excluded.status),
  cover_url      = COALESCE(mangas.cover_url, excluded.cover_url),
  description    = COALESCE(mangas.description, excluded.description),
  popularity     = COALESCE(mangas.popularity, excluded.popularity),
  ranking        = COALESCE(mangas.ranking, excluded.ranking);

`

	for _, m := range mangas {
		genresJSON, _ := json.Marshal(func() []string {
			var g []string
			for _, genre := range m.Genres {
				g = append(g, genre.Name)
			}
			return g
		}())

		_, err := database.DB.Exec(
			stmt,
			sql.Named("name", m.Name), // 🔥 ID = name
			sql.Named("title", m.Title),
			sql.Named("author", joinAuthors(m.Authors, "Story")),
			sql.Named("artist", joinAuthors(m.Authors, "Art")),
			sql.Named("genres", string(genresJSON)),
			sql.Named("chapter_count", m.ChapterCount),
			sql.Named("volume_count", m.VolumeCount),
			sql.Named("published_year", extractYear(m.StartDate)),
			sql.Named("status", m.Status),
			sql.Named("cover_url", m.MainPic.Large),
			sql.Named("description", m.Synopsis),
			sql.Named("popularity", m.Popularity),
			sql.Named("ranking", m.Rank),
		)

		if err != nil {
			log.Printf("❌ Failed to import %s: %v", m.Name, err)
		}

	}
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
func ExportMangaToJSON(db *sql.DB, outFile string) error {
	rows, err := db.Query(`
		SELECT
			id,
			title,
			author,
			artist,
			genres,
			chapter_count,
			volume_count,
			published_year,
			status,
			cover_url,
			description,
			popularity,
			ranking
		FROM mangas
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var result []MangaExport

	for rows.Next() {
		var (
			m      MangaExport
			rawGen string

			artist sql.NullString
		)

		err := rows.Scan(
			&m.ID,
			&m.Title,
			&m.Author,
			&artist,
			&rawGen, // 👈 JSON TEXT
			&m.ChapterCount,
			&m.VolumeCount,
			&m.PublishedYear,
			&m.Status,
			&m.CoverURL,
			&m.Description,
			&m.Popularity,
			&m.Ranking,
		)
		if err != nil {
			return err
		}
		m.Artist = artist.String

		// 🔥 Decode JSON string into []string
		if err := json.Unmarshal([]byte(rawGen), &m.Genres); err != nil {
			return fmt.Errorf("invalid genres JSON for %s: %v", m.ID, err)
		}

		result = append(result, m)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outFile, data, 0644)
}
