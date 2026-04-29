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

type DisplayTag struct {
	Name  string `json:"name"`
	Class string `json:"class"`
}

type SiteWithTags struct {
	Site
	Tags        []string     `json:"tags"`
	DisplayTags []DisplayTag `json:"display_tags"`
	Color       string       `json:"color"`
	Initials    string       `json:"initials"`
	IsFav       bool         `json:"is_fav"`
}

type SiteDisplay struct {
	Site
	Tags        []string     `json:"tags"`
	DisplayTags []DisplayTag `json:"display_tags"`
	Color       string       `json:"color"`
	Initials    string       `json:"initials"`
	IsFav       bool         `json:"is_fav"`
	TodayUV     int64        `json:"today_uv"`
}

type SiteStats struct {
	SiteID  int64 `json:"site_id" db:"site_id"`
	PV      int64 `json:"pv" db:"pv"`
	UV      int64 `json:"uv" db:"uv"`
	TodayPV int64 `json:"today_pv" db:"today_pv"`
	TodayUV int64 `json:"today_uv" db:"today_uv"`
	WeekPV  int64 `json:"week_pv" db:"week_pv"`
	WeekUV  int64 `json:"week_uv" db:"week_uv"`
}

type DashboardStats struct {
	TotalVisits   int64 `json:"total_visits"`
	TodayVisits   int64 `json:"today_visits"`
	WeekVisits    int64 `json:"week_visits"`
	FeaturedCount int64 `json:"featured_count"`
	CategoryCount int64 `json:"category_count"`
	TagCount      int64 `json:"tag_count"`
}

type CategoryStat struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}
