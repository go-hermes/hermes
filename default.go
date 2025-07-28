package hermes

import (
	"fmt"
	"html/template"
)

var (
	parsedDefaultHTML      = template.Must(TemplateBase().Parse(Default{}.HTMLTemplate()))
	parsedDefaultPlainText = template.Must(TemplateBase().Parse(Default{}.PlainTextTemplate()))
)

// Default is the theme by default
type Default struct{}

// Name returns the name of the default theme
func (dt Default) Name() string {
	return "default"
}

// HTMLTemplate returns a Golang template that will generate an HTML email.
func (dt Default) HTMLTemplate() string {
	return getTemplate(fmt.Sprintf(htmlEmail, dt.Name()))
}

// PlainTextTemplate returns a Golang template that will generate an plain text email.
func (dt Default) PlainTextTemplate() string {
	return getTemplate(fmt.Sprintf(plainTextEmail, dt.Name()))
}

func (dt Default) ParsedHTMLTemplate() (*template.Template, error) {
	return parsedDefaultHTML, nil
}

func (dt Default) ParsedPlainTextTemplate() (*template.Template, error) {
	return parsedDefaultPlainText, nil
}
