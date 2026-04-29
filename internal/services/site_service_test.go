package services

import (
	"ai-later-nav/internal/models"
	"testing"
)

func TestBuildDisplayTags_MapsTrimmedNameToStableClass(t *testing.T) {
	tags := []string{"AI对话", "  AI对话  ", "搜索"}

	displayTags := buildDisplayTags(tags)

	if len(displayTags) != 3 {
		t.Fatalf("expected 3 display tags, got %d", len(displayTags))
	}
	if displayTags[0].Name != "AI对话" {
		t.Fatalf("expected first tag name to be trimmed, got %q", displayTags[0].Name)
	}
	if displayTags[1].Name != "AI对话" {
		t.Fatalf("expected second tag name to be trimmed, got %q", displayTags[1].Name)
	}
	if displayTags[0].Class == "" || displayTags[2].Class == "" {
		t.Fatal("expected display tags to have classes")
	}
	if displayTags[0].Class != displayTags[1].Class {
		t.Fatalf("expected trimmed-equivalent tags to share class: %q vs %q", displayTags[0].Class, displayTags[1].Class)
	}
}

func TestBuildSiteDisplay_IncludesDisplayTags(t *testing.T) {
	site := models.SiteWithTags{
		Site: models.Site{
			ID:   1,
			Name: "Test Site",
		},
		Tags: []string{"AI对话", "搜索"},
	}

	display := buildSiteDisplay(site, 12)

	if len(display.DisplayTags) != 2 {
		t.Fatalf("expected 2 display tags, got %d", len(display.DisplayTags))
	}
	if display.DisplayTags[0].Name != "AI对话" {
		t.Fatalf("expected first display tag name, got %q", display.DisplayTags[0].Name)
	}
	if display.TodayUV != 12 {
		t.Fatalf("expected TodayUV 12, got %d", display.TodayUV)
	}
}
