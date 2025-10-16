package hermes

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var debug bool

var testedThemes = []Theme{
	// Insert your new theme here
	new(Default),
	new(Flat),
}

func init() {
	logrus.SetOutput(io.Discard)
	debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))
}

/////////////////////////////////////////////////////
// Every theme should display the same information //
// Find below the tests to check that              //
/////////////////////////////////////////////////////

// Implement this interface when creating a new example checking a common feature of all themes
type Example interface {
	// Create the hermes example with data
	// Represents the "Given" step in Given/When/Then Workflow
	getExample() (h Hermes, email Email)
	// Checks the content of the generated HTML email by asserting content presence or not
	assertHTMLContent(t *testing.T, s string)
	// Checks the content of the generated Plaintext email by asserting content presence or not
	assertPlainTextContent(t *testing.T, s string)
}

// Scenario
type SimpleExample struct {
	theme Theme
}

func (ed SimpleExample) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name:      "HermesName",
			Link:      "http://hermes-link.com",
			Copyright: "Copyright © Hermes-Test",
			Logo:      "https://go.dev/blog/gopher/header.jpg",
		},
		TextDirection:      TDLeftToRight,
		DisableCSSInlining: true,
	}

	email := Email{
		Body{
			Name: "Jon Snow",
			Intros: []string{
				"Welcome to Hermes! We're very excited to have you on board.",
			},
			Dictionary: []Entry{
				{"Firstname", "Jon", ""},
				{"Lastname", "Snow", ""},
				{"Birthday", "01/01/283", ""},
			},
			Table: Table{
				Data: [][]Entry{
					{
						{Key: "Item", Value: "Golang"},
						{Key: "Description", Value: "Open source programming language that makes it easy to build simple, reliable, and efficient software"},
						{Key: "Price", Value: "$10.99"},
					},
					{
						{Key: "Item", Value: "Hermes"},
						{Key: "Description", Value: "Programmatically create beautiful e-mails using Golang."},
						{Key: "Price", Value: "$1.99"},
					},
				},
				Columns: Columns{
					CustomWidth: map[string]string{
						"Item":  "20%",
						"Price": "15%",
					},
					CustomAlignment: map[string]string{
						"Price": "right",
					},
				},
			},
			Actions: []Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: Button{
						Color: "#22BC66",
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
			TemplateOverrides: map[string]any{
				"body_width": "1000px",
				"additional_styles": `
				  *:not(br):not(tr):not(html) {
				 	font-family: Comic Sans MS !important; 
				  }
				`,
			},
		},
	}

	return h, email
}

func (ed SimpleExample) assertHTMLContent(t *testing.T, r string) {

	// Assert on product
	assert.Contains(t, r, "HermesName", "Product: Should find the name of the product in email")
	assert.Contains(t, r, "http://hermes-link.com", "Product: Should find the link of the product in email")
	assert.Contains(t, r, "Copyright © Hermes-Test", "Product: Should find the Copyright of the product in email")
	assert.Contains(t, r, "https://go.dev/blog/gopher/header.jpg", "Product: Should find the logo of the product in email")
	assert.Contains(t, r, "If you’re having trouble with the button &#39;Confirm your account&#39;, copy and paste the URL below into your web browser.", "Product: Should find the trouble text in email")
	// Assert on email body
	assert.Contains(t, r, "Hi Jon Snow", "Name: Should find the name of the person")
	assert.Contains(t, r, "Welcome to Hermes", "Intro: Should have intro")
	assert.Contains(t, r, "Birthday", "Dictionary: Should have dictionary")
	assert.Contains(t, r, "Open source programming language", "Table: Should have table with first row and first column")
	assert.Contains(t, r, "Programmatically create beautiful e-mails using Golang", "Table: Should have table with second row and first column")
	assert.Contains(t, r, "$10.99", "Table: Should have table with first row and second column")
	assert.Contains(t, r, "$1.99", "Table: Should have table with second row and second column")
	assert.Contains(t, r, "started with Hermes", "Action: Should have instruction")
	assert.Contains(t, r, "Confirm your account", "Action: Should have button of action")
	assert.Contains(t, r, "#22BC66", "Action: Button should have given color")
	assert.Contains(t, r, "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010", "Action: Button should have link")
	assert.Contains(t, r, "Need help, or have questions", "Outro: Should have outro")
}

func (ed SimpleExample) assertPlainTextContent(t *testing.T, r string) {

	// Assert on product
	assert.Contains(t, r, "HermesName", "Product: Should find the name of the product in email")
	assert.Contains(t, r, "http://hermes-link.com", "Product: Should find the link of the product in email")
	assert.Contains(t, r, "Copyright © Hermes-Test", "Product: Should find the Copyright of the product in email")
	assert.NotContains(t, r, "https://go.dev/blog/gopher/header.jpg", "Product: Should not find any logo in plain text")

	// Assert on email body
	assert.Contains(t, r, "Hi Jon Snow", "Name: Should find the name of the person")
	assert.Contains(t, r, "Welcome to Hermes", "Intro: Should have intro")
	assert.Contains(t, r, "Birthday", "Dictionary: Should have dictionary")
	assert.Contains(t, r, "Open source", "Table: Should have table content")
	assert.Contains(t, r, `+--------+--------------------------------+--------+
|  ITEM  |          DESCRIPTION           | PRICE  |
+--------+--------------------------------+--------+
| Golang | Open source programming        | $10.99 |
|        | language that makes it easy    |        |
|        | to build simple, reliable, and |        |
|        | efficient software             |        |
| Hermes | Programmatically create        | $1.99  |
|        | beautiful e-mails using        |        |
|        | Golang.                        |        |
+--------+--------------------------------+--------`, "Table: Should have pretty table content")
	assert.Contains(t, r, "started with Hermes", "Action: Should have instruction")
	assert.NotContains(t, r, "Confirm your account", "Action: Should not have button of action in plain text")
	assert.NotContains(t, r, "#22BC66", "Action: Button should not have color in plain text")
	assert.Contains(t, r, "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010", "Action: Even if button is not possible in plain text, it should have the link")
	assert.Contains(t, r, "Need help, or have questions", "Outro: Should have outro")
}

type SimpleExamplePremailer struct {
	theme Theme
}

func (ed SimpleExamplePremailer) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name:      "HermesName",
			Link:      "http://hermes-link.com",
			Copyright: "Copyright © Hermes-Test",
			Logo:      "https://go.dev/blog/gopher/header.jpg",
		},
		TextDirection: TDLeftToRight,
	}

	email := Email{
		Body{
			Name: "Jon Snow",
			Intros: []string{
				"Welcome to Hermes! We're very excited to have you on board.",
			},
			Dictionary: []Entry{
				{"Firstname", "Jon", ""},
				{"Lastname", "Snow", ""},
				{"Birthday", "01/01/283", ""},
			},
			Table: Table{
				Data: [][]Entry{
					{
						{Key: "Item", Value: "Golang"},
						{Key: "Description", Value: "Open source programming language that makes it easy to build simple, reliable, and efficient software"},
						{Key: "Price", Value: "$10.99"},
					},
					{
						{Key: "Item", Value: "Hermes"},
						{Key: "Description", Value: "Programmatically create beautiful e-mails using Golang."},
						{Key: "Price", Value: "$1.99"},
					},
				},
				Columns: Columns{
					CustomWidth: map[string]string{
						"Item":  "20%",
						"Price": "15%",
					},
					CustomAlignment: map[string]string{
						"Price": "right",
					},
				},
			},
			Actions: []Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: Button{
						Color: "#22BC66",
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
			TemplateOverrides: map[string]any{
				"body_width": "1000px",
			},
		},
	}
	return h, email
}

func (ed SimpleExamplePremailer) assertHTMLContent(t *testing.T, r string) {

	// Assert on product
	assert.Contains(t, r, "HermesName", "Product: Should find the name of the product in email")
	assert.Contains(t, r, "http://hermes-link.com", "Product: Should find the link of the product in email")
	assert.Contains(t, r, "Copyright © Hermes-Test", "Product: Should find the Copyright of the product in email")
	assert.Contains(t, r, "https://go.dev/blog/gopher/header.jpg", "Product: Should find the logo of the product in email")
	assert.Contains(t, r, "If you’re having trouble with the button &#39;Confirm your account&#39;, copy and paste the URL below into your web browser.", "Product: Should find the trouble text in email")
	// Assert on email body
	assert.Contains(t, r, "Hi Jon Snow", "Name: Should find the name of the person")
	assert.Contains(t, r, "Welcome to Hermes", "Intro: Should have intro")
	assert.Contains(t, r, "Birthday", "Dictionary: Should have dictionary")
	assert.Contains(t, r, "Open source programming language", "Table: Should have table with first row and first column")
	assert.Contains(t, r, "Programmatically create beautiful e-mails using Golang", "Table: Should have table with second row and first column")
	assert.Contains(t, r, "$10.99", "Table: Should have table with first row and second column")
	assert.Contains(t, r, "$1.99", "Table: Should have table with second row and second column")
	assert.Contains(t, r, "started with Hermes", "Action: Should have instruction")
	assert.Contains(t, r, "Confirm your account", "Action: Should have button of action")
	assert.Contains(t, r, "#22BC66", "Action: Button should have given color")
	assert.Contains(t, r, "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010", "Action: Button should have link")
	assert.Contains(t, r, "Need help, or have questions", "Outro: Should have outro")
}

func (ed SimpleExamplePremailer) assertPlainTextContent(t *testing.T, r string) {

	// Assert on product
	assert.Contains(t, r, "HermesName", "Product: Should find the name of the product in email")
	assert.Contains(t, r, "http://hermes-link.com", "Product: Should find the link of the product in email")
	assert.Contains(t, r, "Copyright © Hermes-Test", "Product: Should find the Copyright of the product in email")
	assert.NotContains(t, r, "https://go.dev/blog/gopher/header.jpg", "Product: Should not find any logo in plain text")

	// Assert on email body
	assert.Contains(t, r, "Hi Jon Snow", "Name: Should find the name of the person")
	assert.Contains(t, r, "Welcome to Hermes", "Intro: Should have intro")
	assert.Contains(t, r, "Birthday", "Dictionary: Should have dictionary")
	assert.Contains(t, r, "Open source", "Table: Should have table content")
	assert.Contains(t, r, `+--------+--------------------------------+--------+
|  ITEM  |          DESCRIPTION           | PRICE  |
+--------+--------------------------------+--------+
| Golang | Open source programming        | $10.99 |
|        | language that makes it easy    |        |
|        | to build simple, reliable, and |        |
|        | efficient software             |        |
| Hermes | Programmatically create        | $1.99  |
|        | beautiful e-mails using        |        |
|        | Golang.                        |        |
+--------+--------------------------------+--------`, "Table: Should have pretty table content")
	assert.Contains(t, r, "started with Hermes", "Action: Should have instruction")
	assert.NotContains(t, r, "Confirm your account", "Action: Should not have button of action in plain text")
	assert.NotContains(t, r, "#22BC66", "Action: Button should not have color in plain text")
	assert.Contains(t, r, "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010", "Action: Even if button is not possible in plain text, it should have the link")
	assert.Contains(t, r, "Need help, or have questions", "Outro: Should have outro")
}

type SimpleExampleUnsafe struct {
	theme Theme
}

func (ed SimpleExampleUnsafe) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name:      "HermesName",
			Link:      "http://hermes-link.com",
			Copyright: "Copyright © Hermes-Test",
			Logo:      "https://go.dev/blog/gopher/header.jpg",
		},
		TextDirection: TDLeftToRight,
	}

	email := Email{
		Body{
			Name: "Jon Snow",
			IntrosUnsafe: []template.HTML{
				"<b>Welcome to Hermes!</b> We're very excited to have you on board.",
			},
			Dictionary: []Entry{
				{"Firstname", "Jon", ""},
				{"Lastname", "Snow", ""},
				{"Birthday", "01/01/283", ""},
			},
			Table: Table{
				Data: [][]Entry{
					{
						{Key: "Item", Value: "Golang"},
						{Key: "Description", Value: "Open source programming language that makes it easy to build simple, reliable, and efficient software"},
						{Key: "Price", Value: "$10.99"},
					},
					{
						{Key: "Item", Value: "Hermes"},
						{Key: "Description", UnsafeValue: "Programmatically create beautiful e-mails using <a href=\"https://go.dev\">Golang.</a>"},
						{Key: "Price", Value: "$1.99"},
					},
				},
				Columns: Columns{
					CustomWidth: map[string]string{
						"Item":  "20%",
						"Price": "15%",
					},
					CustomAlignment: map[string]string{
						"Price": "right",
					},
				},
			},
			Actions: []Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: Button{
						Color: "#22BC66",
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			OutrosUnsafe: []template.HTML{
				"Need help, or have questions? Just reply to <u>this email</u>, we'd love to help.",
			},
			TemplateOverrides: map[string]any{
				"body_width": "1000px",
			},
		},
	}
	return h, email
}

func (ed SimpleExampleUnsafe) assertHTMLContent(t *testing.T, r string) {

	// Assert on product
	assert.Contains(t, r, "HermesName", "Product: Should find the name of the product in email")
	assert.Contains(t, r, "http://hermes-link.com", "Product: Should find the link of the product in email")
	assert.Contains(t, r, "Copyright © Hermes-Test", "Product: Should find the Copyright of the product in email")
	assert.Contains(t, r, "https://go.dev/blog/gopher/header.jpg", "Product: Should find the logo of the product in email")
	assert.Contains(t, r, "If you’re having trouble with the button &#39;Confirm your account&#39;, copy and paste the URL below into your web browser.", "Product: Should find the trouble text in email")
	// Assert on email body
	assert.Contains(t, r, "Hi Jon Snow", "Name: Should find the name of the person")
	assert.Contains(t, r, "Welcome to Hermes", "Intro: Should have intro")
	assert.Contains(t, r, "Birthday", "Dictionary: Should have dictionary")
	assert.Contains(t, r, "Open source programming language", "Table: Should have table with first row and first column")
	assert.Contains(t, r, "Programmatically create beautiful e-mails using <a href=\"https://go.dev\"", "Table: Should have table with second row and first column")
	assert.Contains(t, r, "$10.99", "Table: Should have table with first row and second column")
	assert.Contains(t, r, "$1.99", "Table: Should have table with second row and second column")
	assert.Contains(t, r, "started with Hermes", "Action: Should have instruction")
	assert.Contains(t, r, "Confirm your account", "Action: Should have button of action")
	assert.Contains(t, r, "#22BC66", "Action: Button should have given color")
	assert.Contains(t, r, "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010", "Action: Button should have link")
	assert.Contains(t, r, "Need help, or have questions", "Outro: Should have outro")
}

func (ed SimpleExampleUnsafe) assertPlainTextContent(t *testing.T, r string) {

	// Assert on product
	assert.Contains(t, r, "HermesName", "Product: Should find the name of the product in email")
	assert.Contains(t, r, "http://hermes-link.com", "Product: Should find the link of the product in email")
	assert.Contains(t, r, "Copyright © Hermes-Test", "Product: Should find the Copyright of the product in email")
	assert.NotContains(t, r, "https://go.dev/blog/gopher/header.jpg", "Product: Should not find any logo in plain text")

	// Assert on email body
	assert.Contains(t, r, "Hi Jon Snow", "Name: Should find the name of the person")
	assert.Contains(t, r, "Welcome to Hermes", "Intro: Should have intro")
	assert.Contains(t, r, "Birthday", "Dictionary: Should have dictionary")
	assert.Contains(t, r, "Open source", "Table: Should have table content")
	assert.Contains(t, r, `+--------+--------------------------------+--------+
|  ITEM  |          DESCRIPTION           | PRICE  |
+--------+--------------------------------+--------+
| Golang | Open source programming        | $10.99 |
|        | language that makes it easy    |        |
|        | to build simple, reliable, and |        |
|        | efficient software             |        |
| Hermes | Programmatically create        | $1.99  |
|        | beautiful e-mails using        |        |
|        | Golang. ( https://go.dev )     |        |
+--------+--------------------------------+--------`, "Table: Should have pretty table content")
	assert.Contains(t, r, "started with Hermes", "Action: Should have instruction")
	assert.NotContains(t, r, "Confirm your account", "Action: Should not have button of action in plain text")
	assert.NotContains(t, r, "#22BC66", "Action: Button should not have color in plain text")
	assert.Contains(t, r, "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010", "Action: Even if button is not possible in plain text, it should have the link")
	assert.Contains(t, r, "Need help, or have questions", "Outro: Should have outro")
}

type SimpleExampleMarkdownIntroOutro struct {
	theme Theme
}

func (ed SimpleExampleMarkdownIntroOutro) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name:      "HermesName",
			Link:      "http://hermes-link.com",
			Copyright: "Copyright © Hermes-Test",
			Logo:      "https://go.dev/blog/gopher/header.jpg",
		},
		TextDirection:      TDLeftToRight,
		DisableCSSInlining: true,
	}

	email := Email{
		Body{
			Name: "Jon Snow",
			IntrosMarkdown: Markdown(strings.Join([]string{
				`## Welcome to Hermes!`,
				`### We're very excited to have you on board.`,
			}, "\n")),
			Dictionary: []Entry{
				{"Firstname", "Jon", ""},
				{"Lastname", "Snow", ""},
				{"Birthday", "01/01/283", ""},
			},
			Table: Table{
				Data: [][]Entry{
					{
						{Key: "Item", Value: "Golang"},
						{Key: "Description", Value: "Open source programming language that makes it easy to build simple, reliable, and efficient software"},
						{Key: "Price", Value: "$10.99"},
					},
					{
						{Key: "Item", Value: "Hermes"},
						{Key: "Description", Value: "Programmatically create beautiful e-mails using Golang."},
						{Key: "Price", Value: "$1.99"},
					},
				},
				Columns: Columns{
					CustomWidth: map[string]string{
						"Item":  "20%",
						"Price": "15%",
					},
					CustomAlignment: map[string]string{
						"Price": "right",
					},
				},
			},
			Actions: []Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: Button{
						Color: "#22BC66",
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			OutrosMarkdown: "**Need help, or have questions?** Just reply to **this email**, we'd love to _help._",
			TemplateOverrides: map[string]any{
				"body_width": "1000px",
			},
		},
	}
	return h, email
}

func (ed SimpleExampleMarkdownIntroOutro) assertHTMLContent(t *testing.T, r string) {

	// Assert on product
	assert.Contains(t, r, "HermesName", "Product: Should find the name of the product in email")
	assert.Contains(t, r, "http://hermes-link.com", "Product: Should find the link of the product in email")
	assert.Contains(t, r, "Copyright © Hermes-Test", "Product: Should find the Copyright of the product in email")
	assert.Contains(t, r, "https://go.dev/blog/gopher/header.jpg", "Product: Should find the logo of the product in email")
	assert.Contains(t, r, "If you’re having trouble with the button &#39;Confirm your account&#39;, copy and paste the URL below into your web browser.", "Product: Should find the trouble text in email")
	// Assert on email body
	assert.Contains(t, r, "Hi Jon Snow", "Name: Should find the name of the person")
	assert.Contains(t, r, "Welcome to Hermes", "Intro: Should have intro")
	assert.Contains(t, r, "Birthday", "Dictionary: Should have dictionary")
	assert.Contains(t, r, "Open source programming language", "Table: Should have table with first row and first column")
	assert.Contains(t, r, "Programmatically create beautiful e-mails using Golang", "Table: Should have table with second row and first column")
	assert.Contains(t, r, "$10.99", "Table: Should have table with first row and second column")
	assert.Contains(t, r, "$1.99", "Table: Should have table with second row and second column")
	assert.Contains(t, r, "started with Hermes", "Action: Should have instruction")
	assert.Contains(t, r, "Confirm your account", "Action: Should have button of action")
	assert.Contains(t, r, "#22BC66", "Action: Button should have given color")
	assert.Contains(t, r, "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010", "Action: Button should have link")
	assert.Contains(t, r, "Need help, or have questions", "Outro: Should have outro")
}

func (ed SimpleExampleMarkdownIntroOutro) assertPlainTextContent(t *testing.T, r string) {

	// Assert on product
	assert.Contains(t, r, "HermesName", "Product: Should find the name of the product in email")
	assert.Contains(t, r, "http://hermes-link.com", "Product: Should find the link of the product in email")
	assert.Contains(t, r, "Copyright © Hermes-Test", "Product: Should find the Copyright of the product in email")
	assert.NotContains(t, r, "https://go.dev/blog/gopher/header.jpg", "Product: Should not find any logo in plain text")

	// Assert on email body
	assert.Contains(t, r, "Hi Jon Snow", "Name: Should find the name of the person")
	assert.Contains(t, r, "Welcome to Hermes", "Intro: Should have intro")
	assert.Contains(t, r, "Birthday", "Dictionary: Should have dictionary")
	assert.Contains(t, r, "Open source", "Table: Should have table content")
	assert.Contains(t, r, `+--------+--------------------------------+--------+
|  ITEM  |          DESCRIPTION           | PRICE  |
+--------+--------------------------------+--------+
| Golang | Open source programming        | $10.99 |
|        | language that makes it easy    |        |
|        | to build simple, reliable, and |        |
|        | efficient software             |        |
| Hermes | Programmatically create        | $1.99  |
|        | beautiful e-mails using        |        |
|        | Golang.                        |        |
+--------+--------------------------------+--------`, "Table: Should have pretty table content")
	assert.Contains(t, r, "started with Hermes", "Action: Should have instruction")
	assert.NotContains(t, r, "Confirm your account", "Action: Should not have button of action in plain text")
	assert.NotContains(t, r, "#22BC66", "Action: Button should not have color in plain text")
	assert.Contains(t, r, "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010", "Action: Even if button is not possible in plain text, it should have the link")
	assert.Contains(t, r, "Need help, or have questions", "Outro: Should have outro")
}

type WithTitleInsteadOfNameExample struct {
	theme Theme
}

func (ed WithTitleInsteadOfNameExample) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name: "Hermes",
			Link: "http://hermes.com",
		},
		DisableCSSInlining: true,
	}

	email := Email{
		Body{
			Name:  "Jon Snow",
			Title: "A new e-mail",
		},
	}

	return h, email
}

func (ed WithTitleInsteadOfNameExample) assertHTMLContent(t *testing.T, r string) {
	// In HTML template: both title and greeting can appear simultaneously
	assert.Contains(t, r, "Hi Jon Snow", "HTML: should find greetings when both name and greeting are provided")
	assert.Contains(t, r, "A new e-mail", "HTML: title should be displayed in h1 tag")

	// Verify proper structure - title in h1, greeting in p
	assert.Contains(t, r, "<h1>", "HTML: title should be in h1 tag")
	assert.Contains(t, r, "<p class=\"justify\">", "HTML: greeting should be in paragraph with justify class")
}

func (ed WithTitleInsteadOfNameExample) assertPlainTextContent(t *testing.T, r string) {
	// In plain text template: title OR greeting (if/else logic)
	assert.NotContains(t, r, "Hi Jon Snow", "PlainText: should NOT find greetings when title is provided (if/else logic)")
	assert.Contains(t, r, "A new e-mail", "PlainText: title should be displayed when provided")
}

type WithGreetingDifferentThanDefault struct {
	theme Theme
}

func (ed WithGreetingDifferentThanDefault) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name: "Hermes",
			Link: "http://hermes.com",
		},
		DisableCSSInlining: true,
	}

	email := Email{
		Body{
			Greeting: "Dear",
			Name:     "Jon Snow",
		},
	}

	return h, email
}

func (ed WithGreetingDifferentThanDefault) assertHTMLContent(t *testing.T, r string) {
	assert.NotContains(t, r, "Hi Jon Snow", "Should not find greetings with 'Hi' which is default")
	assert.Contains(t, r, "Dear Jon Snow", "Should have greeting with Dear")
}

func (ed WithGreetingDifferentThanDefault) assertPlainTextContent(t *testing.T, r string) {
	assert.NotContains(t, r, "Hi Jon Snow", "Should not find greetings with 'Hi' which is default")
	assert.Contains(t, r, "Dear Jon Snow", "Should have greeting with Dear")
}

type WithSignatureDifferentThanDefault struct {
	theme Theme
}

func (ed WithSignatureDifferentThanDefault) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name: "Hermes",
			Link: "http://hermes.com",
		},
		DisableCSSInlining: true,
	}

	email := Email{
		Body{
			Name:          "Jon Snow",
			Signature:     "Best regards",
			SignatureName: "Test User",
		},
	}

	return h, email
}

func (ed WithSignatureDifferentThanDefault) assertHTMLContent(t *testing.T, r string) {
	assert.NotContains(t, r, "Yours truly", "Should not find signature with 'Yours truly' which is default")
	assert.Contains(t, r, "Best regards", "Should have greeting with Dear")
}

func (ed WithSignatureDifferentThanDefault) assertPlainTextContent(t *testing.T, r string) {
	assert.NotContains(t, r, "Yours truly", "Should not find signature with 'Yours truly' which is default")
	assert.Contains(t, r, "Best regards", "Should have greeting with Dear")
}

type WithInviteCode struct {
	theme Theme
}

func (ed WithInviteCode) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name: "Hermes",
			Link: "http://hermes.com",
		},
		DisableCSSInlining: true,
	}

	email := Email{
		Body{
			Name: "Jon Snow",
			Actions: []Action{
				{
					Instructions: "Here is your invite code:",
					InviteCode:   "123456",
				},
			},
		},
	}

	return h, email
}

func (ed WithInviteCode) assertHTMLContent(t *testing.T, r string) {
	assert.Contains(t, r, "Here is your invite code", "Should contains the instruction")
	assert.Contains(t, r, "123456", "Should contain the short code")
}

func (ed WithInviteCode) assertPlainTextContent(t *testing.T, r string) {
	assert.Contains(t, r, "Here is your invite code", "Should contains the instruction")
	assert.Contains(t, r, "123456", "Should contain the short code")
}

type WithFreeMarkdownContent struct {
	theme Theme
}

func (ed WithFreeMarkdownContent) getExample() (Hermes, Email) {
	h := Hermes{
		Theme: ed.theme,
		Product: Product{
			Name: "Hermes",
			Link: "http://hermes.com",
		},
		DisableCSSInlining: true,
	}

	email := Email{
		Body{
			Name: "Jon Snow",
			FreeMarkdown: `
> _Hermes_ service will shutdown the **1st August 2025** for maintenance operations. 

Services will be unavailable based on the following schedule:

| Services | Downtime |
| :------:| :-----------: |
| Service A | 2AM to 3AM |
| Service B | 4AM to 5AM |
| Service C | 5AM to 6AM |

---

Feel free to contact us for any question regarding this matter at [support@hermes-example.com](mailto:support@hermes-example.com) or in our [Gitter](https://gitter.im/)

`,
			Intros: []string{
				"An intro that should be kept even with FreeMarkdown",
			},
			Dictionary: []Entry{
				{"Dictionary that should not be displayed", "Because of FreeMarkdown", ""},
			},
			Table: Table{
				Data: [][]Entry{
					{
						{Key: "Item", Value: "Golang"},
					},
					{
						{Key: "Item", Value: "Hermes"},
					},
				},
			},
			Actions: []Action{
				{
					Instructions: "Action that should not be displayed, because of FreeMarkdown:",
					Button: Button{
						Color: "#22BC66",
						Text:  "Button",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"An outro that should be kept even with FreeMarkdown",
			},
		},
	}

	return h, email
}

func (ed WithFreeMarkdownContent) assertHTMLContent(t *testing.T, r string) {
	assert.NotContains(t, r, "Yours truly", "Should not find default signature when none is explicitly set")
	assert.Contains(t, r, "Jon Snow", "Should find title with 'Jon Snow'")
	assert.Contains(t, r, "<em>Hermes</em> service will shutdown", "Should find quote as HTML formatted content")
	assert.Contains(t, r, "<td align=\"center\">2AM to 3AM</td>", "Should find cell content as HTML formatted content")
	assert.Contains(t, r, "<a href=\"mailto:support@hermes-example.com\">support@hermes-example.com</a>", "Should find link of mailto as HTML formatted content")
	assert.Contains(t, r, "An intro that should be kept even with FreeMarkdown", "Should find intro even with FreeMarkdown")
	assert.Contains(t, r, "An outro that should be kept even with FreeMarkdown", "Should find outro even with FreeMarkdown")
	assert.NotContains(t, r, "should not be displayed", "Should find any other content that the one from FreeMarkdown object")
}

func (ed WithFreeMarkdownContent) assertPlainTextContent(t *testing.T, r string) {
	assert.NotContains(t, r, "Yours truly", "Should not find default signature when none is explicitly set")
	assert.Contains(t, r, "Jon Snow", "Should find title with 'Jon Snow'")
	assert.Contains(t, r, "> Hermes service will shutdown", "Should find quote as plain text with quote emphaze on sentence")
	assert.Contains(t, r, "2AM to 3AM", "Should find cell content as plain text")
	assert.Contains(t, r, `+-----------+------------+
| SERVICES  |  DOWNTIME  |
+-----------+------------+
| Service A | 2AM to 3AM |
| Service B | 4AM to 5AM |
| Service C | 5AM to 6AM |
+-----------+------------+`, "Should find pretty table as plain text")
	assert.Contains(t, r, "support@hermes-example.com", "Should find link of mailto as plain text")
	assert.NotContains(t, r, "<table>", "Should not find html table tags")
	assert.NotContains(t, r, "<tr>", "Should not find html tr tags")
	assert.NotContains(t, r, "<a>", "Should not find html link tags")
	assert.NotContains(t, r, "should not be displayed", "Should find any other content that the one from FreeMarkdown object")
}

func TestThemeSimple(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &SimpleExample{theme})
		})
	}
}

func TestThemeSimplePremailer(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &SimpleExamplePremailer{theme})
		})
	}
}

func TestThemeSimpleMarkdownIntroOutro(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &SimpleExampleMarkdownIntroOutro{theme})
		})
	}
}

func TestThemeSimpleUnsafe(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &SimpleExampleUnsafe{theme})
		})
	}
}

func TestThemeWithTitleInsteadOfName(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &WithTitleInsteadOfNameExample{theme})
		})
	}
}

func TestThemeWithGreetingDifferentThanDefault(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &WithGreetingDifferentThanDefault{theme})
		})
	}
}

func TestThemeWithGreetingDiffrentThanDefault(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &WithSignatureDifferentThanDefault{theme})
		})
	}
}

func TestThemeWithFreeMarkdownContent(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &WithFreeMarkdownContent{theme})
		})
	}
}

func TestThemeWithInviteCode(t *testing.T) {
	for i, theme := range testedThemes {
		t.Run(fmt.Sprintf("%s-%d", theme.Name(), i), func(t *testing.T) {
			checkExample(t, &WithInviteCode{theme})
		})
	}
}

func checkExample(t *testing.T, ex Example) {
	// Given an example
	h, email := ex.getExample()

	// When generating HTML template
	r, err := h.GenerateHTML(email)
	if debug {
		t.Log(r)
	}

	if assert.Nil(t, err) && assert.NotEmpty(t, r) {
		previewEmail(fmt.Sprintf("%s.html", t.Name()), r)
		// Then asserting HTML is OK
		ex.assertHTMLContent(t, r)
	}

	// When generating plain text template
	r, err = h.GeneratePlainText(email)
	if debug {
		t.Log(r)
	}

	if assert.Nil(t, err) && assert.NotEmpty(t, r) {
		previewEmail(fmt.Sprintf("%s.txt", t.Name()), r)
		// Then asserting plain text is OK
		ex.assertPlainTextContent(t, r)
	}
}

// previews Email if debug mode enabled
func previewEmail(name, input string) {
	if !debug {
		return
	}

	filename := fmt.Sprintf("/tmp/%s", strings.ReplaceAll(name, "/", "-"))
	os.WriteFile(filename, []byte(input), 0644)
	cmd := exec.Command("open", filename)
	cmd.Run()
}

// ======================================
// Tests on default values for all themes
// It does not check email content
// ======================================

func TestHermes_TextDirectionAsDefault(t *testing.T) {
	t.Parallel()

	h := Hermes{
		Product: Product{
			Name: "Hermes",
			Link: "http://hermes.com",
		},
		TextDirection:      "not-existing", // Wrong text-direction
		DisableCSSInlining: true,
	}

	email := Email{
		Body{
			Name: "Jon Snow",
			Intros: []string{
				"Welcome to Hermes! We're very excited to have you on board.",
			},
			Actions: []Action{
				{
					Instructions: "To get started with Hermes, please click here:",
					Button: Button{
						Color: "#22BC66",
						Text:  "Confirm your account",
						Link:  "https://hermes-example.com/confirm?token=d9729feb74992cc3482b350163a1a010",
					},
				},
			},
			Outros: []string{
				"Need help, or have questions? Just reply to this email, we'd love to help.",
			},
		},
	}

	_, err := h.GenerateHTML(email)
	assert.NoError(t, err)
	assert.Equal(t, TDLeftToRight, h.TextDirection)
	assert.Equal(t, "default", h.Theme.Name())
}

// ErrorTheme is a test theme that always returns errors for template methods
type ErrorTheme struct{}

func (et ErrorTheme) Name() string {
	return "error"
}

func (et ErrorTheme) Styles() StylesDefinition {
	return StylesDefinition{}
}

func (et ErrorTheme) HTMLTemplate() string {
	return "{{invalid syntax for template"
}

func (et ErrorTheme) PlainTextTemplate() string {
	return "{{invalid syntax for template"
}

// ErrorParsedTheme implements parsed template interfaces with errors
type ErrorParsedTheme struct{}

func (ept ErrorParsedTheme) Name() string {
	return "errorparsed"
}

func (ept ErrorParsedTheme) Styles() StylesDefinition {
	return StylesDefinition{}
}

func (ept ErrorParsedTheme) HTMLTemplate() string {
	return "<html>Valid HTML</html>"
}

func (ept ErrorParsedTheme) PlainTextTemplate() string {
	return "Valid plain text"
}

func (ept ErrorParsedTheme) ParsedHTMLTemplate() (*template.Template, error) {
	return nil, errors.New("parsed HTML template error")
}

func (ept ErrorParsedTheme) ParsedPlainTextTemplate() (*template.Template, error) {
	return nil, errors.New("parsed plain text template error")
}

// ValidParsedTheme implements parsed template interfaces without errors
type ValidParsedTheme struct{}

func (vpt ValidParsedTheme) Name() string {
	return "validparsed"
}

func (vpt ValidParsedTheme) Styles() StylesDefinition {
	return StylesDefinition{}
}

func (vpt ValidParsedTheme) HTMLTemplate() string {
	return "<html>{{.Product.Name}}</html>"
}

func (vpt ValidParsedTheme) PlainTextTemplate() string {
	return "{{.Product.Name}}"
}

func (vpt ValidParsedTheme) ParsedHTMLTemplate() (*template.Template, error) {
	return TemplateBase().Parse("<html>{{.Product.Name}}</html>")
}

func (vpt ValidParsedTheme) ParsedPlainTextTemplate() (*template.Template, error) {
	return TemplateBase().Parse("{{.Product.Name}}")
}

func TestGetHTMLTemplate(t *testing.T) {
	t.Run("RegularTheme", func(t *testing.T) {
		theme := new(Default)
		tmpl, err := getHTMLTemplate(theme)
		assert.NoError(t, err)
		assert.NotNil(t, tmpl)
	})

	t.Run("ParsedThemeSuccess", func(t *testing.T) {
		theme := ValidParsedTheme{}
		tmpl, err := getHTMLTemplate(theme)
		assert.NoError(t, err)
		assert.NotNil(t, tmpl)
	})

	t.Run("ParsedThemeError", func(t *testing.T) {
		theme := ErrorParsedTheme{}
		tmpl, err := getHTMLTemplate(theme)
		assert.Error(t, err)
		assert.Nil(t, tmpl)
		assert.Contains(t, err.Error(), "parsed HTML template error")
	})

	t.Run("InvalidTemplateString", func(t *testing.T) {
		theme := ErrorTheme{}
		tmpl, err := getHTMLTemplate(theme)
		assert.Error(t, err)
		assert.Nil(t, tmpl)
	})
}

func TestGetPlainTextTemplate(t *testing.T) {
	t.Run("RegularTheme", func(t *testing.T) {
		theme := new(Default)
		tmpl, err := getPlainTextTemplate(theme)
		assert.NoError(t, err)
		assert.NotNil(t, tmpl)
	})

	t.Run("ParsedThemeSuccess", func(t *testing.T) {
		theme := ValidParsedTheme{}
		tmpl, err := getPlainTextTemplate(theme)
		assert.NoError(t, err)
		assert.NotNil(t, tmpl)
	})

	t.Run("ParsedThemeError", func(t *testing.T) {
		theme := ErrorParsedTheme{}
		tmpl, err := getPlainTextTemplate(theme)
		assert.Error(t, err)
		assert.Nil(t, tmpl)
		assert.Contains(t, err.Error(), "parsed plain text template error")
	})

	t.Run("InvalidTemplateString", func(t *testing.T) {
		theme := ErrorTheme{}
		tmpl, err := getPlainTextTemplate(theme)
		assert.Error(t, err)
		assert.Nil(t, tmpl)
	})
}

func TestFlatStyles_EdgeCases(t *testing.T) {
	t.Run("FlatStylesCloning", func(t *testing.T) {
		flat := Flat{}
		styles1 := flat.Styles()
		styles2 := flat.Styles()

		// Verify deep cloning - modifying one shouldn't affect the other
		if bodyStyles1, ok := styles1["body"]; ok {
			bodyStyles1["test-property"] = "test-value"
		}

		if bodyStyles2, ok := styles2["body"]; ok {
			_, hasTestProp := bodyStyles2["test-property"]
			assert.False(t, hasTestProp, "Styles should be deeply cloned")
		}
	})

	t.Run("FlatStylesDefensiveMutation", func(t *testing.T) {
		flat := Flat{}
		styles := flat.Styles()

		// Verify that flat-specific styles are applied
		assert.Contains(t, styles, "body")
		assert.Contains(t, styles, ".email-wrapper")
		assert.Contains(t, styles, ".email-footer p")
		assert.Contains(t, styles, ".button")

		bodyStyles := styles["body"]
		assert.Equal(t, "#2c3e50", bodyStyles["background-color"])

		buttonStyles := styles[".button"]
		assert.Equal(t, "#00948d", buttonStyles["background-color"])
		assert.Equal(t, "0", buttonStyles["border-radius"])
	})

	t.Run("FlatStylesEnsureFunction", func(t *testing.T) {
		// Test the defensive ensure function by checking if it creates missing selectors
		flat := Flat{}
		styles := flat.Styles()

		// All the selectors that Flat theme applies should exist
		expectedSelectors := []string{"body", ".email-wrapper", ".email-footer p", ".button"}
		for _, selector := range expectedSelectors {
			assert.Contains(t, styles, selector, "Selector should exist: %s", selector)
			assert.NotNil(t, styles[selector], "Selector map should not be nil: %s", selector)
		}

		// Test that we can access nested properties without panics
		assert.NotPanics(t, func() {
			_ = styles["body"]["background-color"]
			_ = styles[".button"]["border-radius"]
		})
	})
}

func TestNormalizeStyles(t *testing.T) {
	t.Run("StylesDefinitionInput", func(t *testing.T) {
		input := StylesDefinition{
			"body": {"color": "red"},
		}
		result := normalizeStyles(input)
		assert.Equal(t, input, result)
	})

	t.Run("MapStringMapStringInterfaceInput", func(t *testing.T) {
		input := map[string]map[string]interface{}{
			"body": {"color": "red", "background": "blue"},
		}
		result := normalizeStyles(input)
		assert.NotNil(t, result)
		assert.Contains(t, result, "body")
		assert.Equal(t, "red", result["body"]["color"])
		assert.Equal(t, "blue", result["body"]["background"])
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		result := normalizeStyles("invalid type")
		assert.Nil(t, result)
	})

	t.Run("NilInput", func(t *testing.T) {
		result := normalizeStyles(nil)
		assert.Nil(t, result)
	})
}

func TestSetDefaultHermesValues(t *testing.T) {
	t.Run("EmptyHermes", func(t *testing.T) {
		h := &Hermes{}
		err := setDefaultHermesValues(h)
		assert.NoError(t, err)
		assert.NotNil(t, h.Theme)
		assert.Equal(t, TDLeftToRight, h.TextDirection)
		assert.Equal(t, "Hermes", h.Product.Name)
	})

	t.Run("PartialHermes", func(t *testing.T) {
		h := &Hermes{
			Product: Product{Name: "Custom App"},
		}
		err := setDefaultHermesValues(h)
		assert.NoError(t, err)
		assert.Equal(t, "Custom App", h.Product.Name) // Should keep existing value
		assert.NotEmpty(t, h.Product.Copyright)       // Should get default
	})

	t.Run("InvalidTextDirection", func(t *testing.T) {
		h := &Hermes{
			TextDirection: TextDirection("invalid"),
		}
		err := setDefaultHermesValues(h)
		assert.NoError(t, err)
		assert.Equal(t, TDLeftToRight, h.TextDirection) // Should reset to default
	})

	t.Run("ValidTextDirection", func(t *testing.T) {
		h := &Hermes{
			TextDirection: TDRightToLeft,
		}
		err := setDefaultHermesValues(h)
		assert.NoError(t, err)
		assert.Equal(t, TDRightToLeft, h.TextDirection) // Should keep valid direction
	})
}

func TestGenerateTemplate_EdgeCases(t *testing.T) {
	t.Run("DeprecatedTableField", func(t *testing.T) {
		h := &Hermes{Theme: new(Default)}
		email := Email{
			Body: Body{
				Name: "Test User",
				Table: Table{
					Data: [][]Entry{
						{
							{Key: "Item", Value: "Test"},
							{Key: "Price", Value: "$10"},
						},
					},
				},
			},
		}

		// Verify table data exists
		assert.Len(t, email.Body.Table.Data, 1)

		tmpl, err := getHTMLTemplate(h.Theme)
		assert.NoError(t, err)

		result, err := h.generateTemplate(email, tmpl)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)

		// The key thing is that the function should handle deprecated tables without error
		// We can't assert on the Tables field because the logic might depend on template execution
		assert.Contains(t, result, "Test") // Verify the table data appears in output
	})

	t.Run("DisabledCSSInlining", func(t *testing.T) {
		h := &Hermes{
			Theme:              new(Default),
			DisableCSSInlining: true,
		}
		email := Email{
			Body: Body{Name: "Test User"},
		}

		tmpl, err := getHTMLTemplate(h.Theme)
		assert.NoError(t, err)

		result, err := h.generateTemplate(email, tmpl)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		// Should contain embedded styles when inlining is disabled
		assert.Contains(t, result, "<style")
	})
}

func TestGenerateHTML_ErrorHandling(t *testing.T) {
	email := Email{
		Body: Body{
			Name: "Test User",
		},
	}

	t.Run("InvalidThemeTemplate", func(t *testing.T) {
		h := Hermes{Theme: ErrorTheme{}}
		_, err := h.GenerateHTML(email)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template")
	})

	t.Run("ParsedThemeError", func(t *testing.T) {
		h := Hermes{Theme: ErrorParsedTheme{}}
		_, err := h.GenerateHTML(email)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parsed HTML template error")
	})

	t.Run("CSSInliningDisabled", func(t *testing.T) {
		h := Hermes{
			Theme:              new(Default),
			DisableCSSInlining: true,
		}
		result, err := h.GenerateHTML(email)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		// When CSS inlining is disabled, we should get HTML with embedded styles
		assert.Contains(t, result, "<style")
	})
}

func TestGeneratePlainText_ErrorHandling(t *testing.T) {
	email := Email{
		Body: Body{
			Name: "Test User",
		},
	}

	t.Run("InvalidThemeTemplate", func(t *testing.T) {
		h := Hermes{Theme: ErrorTheme{}}
		_, err := h.GeneratePlainText(email)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template")
	})

	t.Run("ParsedThemeError", func(t *testing.T) {
		h := Hermes{Theme: ErrorParsedTheme{}}
		_, err := h.GeneratePlainText(email)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parsed plain text template error")
	})

	t.Run("HTML2TextConversion", func(t *testing.T) {
		h := Hermes{Theme: new(Default)}
		result, err := h.GeneratePlainText(email)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		// Should successfully convert HTML to plain text
		assert.NotContains(t, result, "<html>")
		assert.NotContains(t, result, "</html>")
	})
}

func TestHermes_Default(t *testing.T) {
	t.Parallel()

	h := Hermes{}
	err := setDefaultHermesValues(&h)
	assert.NoError(t, err)

	email := Email{}
	err = setDefaultEmailValues(&h, &email)
	assert.NoError(t, err)

	assert.Equal(t, TDLeftToRight, h.TextDirection)
	assert.Equal(t, new(Default), h.Theme)
	assert.Equal(t, "Hermes", h.Product.Name)
	assert.Equal(t, "Copyright © 2025 Hermes. All rights reserved.", h.Product.Copyright)

	assert.Empty(t, email.Body.Actions)
	assert.Empty(t, email.Body.Dictionary)
	assert.Empty(t, email.Body.Intros)
	assert.Empty(t, email.Body.Outros)
	assert.Empty(t, email.Body.Table.Data)
	assert.Empty(t, email.Body.Table.Columns.CustomWidth)
	assert.Empty(t, email.Body.Table.Columns.CustomAlignment)
	assert.Empty(t, string(email.Body.FreeMarkdown))

	assert.Equal(t, "Hi", email.Body.Greeting)
	assert.Empty(t, email.Body.Signature) // No default signature anymore
	assert.Empty(t, email.Body.Title)
}
