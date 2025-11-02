package models

import "time"

type User struct {
	UserID       int           `json:"user_id" db:"id"`
	Username     string        `json:"username" db:"username"`
	PasswordHash string        `json:"password_hash" db:"password_hash"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
	ReadingLists *ReadingLists `json:"reading_lists,omitempty"`
}
type ReadingEntry struct {
	MangaID        string    `json:"manga_id" db:"manga_id"`
	CurrentChapter int       `json:"current_chapter" db:"current_chapter"`
	Status         string    `json:"status" db:"status"`
	LastUpdated    time.Time `json:"last_updated" db:"last_updated"`
}

type ReadingLists struct {
	Reading    []ReadingEntry `json:"reading"`
	Completed  []ReadingEntry `json:"completed"`
	PlanToRead []ReadingEntry `json:"plan_to_read"`
}
