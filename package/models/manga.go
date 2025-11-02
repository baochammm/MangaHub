package models

type Manga struct {
	ID            string   `json:"id" db:"id"`
	Title         string   `json:"title" db:"title"`
	Author        string   `json:"author" db:"author"`
	Artist        string   `json:"artist" db:"artist"`
	Genres        []string `json:"genres" db:"genres"`
	ChapterCount  int      `json:"chapter_count" db:"chapter_count"`
	PublishedYear int      `json:"published_year" db:"published_year"`
	Status        string   `json:"status" db:"status"`
	CoverURL      string   `json:"cover_url" db:"cover_url"`
	Description   string   `json:"description" db:"description"`
}
