package manga

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/baochammm/mangahub/package/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) GetAll(page int, pageSize int) (*models.PaginatedMangas, error) {
	var totalItems int
	err := r.DB.QueryRow(`SELECT COUNT(*) FROM mangas`).Scan(&totalItems)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	// 2. Fetch paginated rows
	rows, err := r.DB.Query(`
		SELECT id, title, author, artist, genres, chapter_count,
		       published_year, status, cover_url, description
		FROM mangas
		LIMIT ? OFFSET ?
	`, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mangas []models.Manga

	for rows.Next() {
		var m models.Manga
		var genresJSON sql.NullString

		if err := rows.Scan(
			&m.ID,
			&m.Title,
			&m.Author,
			&m.Artist,
			&genresJSON,
			&m.ChapterCount,
			&m.PublishedYear,
			&m.Status,
			&m.CoverURL,
			&m.Description,
		); err != nil {
			return nil, err
		}

		if genresJSON.Valid {
			_ = json.Unmarshal([]byte(genresJSON.String), &m.Genres)
		}

		mangas = append(mangas, m)
	}

	return &models.PaginatedMangas{
		Items:      mangas,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}, nil
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
func (r *Repository) UpdateManga(m models.Manga) (bool, error) {
	var genresJSON string
	if len(m.Genres) > 0 {
		data, err := json.Marshal(m.Genres)
		if err != nil {
			return false, err
		}
		genresJSON = string(data)
	}
	_, err := r.DB.Exec(`
		UPDATE mangas
		SET title = ?, author = ?, artist = ?, genres = ?, chapter_count = ?,
		published_year = ?, status = ?, cover_url = ?, description = ?
		WHERE id = ?
	`, m.Title, m.Author, m.Artist, genresJSON, m.ChapterCount,
		m.PublishedYear, m.Status, m.CoverURL, m.Description, m.ID,
	)
	if err != nil {
		return false, err
	}
	return true, nil

}

func (r *Repository) ExistsByTitle(title string) (bool, error) {
	row := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM mangas
		WHERE id = ?
	`, title)
	var count int
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}
