package utils

import (
	"strings"
	"testing"
)

func TestGenerateColorFromName(t *testing.T) {
	color := GenerateColorFromName("ChatGPT")
	if !strings.HasPrefix(color, "#") || len(color) != 7 {
		t.Errorf("expected hex color like #xxxxxx, got %s", color)
	}

	// 相同名称应产生相同颜色
	color2 := GenerateColorFromName("ChatGPT")
	if color != color2 {
		t.Errorf("same name should produce same color: %s vs %s", color, color2)
	}

	// 不同名称应产生不同颜色
	color3 := GenerateColorFromName("Claude")
	if color == color3 {
		t.Errorf("different names should produce different colors: %s vs %s", color, color3)
	}
}

func TestGenerateColorFromName_Empty(t *testing.T) {
	color := GenerateColorFromName("")
	if !strings.HasPrefix(color, "#") || len(color) != 7 {
		t.Errorf("empty name should still produce valid color, got %s", color)
	}
}

func TestGetInitialsFromName(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", "?"},
		{"A", "A"},
		{"AI", "A"},
		{"ChatGPT", "Ch"},
		{"Open AI", "OA"},
		{"My Tool", "MT"},
	}

	for _, tt := range tests {
		got := GetInitialsFromName(tt.name)
		if got != tt.expected {
			t.Errorf("GetInitialsFromName(%q) = %q, want %q", tt.name, got, tt.expected)
		}
	}
}

func TestGetInitialsFromName_Chinese(t *testing.T) {
	got := GetInitialsFromName("百度")
	if got != "百" {
		t.Errorf("GetInitialsFromName(百度) = %q, want 百", got)
	}

	got = GetInitialsFromName("谷")
	if got != "谷" {
		t.Errorf("GetInitialsFromName(谷) = %q, want 谷", got)
	}

	got = GetInitialsFromName("百度搜索引擎")
	if got != "百度" {
		t.Errorf("GetInitialsFromName(百度搜索引擎) = %q, want 百度", got)
	}
}
