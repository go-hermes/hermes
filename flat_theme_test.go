package hermes

import (
	"strings"
	"testing"
)

func TestFlatThemeCSS(t *testing.T) {
	h := Hermes{
		Theme: new(Flat), // Use flat theme
		Product: Product{
			Name: "Test Company",
			Link: "https://example.com",
		},
		DisableCSSInlining: true, // Disable CSS inlining to see raw template output
	}

	email := Email{
		Body: Body{
			Name: "John Doe",
			Intros: []string{
				"Welcome to our flat-themed service!",
			},
		},
	}

	html, err := h.GenerateHTML(email)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== Flat Theme CSS Test ===")
	t.Logf("Generated HTML length: %d", len(html))

	// Check for flat theme-specific styles
	if strings.Contains(html, "background-color: #2c3e50") {
		t.Log("✓ Flat theme dark background applied correctly")
	} else {
		t.Error("✗ Flat theme dark background not found")
	}

	if strings.Contains(html, "color: #eaeaea") {
		t.Log("✓ Flat theme footer text color applied correctly")
	} else {
		t.Error("✗ Flat theme footer text color not found")
	}

	if strings.Contains(html, "background-color: #00948d") {
		t.Log("✓ Flat theme button color applied correctly")
	} else {
		t.Error("✗ Flat theme button color not found")
	}

	// Test with CSS overrides on flat theme
	email2 := Email{
		Body: Body{
			Name: "Jane Doe",
			Intros: []string{
				"Welcome with custom styles!",
			},
			TemplateOverrides: map[string]any{
				"body_width": "800px",
				"css": map[string]map[string]interface{}{
					"body": {
						"background-color": "#FF0000", // Override to red
					},
					".custom-flat": {
						"color": "#FFFF00", // Yellow text
					},
				},
			},
		},
	}

	html2, err := h.GenerateHTML(email2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("\n=== Flat Theme with Overrides ===")

	// Check that override worked
	if strings.Contains(html2, "background-color: #FF0000") {
		t.Log("✓ Body background override applied on flat theme")
	} else {
		t.Error("✗ Body background override not found on flat theme")
	}

	// Check that body width override worked
	if strings.Contains(html2, "width: 800px") {
		t.Log("✓ Body width override applied on flat theme")
	} else {
		t.Error("✗ Body width override not found on flat theme")
	}

	// Check that media query uses custom width
	if strings.Contains(html2, "max-width: 800px") {
		t.Log("✓ Media query updated with custom width on flat theme")
	} else {
		t.Error("✗ Media query not updated on flat theme")
	}

	// Check custom CSS class
	if strings.Contains(html2, ".custom-flat") {
		t.Log("✓ Custom CSS class added on flat theme")
	} else {
		t.Error("✗ Custom CSS class not found on flat theme")
	}
}

func TestDefaultThemeStillWorks(t *testing.T) {
	h := Hermes{
		Theme: new(Default), // Use default theme
		Product: Product{
			Name: "Test Company",
			Link: "https://example.com",
		},
		DisableCSSInlining: true,
	}

	email := Email{
		Body: Body{
			Name: "John Doe",
			Intros: []string{
				"Welcome to our default-themed service!",
			},
		},
	}

	html, err := h.GenerateHTML(email)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== Default Theme CSS Test ===")

	// Check for default theme styles (should not have flat customizations)
	if strings.Contains(html, "background-color: #F2F4F6") {
		t.Log("✓ Default theme light background preserved")
	} else {
		t.Error("✗ Default theme background not found")
	}

	if strings.Contains(html, "background-color: #3869D4") {
		t.Log("✓ Default theme button color preserved")
	} else {
		t.Error("✗ Default theme button color not found")
	}

	// Should NOT have flat theme colors
	if strings.Contains(html, "background-color: #2c3e50") {
		t.Error("✗ Flat theme colors incorrectly applied to default theme")
	} else {
		t.Log("✓ Flat theme colors not applied to default theme")
	}
}
