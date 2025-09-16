package hermes

import (
	"fmt"
	"html/template"
)

var (
	parsedFlatHTML      = template.Must(TemplateBase().Parse(Flat{}.HTMLTemplate()))
	parsedFlatPlainText = template.Must(TemplateBase().Parse(Flat{}.PlainTextTemplate()))
)

// Flat is another built-in theme
type Flat struct{}

// Name returns the name of the flat theme
func (dt Flat) Name() string {
	return "flat"
}

func (dt Flat) Styles() StylesDefinition {
	base := GetDefaultStyles()
	styles := StylesDefinition{}
	for sel, props := range base {
		cloned := map[string]any{}
		for k, v := range props {
			cloned[k] = v
		}
		styles[sel] = cloned
	}

	// Ensure keys exist before mutation (defensive for parser changes)
	ensure := func(sel string) map[string]any {
		if m, ok := styles[sel]; ok && m != nil {
			return m
		}
		m := map[string]any{}
		styles[sel] = m
		return m
	}

	ensure("body")["background-color"] = "#2c3e50"
	ensure(".email-wrapper")["background-color"] = "#2c3e50"
	ensure(".email-footer p")["color"] = "#eaeaea"
	ensure(".button")["background-color"] = "#00948d"
	ensure(".button")["border-radius"] = "0"

	return styles
}

// HTMLTemplate returns a Golang template that will generate an HTML email.
func (dt Flat) HTMLTemplate() string {
	// Reuse the default HTML template; styling differences handled via CSS & template vars
	return getTemplate(fmt.Sprintf(htmlEmail, "default"))
}

// PlainTextTemplate returns a Golang template that will generate an plain text email.
func (dt Flat) PlainTextTemplate() string {
	return getTemplate(fmt.Sprintf(plainTextEmail, "default"))
}

func (dt Flat) ParsedHTMLTemplate() (*template.Template, error) {
	return parsedFlatHTML, nil
}

func (dt Flat) ParsedPlainTextTemplate() (*template.Template, error) {
	return parsedFlatPlainText, nil
}
