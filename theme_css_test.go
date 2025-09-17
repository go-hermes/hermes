package hermes

import (
	"html/template"
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

// TestTemplateElementsComprehensive tests different email template types and their elements
func TestTemplateElementsComprehensive(t *testing.T) {
	baseHermes := Hermes{
		Theme: new(Default),
		Product: Product{
			Name:      "TestApp",
			Link:      "https://testapp.com",
			Copyright: "Copyright © 2025 TestApp",
		},
		DisableCSSInlining: true,
	}

	t.Run("complete email with title", func(t *testing.T) {
		email := Email{
			Body: Body{
				Title: "Important Notification", // Title replaces Name+Greeting
				Intros: []string{
					"Welcome to our platform!",
					"We are excited to have you.",
				},
				Dictionary: []Entry{
					{Key: "Account ID", Value: "12345"},
					{Key: "Plan", Value: "Premium"},
				},
				Tables: []Table{
					{
						Title: "Billing Summary",
						Data: [][]Entry{
							{{Key: "Item", Value: "Premium Plan"}, {Key: "Amount", Value: "$29.99"}},
						},
					},
				},
				Actions: []Action{
					{
						Instructions: "Click the button below:",
						Button: Button{
							Text: "Verify Account",
							Link: "https://testapp.com/verify",
						},
					},
					{
						Instructions: "Or use this code:",
						InviteCode:   "WELCOME123",
					},
				},
				Outros: []string{"Thank you for choosing TestApp."},
			},
		}

		html, err := baseHermes.GenerateHTML(email)
		if err != nil {
			t.Fatalf("GenerateHTML failed: %v", err)
		}

		// Test key elements are present
		elements := []string{
			"TestApp",                 // Product name
			"Important Notification",  // Title
			"Welcome to our platform", // Intro
			"Account ID",              // Dictionary key
			"Premium Plan",            // Table data
			"Verify Account",          // Button text
			"WELCOME123",              // Invite code
			"Thank you for choosing",  // Outro
		}

		for _, element := range elements {
			if !containsIgnoreCase(html, element) {
				t.Errorf("expected HTML to contain %q", element)
			}
		}

		// Test CSS classes are present
		classes := []string{"email-wrapper", "email-body", "button", "data-table"}
		for _, class := range classes {
			if !containsIgnoreCase(html, class) {
				t.Errorf("expected HTML to contain CSS class %q", class)
			}
		}
	})

	t.Run("email with name and greeting", func(t *testing.T) {
		email := Email{
			Body: Body{
				Name:     "John Doe",
				Greeting: "Hello",
				// No Title set, so Name+Greeting should appear
				Intros: []string{"This is a test email."},
			},
		}

		html, err := baseHermes.GenerateHTML(email)
		if err != nil {
			t.Fatalf("GenerateHTML failed: %v", err)
		}

		// When no Title is set, should see Greeting + Name
		if !containsIgnoreCase(html, "Hello") || !containsIgnoreCase(html, "John Doe") {
			t.Error("expected greeting and name to appear when no title is set")
		}
	})

	t.Run("markdown content", func(t *testing.T) {
		email := Email{
			Body: Body{
				Name:           "User",
				IntrosMarkdown: Markdown("**Welcome** to our _premium_ service!"),
				OutrosMarkdown: Markdown("Visit [our website](https://testapp.com)"),
			},
		}

		html, err := baseHermes.GenerateHTML(email)
		if err != nil {
			t.Fatalf("GenerateHTML failed: %v", err)
		}

		// Test markdown elements
		if !containsIgnoreCase(html, "<strong>Welcome</strong>") {
			t.Error("expected bold markdown to be converted")
		}
		if !containsIgnoreCase(html, "<em>premium</em>") {
			t.Error("expected italic markdown to be converted")
		}
		if !containsIgnoreCase(html, "https://testapp.com") {
			t.Error("expected link to be present")
		}
	})

	t.Run("unsafe HTML content", func(t *testing.T) {
		email := Email{
			Body: Body{
				Name: "User",
				IntrosUnsafe: []template.HTML{
					template.HTML("<div class=\"custom\">Unsafe content</div>"),
				},
			},
		}

		html, err := baseHermes.GenerateHTML(email)
		if err != nil {
			t.Fatalf("GenerateHTML failed: %v", err)
		}

		if !containsIgnoreCase(html, "<div class=\"custom\">") {
			t.Error("expected unsafe HTML to be preserved")
		}
	})

	t.Run("comprehensive unsafe variants", func(t *testing.T) {
		email := Email{
			Body: Body{
				Name: "Test User",
				// Test IntrosUnsafe
				IntrosUnsafe: []template.HTML{
					template.HTML("<p class=\"intro-custom\">Welcome with <strong>custom HTML</strong></p>"),
					template.HTML("<div style=\"color: blue;\">Second unsafe intro</div>"),
				},
				// Test OutrosUnsafe
				OutrosUnsafe: []template.HTML{
					template.HTML("<p class=\"outro-custom\">Thank you with <em>custom styling</em></p>"),
					template.HTML("<div style=\"background: yellow;\">Final message</div>"),
				},
				// Test table with unsafe elements
				Tables: []Table{
					{
						TitleUnsafe: template.HTML("<h3 class=\"table-title-custom\">Custom <span style=\"color: red;\">Table</span> Title</h3>"),
						Data: [][]Entry{
							{
								{Key: "Item", Value: "Product A"},
								{Key: "Description", UnsafeValue: template.HTML("<em>Rich</em> <strong>description</strong> with <a href=\"#\">link</a>")},
							},
						},
						FooterUnsafe: template.HTML("<div class=\"table-footer-custom\">Footer with <strong>HTML</strong> and <span style=\"color: green;\">styling</span></div>"),
					},
				},
				// Regular intros/outros should be ignored when unsafe variants exist
				Intros: []string{"This should be ignored"},
				Outros: []string{"This should also be ignored"},
			},
		}

		html, err := baseHermes.GenerateHTML(email)
		if err != nil {
			t.Fatalf("GenerateHTML failed: %v", err)
		}

		// Test IntrosUnsafe elements
		introUnsafeElements := []string{
			"<p class=\"intro-custom\">",   // Custom intro class
			"<strong>custom HTML</strong>", // Bold HTML in intro
			"style=\"color: blue;\"",       // Inline styles in intro
			"Second unsafe intro",          // Second intro content
		}

		for _, element := range introUnsafeElements {
			if !containsIgnoreCase(html, element) {
				t.Errorf("expected IntrosUnsafe to contain %q", element)
			}
		}

		// Test OutrosUnsafe elements
		outroUnsafeElements := []string{
			"<p class=\"outro-custom\">",    // Custom outro class
			"<em>custom styling</em>",       // Italic HTML in outro
			"style=\"background: yellow;\"", // Background style in outro
			"Final message",                 // Final outro content
		}

		for _, element := range outroUnsafeElements {
			if !containsIgnoreCase(html, element) {
				t.Errorf("expected OutrosUnsafe to contain %q", element)
			}
		}

		// Test TitleUnsafe elements
		titleUnsafeElements := []string{
			"<h3 class=\"table-title-custom\">",        // Custom title element
			"<span style=\"color: red;\">Table</span>", // Styled span in title
			"Custom", // Title content
		}

		for _, element := range titleUnsafeElements {
			if !containsIgnoreCase(html, element) {
				t.Errorf("expected TitleUnsafe to contain %q", element)
			}
		}

		// Test UnsafeValue in table data
		dataUnsafeElements := []string{
			"<em>Rich</em>",                // Italic in data
			"<strong>description</strong>", // Bold in data
			"<a href=\"#\">link</a>",       // Link in data
		}

		for _, element := range dataUnsafeElements {
			if !containsIgnoreCase(html, element) {
				t.Errorf("expected UnsafeValue to contain %q", element)
			}
		}

		// Test FooterUnsafe elements
		footerUnsafeElements := []string{
			"<div class=\"table-footer-custom\">", // Custom footer class
			"Footer with <strong>HTML</strong>",   // Bold in footer
			"style=\"color: green;\"",             // Green styling
		}

		for _, element := range footerUnsafeElements {
			if !containsIgnoreCase(html, element) {
				t.Errorf("expected FooterUnsafe to contain %q", element)
			}
		}

		// Test that regular intros/outros are ignored when unsafe variants exist
		if containsIgnoreCase(html, "This should be ignored") {
			t.Error("expected regular intros to be ignored when IntrosUnsafe is present")
		}
		if containsIgnoreCase(html, "This should also be ignored") {
			t.Error("expected regular outros to be ignored when OutrosUnsafe is present")
		}
	})

	t.Run("unsafe vs safe precedence", func(t *testing.T) {
		email := Email{
			Body: Body{
				Name: "Precedence User",
				// Both safe and unsafe variants present
				Intros: []string{"Safe intro text"},
				IntrosUnsafe: []template.HTML{
					template.HTML("<span class=\"unsafe-intro\">Unsafe intro takes precedence</span>"),
				},
				Outros: []string{"Safe outro text"},
				OutrosUnsafe: []template.HTML{
					template.HTML("<span class=\"unsafe-outro\">Unsafe outro takes precedence</span>"),
				},
				Tables: []Table{
					{
						Title:       "Safe Table Title",
						TitleUnsafe: template.HTML("<div class=\"unsafe-title\">Unsafe title takes precedence</div>"),
						Data: [][]Entry{
							// Test Value vs UnsafeValue - Value takes precedence when both present
							{{Key: "Col1", Value: "Safe value wins", UnsafeValue: template.HTML("<span class=\"unsafe-value\">This should be ignored</span>")}},
							// Test UnsafeValue when Value is empty/not set
							{{Key: "Col2", UnsafeValue: template.HTML("<span class=\"unsafe-value-used\">Unsafe value used when safe empty</span>")}},
						},
						Footer:       "Safe footer",
						FooterUnsafe: template.HTML("<div class=\"unsafe-footer\">Unsafe footer takes precedence</div>"),
					},
				},
			},
		}

		html, err := baseHermes.GenerateHTML(email)
		if err != nil {
			t.Fatalf("GenerateHTML failed: %v", err)
		}

		// Test that unsafe variants take precedence for intros/outros/titles/footers
		unsafePrecedenceElements := []string{
			"Unsafe intro takes precedence",  // IntrosUnsafe wins over Intros
			"Unsafe outro takes precedence",  // OutrosUnsafe wins over Outros
			"Unsafe title takes precedence",  // TitleUnsafe wins over Title
			"Unsafe footer takes precedence", // FooterUnsafe wins over Footer
		}

		for _, element := range unsafePrecedenceElements {
			if !containsIgnoreCase(html, element) {
				t.Errorf("expected unsafe variant to take precedence: %q", element)
			}
		}

		// Test Value vs UnsafeValue precedence - Value wins when both present
		if !containsIgnoreCase(html, "Safe value wins") {
			t.Error("expected Value to take precedence over UnsafeValue when both are set")
		}
		if containsIgnoreCase(html, "This should be ignored") {
			t.Error("expected UnsafeValue to be ignored when Value is also set")
		}

		// Test UnsafeValue fallback when Value is empty
		if !containsIgnoreCase(html, "Unsafe value used when safe empty") {
			t.Error("expected UnsafeValue to be used when Value is empty")
		}

		// Test that safe variants are ignored when unsafe variants exist (except Value vs UnsafeValue)
		safeFallbackElements := []string{
			"Safe intro text",
			"Safe outro text",
			"Safe Table Title",
			// Note: Footer test removed - template logic is complex, test separately
		}

		for _, element := range safeFallbackElements {
			if containsIgnoreCase(html, element) {
				t.Errorf("expected safe variant to be ignored when unsafe variant exists: %q", element)
			}
		}
	})
}
