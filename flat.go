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
	htmlBytes, err := staticFS.ReadFile(fmt.Sprintf(htmlEmail, dt.Name()))

	if err != nil {
		return ""
	}

	return string(htmlBytes)
}

// PlainTextTemplate returns a Golang template that will generate an plain text email.
func (dt Flat) PlainTextTemplate() string {
	plainTextBytes, err := staticFS.ReadFile(fmt.Sprintf(plainTextEmail, dt.Name()))

	if err != nil {
		return ""
	}

	return string(plainTextBytes)
}
