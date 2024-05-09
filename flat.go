package hermes

import "fmt"

// Flat is a theme
type Flat struct{}

// Name returns the name of the flat theme
func (dt Flat) Name() string {
	return "flat"
}

// HTMLTemplate returns a Golang template that will generate an HTML email.
func (dt Flat) HTMLTemplate() string {
	return getTemplate(fmt.Sprintf(htmlEmail, dt.Name()))
}

// PlainTextTemplate returns a Golang template that will generate an plain text email.
func (dt Flat) PlainTextTemplate() string {
	return getTemplate(fmt.Sprintf(plainTextEmail, dt.Name()))
}
