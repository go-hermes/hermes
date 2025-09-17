package hermes

import (
	"embed"

	"github.com/sirupsen/logrus"
)

var (
	//go:embed templates
	staticFS embed.FS
)

const (
	htmlEmail      = "templates/%s.tpl.html"
	plainTextEmail = "templates/%s.tpl.txt"
)

func getTemplate(name string) string {
	htmlBytes, err := staticFS.ReadFile(name)

	if err != nil {
		logrus.Fatal(err)
	}

	return string(htmlBytes)
}

func GetDefaultStyles() StylesDefinition {
	cssBytes, err := staticFS.ReadFile("templates/default.css")
	if err != nil {
		logrus.Fatal(err)
	}
	return ParseStylesDefinition(string(cssBytes))
}
