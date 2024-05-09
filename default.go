package hermes

import (
	"embed"
	"fmt"
)

var (
	//go:embed templates
	staticFS embed.FS
)

const (
	htmlEmail      = "templates/%s.tpl.html"
	plainTextEmail = "templates/%s.tpl.txt"
)

// Default is the theme by default
type Default struct{}

// Name returns the name of the default theme
func (dt Default) Name() string {
	return "default"
}

// HTMLTemplate returns a Golang template that will generate an HTML email.
func (dt Default) HTMLTemplate() string {
	htmlBytes, err := staticFS.ReadFile(fmt.Sprintf(htmlEmail, dt.Name()))

	if err != nil {
		return ""
	}

	return string(htmlBytes)
}

// PlainTextTemplate returns a Golang template that will generate an plain text email.
func (dt Default) PlainTextTemplate() string {
	plainTextBytes, err := staticFS.ReadFile(fmt.Sprintf(plainTextEmail, dt.Name()))

	if err != nil {
		return ""
	}

	return string(plainTextBytes)
}
