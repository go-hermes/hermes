package hermes

import (
	"strings"
	"testing"
)

// normalizeColorCase converts hex colors to lowercase for case-insensitive comparisons
func normalizeColorCase(html string) string {
	// Simple approach: just lowercase the entire string for testing
	return strings.ToLower(html)
}

// containsIgnoreCase checks if html contains target (case-insensitive)
func containsIgnoreCase(html, target string) bool {
	return strings.Contains(normalizeColorCase(html), normalizeColorCase(target))
}

// Consolidated CSS / theme tests replacing css_override_test.go, flat_theme_test.go and body_width_test.go.
// Scenarios exercise baseline and override behaviors across themes plus body width & additional styles.
func TestThemeAndCSSScenarios(t *testing.T) {
	type assertion struct {
		contains    []string
		notContains []string
		name        string
	}

	scenarios := []struct {
		name       string
		hermes     Hermes
		email      Email
		assertions []assertion
	}{
		{
			name: "default theme baseline",
			hermes: Hermes{ // default theme implied
				Product:            Product{Name: "Test", Link: "https://example.com"},
				DisableCSSInlining: true,
			},
			email: Email{Body: Body{Intros: []string{"Hi"}}},
			assertions: []assertion{{
				name:        "default styles present and no flat colors",
				contains:    []string{"background-color: #F2F4F6", "background-color: #3869D4", "font-family:"},
				notContains: []string{"background-color: #2c3e50", "background-color: #00948d"},
			}},
		},
		{
			name:   "default theme with css overrides",
			hermes: Hermes{Product: Product{Name: "Test", Link: "https://example.com"}, DisableCSSInlining: true},
			email: Email{Body: Body{Intros: []string{"Hi"}, TemplateOverrides: map[string]any{
				"css": map[string]map[string]any{
					"body":          {"background-color": "#FF0000", "color": "#FFFFFF"},
					".custom-class": {"font-size": "20px", "color": "#00FF00"},
				},
			}}},
			assertions: []assertion{{
				name:        "overrides applied and defaults still merged",
				contains:    []string{"background-color: #FF0000", "color: #FFFFFF", ".custom-class", "font-family:"},
				notContains: []string{"background-color: #2c3e50"},
			}},
		},
		{
			name:   "flat theme baseline",
			hermes: Hermes{Theme: new(Flat), Product: Product{Name: "Test", Link: "https://example.com"}, DisableCSSInlining: true},
			email:  Email{Body: Body{Intros: []string{"Hi"}}},
			assertions: []assertion{{
				name:        "flat styles present and default light bg absent",
				contains:    []string{"background-color: #2c3e50", "background-color: #00948d"},
				notContains: []string{"background-color: #F2F4F6"},
			}},
		},
		{
			name:   "flat theme with overrides and body width",
			hermes: Hermes{Theme: new(Flat), Product: Product{Name: "Test", Link: "https://example.com"}, DisableCSSInlining: true},
			email: Email{Body: Body{Intros: []string{"Hi"}, TemplateOverrides: map[string]any{
				"body_width": "800px",
				"css": map[string]map[string]any{
					"body":         {"background-color": "#FF0000"},
					".custom-flat": {"color": "#FFFF00"},
				},
			}}},
			assertions: []assertion{{
				name:        "flat overrides and width applied",
				contains:    []string{"background-color: #FF0000", ".custom-flat", "width: 800px", "max-width: 800px"},
				notContains: []string{"background-color: #F2F4F6"},
			}},
		},
		{
			name:   "body width and additional styles combo",
			hermes: Hermes{Product: Product{Name: "Test", Link: "https://example.com"}, DisableCSSInlining: true},
			email: Email{Body: Body{Intros: []string{"Hi"}, TemplateOverrides: map[string]any{
				"body_width":        "700px",
				"additional_styles": ".custom { color: red; }",
			}}},
			assertions: []assertion{{
				name:     "width & additional styles present",
				contains: []string{"width: 700px", "max-width: 700px", "color: red"},
			}},
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			html, err := sc.hermes.GenerateHTML(sc.email)
			if err != nil {
				t.Fatalf("GenerateHTML failed: %v", err)
			}
			for _, as := range sc.assertions {
				for _, needle := range as.contains {
					if needle == "" { // skip empty
						continue
					}
					if !containsIgnoreCase(html, needle) {
						t.Errorf("assertion '%s': expected to contain %q", as.name, needle)
					}
				}
				for _, needle := range as.notContains {
					if needle == "" {
						continue
					}
					if containsIgnoreCase(html, needle) {
						t.Errorf("assertion '%s': expected NOT to contain %q", as.name, needle)
					}
				}
			}
		})
	}
}

func TestTableTitleBehavior(t *testing.T) {
	h := Hermes{}
	row := []Entry{{Key: "Col1", Value: "Val1"}}

	t.Run("unsafe title overrides safe title", func(t *testing.T) {
		email := Email{Body: Body{Tables: []Table{{
			Title:       "Safe Title",
			TitleUnsafe: "<em>Unsafe Title</em>",
			Data:        [][]Entry{row},
		}}}}

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
	})

	t.Run("safe title fallback when no unsafe title", func(t *testing.T) {
		email := Email{Body: Body{Tables: []Table{{
			Title: "Only Safe Title",
			Data:  [][]Entry{row},
		}}}}

		html, err := h.GenerateHTML(email)
		if err != nil {
			t.Fatalf("GenerateHTML failed: %v", err)
		}
		if !strings.Contains(html, "Only Safe Title") {
			t.Fatalf("expected safe title to be rendered")
		}
	})
}
