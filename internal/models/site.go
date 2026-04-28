package models

import "time"

type Site struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	URL         string    `json:"url" db:"url"`
	Description string    `json:"description" db:"description"`
	Logo        string    `json:"logo" db:"logo"`
	Category    string    `json:"category" db:"category"`
	Rating      float64   `json:"rating" db:"rating"`
	Visits      int64     `json:"visits" db:"visits"`
	Featured    bool      `json:"featured" db:"featured"`
	Deleted     bool      `json:"deleted" db:"deleted"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type SiteWithTags struct {
	Site
	Tags      []string `json:"tags"`
	Color     string   `json:"color"`
	Initials  string   `json:"initials"`
	IsFav     bool     `json:"is_fav"`
}

type SiteDisplay struct {
	Site
	Tags     []string `json:"tags"`
	Color    string   `json:"color"`
	Initials string   `json:"initials"`
	IsFav    bool     `json:"is_fav"`
}
