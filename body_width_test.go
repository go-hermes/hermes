package hermes

import (
	"strings"
	"testing"
)

func TestBodyWidthOverride(t *testing.T) {
	h := Hermes{
		Product: Product{
			Name: "Test Product",
			Link: "https://example.com",
		},
	}

	email := Email{
		Body: Body{
			Name:    "Test User",
			Intros:  []string{"Welcome to our service!"},
			Outros:  []string{"If you have any questions, please contact us."},
			Actions: []Action{},
			TemplateOverrides: map[string]any{
				"body_width": "800px",
			},
		},
	}

	// Generate HTML with custom body width
	generatedHTML, err := h.GenerateHTML(email)
	if err != nil {
		t.Fatalf("Error generating HTML: %v", err)
	}

	htmlContent := generatedHTML

	// Check that email-body_inner uses the custom width
	if !strings.Contains(htmlContent, ".email-body_inner") {
		t.Errorf("email-body_inner selector not found in generated CSS")
	}

	// Check that the custom width is applied
	if !strings.Contains(htmlContent, "width: 800px") {
		t.Errorf("Custom body width (800px) not found in generated CSS")
	}

	// Check that media query uses the custom width
	if !strings.Contains(htmlContent, "max-width: 800px") {
		t.Errorf("Custom body width not applied to media query max-width")
	}

	t.Logf("✓ Body width override test passed")
	t.Logf("✓ Custom width (800px) applied to CSS selectors")
	t.Logf("✓ Media query max-width updated to custom width")
}

func TestBodyWidthOverrideWithAdditionalStyles(t *testing.T) {
	h := Hermes{
		Product: Product{
			Name: "Test Product",
			Link: "https://example.com",
		},
		DisableCSSInlining: true, // Disable CSS inlining for testing
	}

	email := Email{
		Body: Body{
			Name:   "Test User",
			Intros: []string{"Welcome!"},
			TemplateOverrides: map[string]any{
				"body_width":        "700px",
				"additional_styles": ".custom { color: red; }",
			},
		},
	}

	// Generate HTML
	generatedHTML, err := h.GenerateHTML(email)
	if err != nil {
		t.Fatalf("Error generating HTML: %v", err)
	}

	t.Logf("Generated HTML length: %d", len(generatedHTML))

	// Find and log the style section for debugging
	styleStart := strings.Index(generatedHTML, "<style")
	styleEnd := strings.Index(generatedHTML, "</style>")
	if styleStart != -1 && styleEnd != -1 {
		styleSection := generatedHTML[styleStart : styleEnd+8]
		t.Logf("Style section: %s", styleSection)
	}

	// Check body width override
	if !strings.Contains(generatedHTML, "width: 700px") {
		t.Errorf("Custom body width (700px) not found")
	}

	// Check additional styles (might be processed by css filter)
	if !strings.Contains(generatedHTML, "color: red") {
		t.Errorf("Additional styles not found")
	}

	// Check media query
	if !strings.Contains(generatedHTML, "max-width: 700px") {
		t.Errorf("Media query not updated with custom width")
	}

	t.Logf("✓ Body width and additional styles working together")
}
