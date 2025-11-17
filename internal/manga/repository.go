package manga

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

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
func (r *Repository) GetByID(id string) (*models.Manga, error) {
	row := r.DB.QueryRow(`
		SELECT id, title, author, artist, genres, chapter_count, published_year, status, cover_url, description
		FROM mangas WHERE id = ?
	`, id)

	var m models.Manga
	var genresJSON string
	if err := row.Scan(
		&m.ID, &m.Title, &m.Author, &m.Artist, &genresJSON,
		&m.ChapterCount, &m.PublishedYear, &m.Status,
		&m.CoverURL, &m.Description,
	); err != nil {
		return nil, err
	}

	if genresJSON != "" {
		_ = json.Unmarshal([]byte(genresJSON), &m.Genres)
	}
	return &m, nil
}

func (r *Repository) SearchByTitle(query string) ([]models.Manga, error) {
	searchTerm := strings.ToLower(query)
	rows, err := r.DB.Query(`
		SELECT id, title, author, artist, genres, chapter_count,
		       published_year, status, cover_url, description
		FROM mangas
		WHERE ' ' || LOWER(title) || ' ' LIKE '% ' || ? || ' %'
	`, searchTerm)
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

func (r *Repository) FilterByGenre(genres []string) ([]models.Manga, error) {
	if len(genres) == 0 {
		return nil, nil
	}

	placeholders := strings.Repeat("?,", len(genres))
	placeholders = strings.TrimRight(placeholders, ",")

	args := make([]interface{}, len(genres))
	for i, g := range genres {
		args[i] = strings.ToLower(g)
	}

	query := fmt.Sprintf(`
		SELECT id, title, author, artist, genres, chapter_count,
		       published_year, status, cover_url, description
		FROM mangas
		WHERE (
			SELECT COUNT(DISTINCT json_each.value)
			FROM json_each(genres)
			WHERE LOWER(json_each.value) IN (%s)
		) = ?
	`, placeholders)

	rows, err := r.DB.Query(query, append(args, len(genres))...)
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
