package hermes

import (
	"strings"
	"testing"
)

func TestTableTitleUnsafeOverrides(t *testing.T) {
	h := Hermes{}
	row := []Entry{{Key: "Col1", Value: "Val1"}}
	email := Email{Body: Body{Tables: []Table{}}}
	email.Body.Tables = append(email.Body.Tables, Table{
		Title:       "Safe Title",
		TitleUnsafe: "<em>Unsafe Title</em>",
		Data:        [][]Entry{row},
	})
	html, err := h.GenerateHTML(email)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}
	if !strings.Contains(html, "<em>Unsafe Title</em>") {
		t.Fatalf("expected unsafe title HTML to be rendered")
	}
	if strings.Contains(html, "Safe Title") {
		t.Fatalf("safe title should not appear when unsafe provided")
	}
}

func TestTableTitleSafeFallback(t *testing.T) {
	h := Hermes{}
	row := []Entry{{Key: "Col1", Value: "Val1"}}
	email := Email{Body: Body{Tables: []Table{}}}
	email.Body.Tables = append(email.Body.Tables, Table{
		Title: "Only Safe Title",
		Data:  [][]Entry{row},
	})
	html, err := h.GenerateHTML(email)
	if err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}
	if !strings.Contains(html, "Only Safe Title") {
		t.Fatalf("expected safe title to be rendered")
	}
}
