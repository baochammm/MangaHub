package user

import (
	"database/sql"
	"errors"
	"time"

	"github.com/baochammm/mangahub/package/models"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

func (r *Repository) MangaExists(mangaID string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
        SELECT COUNT(*)
        FROM mangas
        WHERE id = ?
    `, mangaID).Scan(&count)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
func (r *Repository) AddReadingEntry(userID int64, entry models.ReadingEntry) error {
	_, err := r.DB.Exec(`
		INSERT INTO reading_list (user_id, manga_id, current_chapter, status, last_updated)
		VALUES (?, ?, ?, ?, ?)
	`, userID, entry.MangaID, entry.CurrentChapter, entry.Status, entry.LastUpdated.Format(time.RFC3339))
	return err
}
func (r *Repository) ReadingEntryExists(userID int64, mangaID string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
        SELECT COUNT(*) FROM reading_list
        WHERE user_id = ? AND manga_id = ?
    `, userID, mangaID).Scan(&count)

	if err != nil {
		return false, err
	}
	return count > 0, nil
}
func (r *Repository) UpdateReadingStatus(userID int64, entry models.ReadingEntry) error {
	if entry.CurrentChapter != 0 {
		_, err := r.DB.Exec(`
		UPDATE reading_list
		SET status = ?, current_chapter = ?,  last_updated = ?
		WHERE user_id = ? AND manga_id = ?
	`, entry.Status, entry.CurrentChapter, entry.LastUpdated.Format(time.RFC3339), userID, entry.MangaID)
		return err
	}
	_, err := r.DB.Exec(`
		UPDATE reading_list
		SET status = ?, last_updated = ?
		WHERE user_id = ? AND manga_id = ?
	`, entry.Status, entry.LastUpdated.Format(time.RFC3339), userID, entry.MangaID)
	return err

}
func (r *Repository) UpdateReadingProgress(userID int64, entry models.ReadingEntry) error {
	if entry.Status != "" {
		_, err := r.DB.Exec(`
		UPDATE reading_list
		SET status = ?, current_chapter = ?,  last_updated = ?
		WHERE user_id = ? AND manga_id = ?
	`, entry.Status, entry.CurrentChapter, entry.LastUpdated.Format(time.RFC3339), userID, entry.MangaID)
		return err
	}
	_, err := r.DB.Exec(`
		UPDATE reading_list
		SET current_chapter = ?, last_updated = ?
		WHERE user_id = ? AND manga_id = ?
	`, entry.CurrentChapter, entry.LastUpdated.Format(time.RFC3339), userID, entry.MangaID)
	return err

}
func (r *Repository) DeleteReadingEntry(userID int64, entry models.ReadingEntry) error {
	_, err := r.DB.Exec(`
		DELETE FROM reading_list
		WHERE user_id = ? AND manga_id = ?
	`, userID, entry.MangaID)
	return err
}
func (r *Repository) IsMangaInUserLibrary(userID int64, mangaID string) (bool, error) {
	var count int
	err := r.DB.QueryRow(`
		SELECT COUNT(*)
		FROM reading_list
		WHERE user_id = ? AND manga_id = ?
	`, userID, mangaID).Scan(&count)
	if err != nil {
		return false, errors.New("failed to check manga in user library: " + err.Error())
	}
	return count > 0, nil
}
func (r *Repository) GetUserReadingLists(userID int64) (*models.ReadingLists, error) {
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

	return user.ReadingLists, nil
}
func (r *Repository) GetUserReadingListsViaStatus(userID int64, status string) ([]models.ReadingEntry, error) {
	if status != "reading" && status != "completed" && status != "plan_to_read" {
		return nil, errors.New("invalid status value")
	}
	rows, err := r.DB.Query(`
		SELECT manga_id, current_chapter, status, last_updated
		FROM reading_list WHERE user_id = ? AND status = ?
	`, userID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.ReadingEntry
	for rows.Next() {
		var entry models.ReadingEntry
		var lastUpdated string
		if err := rows.Scan(&entry.MangaID, &entry.CurrentChapter, &entry.Status, &lastUpdated); err != nil {
			return nil, err
		}
		entry.LastUpdated, _ = time.Parse(time.RFC3339, lastUpdated)

		switch status {
		case "reading":
			entries = append(entries, entry)
		case "completed":
			entries = append(entries, entry)
		case "plan_to_read":
			entries = append(entries, entry)
		}
	}

	// readingList := &models.ReadingLists{
	// 	Reading:    reading,
	// 	Completed:  completed,
	// 	PlanToRead: plan,
	// }

	return entries, nil
}
