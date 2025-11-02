package user

import (
	"database/sql"
	"time"

	"github.com/baochammm/mangahub/package/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) CreateUser(user models.User) error {
	_, err := r.DB.Exec(`INSERT INTO users (username, password_hash) VALUES (?, ?)`, user.Username, user.PasswordHash)
	return err
}

func (r *Repository) AddReadingEntry(userID int, entry models.ReadingEntry) error {
	_, err := r.DB.Exec(`
		INSERT INTO reading_list (user_id, manga_id, current_chapter, status, last_updated)
		VALUES (?, ?, ?, ?, ?)
	`, userID, entry.MangaID, entry.CurrentChapter, entry.Status, entry.LastUpdated.Format(time.RFC3339))
	return err
}

func (r *Repository) GetUserWithLists(userID int) (*models.User, error) {
	user := models.User{}
	err := r.DB.QueryRow(`SELECT id, username FROM users WHERE id = ?`, userID).
		Scan(&user.UserID, &user.Username)
	if err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(`
		SELECT manga_id, current_chapter, status, last_updated
		FROM reading_list WHERE user_id = ?
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reading, completed, plan []models.ReadingEntry
	for rows.Next() {
		var entry models.ReadingEntry
		var lastUpdated string
		if err := rows.Scan(&entry.MangaID, &entry.CurrentChapter, &entry.Status, &lastUpdated); err != nil {
			return nil, err
		}
		entry.LastUpdated, _ = time.Parse(time.RFC3339, lastUpdated)

		switch entry.Status {
		case "reading":
			reading = append(reading, entry)
		case "completed":
			completed = append(completed, entry)
		case "plan_to_read":
			plan = append(plan, entry)
		}
	}

	user.ReadingLists = &models.ReadingLists{
		Reading:    reading,
		Completed:  completed,
		PlanToRead: plan,
	}

	return &user, nil
}
