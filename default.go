package hermes

import (
	"fmt"
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
