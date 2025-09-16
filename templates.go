package hermes

import (
	"embed"
	"encoding/json"

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
	stylesBytes, err := staticFS.ReadFile("templates/default.css.json")
	if err != nil {
		logrus.Fatal(err)
	}

	var styles StylesDefinition
	err = json.Unmarshal(stylesBytes, &styles)
	if err != nil {
		logrus.Fatal(err)
	}

	return styles
}
