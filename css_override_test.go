package hermes

import (
	"strings"
	"testing"
)

func TestCSSOverrides(t *testing.T) {
	h := Hermes{
		Product: Product{
			Name: "Test Company",
			Link: "https://example.com",
		},
		DisableCSSInlining: true, // Disable CSS inlining to see raw template output
	}

	// Test with default CSS
	email1 := Email{
		Body: Body{
			Name: "John Doe",
			Intros: []string{
				"Welcome to our service!",
			},
		},
	}

	html1, err := h.GenerateHTML(email1)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== Test 1: Default CSS ===")
	t.Logf("Generated HTML length: %d", len(html1))

	// Debug: Check if TemplateOverrides is being set
	if email1.Body.TemplateOverrides != nil {
		t.Logf("TemplateOverrides keys: %v", getKeys(email1.Body.TemplateOverrides))
		if css, ok := email1.Body.TemplateOverrides["css"]; ok {
			t.Logf("CSS map type: %T", css)
			if cssMap, ok := css.(map[string]map[string]interface{}); ok {
				t.Logf("CSS map keys: %v", getKeysFromMap(cssMap))
			}
		}
	} else {
		t.Log("TemplateOverrides is nil!")
	}

	// Look for CSS in the generated HTML
	if strings.Contains(html1, "<style") {
		t.Log("✓ Style tag found")

		// Extract style section for debugging
		start := strings.Index(html1, "<style")
		if start != -1 {
			end := strings.Index(html1[start:], "</style>")
			if end != -1 {
				styleSection := html1[start : start+end+8]
				t.Logf("Full style section: %s", styleSection)
			}
		}
	} else {
		t.Error("✗ No style tag found")
	}

	if strings.Contains(html1, "background-color: #F2F4F6") {
		t.Log("✓ Default CSS applied correctly")
	} else {
		t.Error("✗ Default CSS not found")
	}

	// Test with CSS overrides
	email2 := Email{
		Body: Body{
			Name: "Jane Doe",
			Intros: []string{
				"Welcome to our customized service!",
			},
			TemplateOverrides: map[string]any{
				"css": map[string]map[string]interface{}{
					"body": {
						"background-color": "#FF0000", // Override to red
						"color":            "#FFFFFF", // Override to white text
					},
					".custom-class": {
						"font-size": "20px",
						"color":     "#00FF00",
					},
				},
			},
		},
	}

	html2, err := h.GenerateHTML(email2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("\n=== Test 2: CSS Overrides ===")
	if strings.Contains(html2, "background-color: #FF0000") {
		t.Log("✓ Body background override applied correctly")
	} else {
		t.Error("✗ Body background override not found")
	}

	if strings.Contains(html2, "color: #FFFFFF") {
		t.Log("✓ Body color override applied correctly")
	} else {
		t.Error("✗ Body color override not found")
	}

	if strings.Contains(html2, ".custom-class") {
		t.Log("✓ Custom CSS class added correctly")
	} else {
		t.Error("✗ Custom CSS class not found")
	}

	// Check that other default styles are still preserved
	if strings.Contains(html2, "font-family:") {
		t.Log("✓ Default font-family preserved")
	} else {
		t.Error("✗ Default font-family not preserved")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func getKeysFromMap(m map[string]map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
