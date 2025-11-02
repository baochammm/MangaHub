package manga

import (
	"database/sql"
	"encoding/json"

	"github.com/baochammm/mangahub/package/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetAll() ([]models.Manga, error) {
	rows, err := r.DB.Query(`
		SELECT id, title, author, artist, genres, chapter_count,
		       published_year, status, cover_url, description
		FROM mangas
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mangas []models.Manga
	for rows.Next() {
		var m models.Manga
		var genresJSON string

		if err := rows.Scan(
			&m.ID, &m.Title, &m.Author, &m.Artist, &genresJSON,
			&m.ChapterCount, &m.PublishedYear, &m.Status,
			&m.CoverURL, &m.Description,
		); err != nil {
			return nil, err
		}

		if genresJSON != "" {
			_ = json.Unmarshal([]byte(genresJSON), &m.Genres)
		}
		mangas = append(mangas, m)
	}
	return mangas, nil
}
