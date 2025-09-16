package hermes

import (
	"bytes"
	"html/template"

	"dario.cat/mergo"
	"github.com/Masterminds/sprig/v3"
	"github.com/jaytaylor/html2text"
	"github.com/russross/blackfriday/v2"
	"github.com/sirupsen/logrus"
	"github.com/vanng822/go-premailer/premailer"
)

// Hermes is an instance of the hermes email generator
type Hermes struct {
	Theme              Theme
	TextDirection      TextDirection
	Product            Product
	DisableCSSInlining bool
}

type ThemedTemplate interface {
	Theme
	ParsedHTMLTheme
	ParsedPlainTextTheme
}

// Theme is an interface to implement when creating a new theme
type Theme interface {
	Name() string              // The name of the theme
	HTMLTemplate() string      // The golang template for HTML emails
	PlainTextTemplate() string // The golang templte for plain text emails (can be basic HTML)
}

// ParsedHTMLTheme is implemented by themes that parse their HTML
// template themselves.
type ParsedHTMLTheme interface {
	ParsedHTMLTemplate() (*template.Template, error)
}

// ParsedPlainTextTheme is implemented by themes that parse their
// plain text template themselves.
type ParsedPlainTextTheme interface {
	ParsedPlainTextTemplate() (*template.Template, error)
}

// TextDirection of the text in HTML email
type TextDirection string

var templateFuncs = template.FuncMap{
	"url": func(s string) template.URL {
		return template.URL(s)
	},
	"css": func(in any) template.CSS {
		s, ok := in.(string)
		if !ok {
			return ""
		}

		return template.CSS(s)
	},
}

// TDLeftToRight is the text direction from left to right (default)
const TDLeftToRight TextDirection = "ltr"

// TDRightToLeft is the text direction from right to left
const TDRightToLeft TextDirection = "rtl"

// Product represents your company product (brand)
// Appears in header & footer of e-mails
type Product struct {
	Name      string
	Link      string // e.g. https://matcornic.github.io
	Logo      string // e.g. https://matcornic.github.io/img/logo.png
	Copyright string // Copyright © 2019 Hermes. All rights reserved.
	// TroubleText is the sentence at the end of the email for users having trouble with the button
	// (default to `If you’re having trouble with the button '{ACTION}',
	// copy and paste the URL below into your web browser.`)
	TroubleText string
}

// Email is the email containing a body
type Email struct {
	Body Body
}

// Markdown is a HTML template (a string) representing Markdown content
// https://en.wikipedia.org/wiki/Markdown
type Markdown template.HTML

// Body is the body of the email, containing all interesting data
type Body struct {
	Name              string          // The name of the contacted person
	Intros            []string        // Intro sentences, first displayed in the email
	IntrosMarkdown    Markdown        // Intro in markdown, will override Intros
	IntrosUnsafe      []template.HTML // IntrosUnsafe is a list of unsafe HTML intro sentences
	Dictionary        []Entry         // A list of key+value (useful for displaying parameters/settings/personal info)
	Table             Table           // (DEPRECATED: Use Tables field instead) Table is an table where you can put data (pricing grid, a bill, and so on)
	Tables            []Table         // Tables is a list of tables where you can put data (pricing grid, a bill, and so on)
	Actions           []Action        // Actions are a list of actions that the user will be able to execute via a button click
	OutrosMarkdown    Markdown        // Outro in markdown, will override Outros
	OutrosUnsafe      []template.HTML // OutrosUnsafe is a list of unsafe HTML outro sentences
	Outros            []string        // Outro sentences, last displayed in the email
	Greeting          string          // Greeting for the contacted person (default to 'Hi')
	Signature         string          // Signature for the contacted person (default to 'Yours truly')
	Title             string          // Title replaces the greeting+name when set
	FreeMarkdown      Markdown        // Free markdown content that replaces all content other than header and footer
	TemplateOverrides map[string]any  // TemplateOverrides is a map of key-value pairs that can be used to override the default template values or inject additional styles
}

// ToHTML converts Markdown to HTML
func (c Markdown) ToHTML() template.HTML {
	return template.HTML(blackfriday.Run([]byte(c)))
}

// Entry is a simple entry of a map
// Allows using a slice of entries instead of a map
// Because Golang maps are not ordered
type Entry struct {
	Key         string
	Value       string
	UnsafeValue template.HTML
}

// Table is an table where you can put data (pricing grid, a bill, and so on)
type Table struct {
	Title   string    // Title of the table
	Data    [][]Entry // Contains data
	Columns Columns   // Contains meta-data for display purpose (width, alignement)
}

// Columns contains meta-data for the different columns
type Columns struct {
	CustomWidth     map[string]string
	CustomAlignment map[string]string
}

// Action is anything the user can act on (i.e., click on a button, view an invite code)
type Action struct {
	Instructions string
	Button       Button
	InviteCode   string
}

// Button defines an action to launch
type Button struct {
	Color     string
	TextColor string
	Text      string
	Link      string
}

// Template is the struct given to Golang templating
// Root object in a template is this struct
type Template struct {
	Hermes Hermes
	Email  Email
}

func setDefaultEmailValues(h *Hermes, e *Email) error {
	// Default values of an email
	defaultEmail := Email{
		Body: Body{
			Intros:     []string{},
			Dictionary: []Entry{},
			Outros:     []string{},
			Signature:  "Yours truly",
			Greeting:   "Hi",
		},
	}
	// Merge the given email with default one
	// Default one overrides all zero values
	err := mergo.Merge(e, defaultEmail)
	if err != nil {
		return err
	}

	// Default CSS map for the template
	defaultCSS := map[string]map[string]interface{}{
		"*:not(br):not(tr):not(html)": {
			"font-family":        "Arial, 'Helvetica Neue', Helvetica, sans-serif",
			"-webkit-box-sizing": "border-box",
			"box-sizing":         "border-box",
		},
		"body": {
			"width":                    "100% !important",
			"height":                   "100%",
			"margin":                   "0",
			"line-height":              "1.4",
			"background-color":         "#F2F4F6",
			"color":                    "#74787E",
			"-webkit-text-size-adjust": "none",
		},
		"a": {
			"color": "#3869D4",
		},
		".email-wrapper": {
			"width":            "100%",
			"margin":           "0",
			"padding":          "0",
			"background-color": "#F2F4F6",
		},
		".email-content": {
			"width":   "100%",
			"margin":  "0",
			"padding": "0",
		},
		".email-masthead": {
			"padding":    "25px 0",
			"text-align": "center",
		},
		".email-masthead_logo": {
			"max-width": "400px",
			"border":    "0",
		},
		".email-masthead_name": {
			"font-size":       "16px",
			"font-weight":     "bold",
			"color":           "#2F3133",
			"text-decoration": "none",
			"text-shadow":     "0 1px 0 white",
		},
		".email-logo": {
			"max-height": "50px",
		},
		".email-body": {
			"width":            "100%",
			"margin":           "0",
			"padding":          "0",
			"border-top":       "1px solid #EDEFF2",
			"border-bottom":    "1px solid #EDEFF2",
			"background-color": "#FFF",
		},
		".email-body_inner": {
			"width":   "570px", // Will be updated based on body_width override
			"margin":  "0 auto",
			"padding": "0",
		},
		".email-footer": {
			"width":      "570px", // Will be updated based on body_width override
			"margin":     "0 auto",
			"padding":    "0",
			"text-align": "center",
		},
		".email-footer p": {
			"color": "#AEAEAE",
		},
		".body-action": {
			"width":      "100%",
			"margin":     "30px auto",
			"padding":    "0",
			"text-align": "center",
		},
		".body-dictionary": {
			"margin":   "20px auto 10px",
			"overflow": "hidden",
			"padding":  "0",
			"width":    "100%",
		},
		".body-dictionary dd": {
			"display":       "inline-block",
			"margin":        "0 0 10px 0",
			"width":         "49%",
			"margin-left":   "0",
			"margin-bottom": "10px",
		},
		".body-dictionary dt": {
			"clear":       "both",
			"color":       "#000",
			"display":     "inline-block",
			"font-weight": "bold",
			"width":       "49%",
		},
		".body-sub": {
			"margin-top":   "25px",
			"padding-top":  "25px",
			"border-top":   "1px solid #EDEFF2",
			"table-layout": "fixed",
		},
		".body-sub a": {
			"word-break": "break-all",
		},
		".content-cell": {
			"padding": "35px",
		},
		".align-right": {
			"text-align": "right",
		},
		"h1": {
			"margin-top":  "0",
			"color":       "#2F3133",
			"font-size":   "19px",
			"font-weight": "bold",
		},
		"h2": {
			"margin-top":  "0",
			"color":       "#2F3133",
			"font-size":   "16px",
			"font-weight": "bold",
		},
		"h3": {
			"margin-top":  "0",
			"color":       "#2F3133",
			"font-size":   "14px",
			"font-weight": "bold",
		},
		"blockquote": {
			"margin":       "25px 0",
			"padding-left": "10px",
			"border-left":  "10px solid #F0F2F4",
		},
		"blockquote p": {
			"font-size": "1.1rem",
			"color":     "#999",
		},
		"blockquote cite": {
			"display":    "block",
			"text-align": "right",
			"color":      "#666",
			"font-size":  "1.2rem",
		},
		"cite": {
			"display":   "block",
			"font-size": "0.925rem",
		},
		"cite:before": {
			"content": "\"\\2014 \\0020\"",
		},
		"p": {
			"margin-top":  "0",
			"color":       "#74787E",
			"font-size":   "16px",
			"line-height": "1.5em",
		},
		"p.sub": {
			"font-size": "12px",
		},
		"p.center": {
			"text-align": "center",
		},
		"table": {
			"width": "100%",
		},
		"th": {
			"padding":        "0px 5px",
			"padding-bottom": "8px",
			"border-bottom":  "1px solid #EDEFF2",
		},
		"th p": {
			"margin":    "0",
			"color":     "#9BA2AB",
			"font-size": "12px",
		},
		"td": {
			"padding":     "10px 5px",
			"color":       "#74787E",
			"font-size":   "15px",
			"line-height": "18px",
		},
		".content": {
			"align":   "center",
			"padding": "0",
		},
		".data-wrapper": {
			"width":   "100%",
			"margin":  "0",
			"padding": "35px 0",
		},
		".data-table": {
			"width":  "100%",
			"margin": "0",
		},
		".data-table th": {
			"text-align":     "left",
			"padding":        "0px 5px",
			"padding-bottom": "8px",
			"border-bottom":  "1px solid #EDEFF2",
		},
		".data-table th p": {
			"margin":    "0",
			"color":     "#9BA2AB",
			"font-size": "12px",
		},
		".data-table td": {
			"padding":     "10px 5px",
			"color":       "#74787E",
			"font-size":   "15px",
			"line-height": "18px",
		},
		".data-wrapper caption": {
			"text-align":    "left",
			"font-weight":   "bold",
			"margin-bottom": "4px",
		},
		".invite-code": {
			"display":          "inline-block",
			"padding-top":      "20px",
			"padding-right":    "36px",
			"padding-bottom":   "16px",
			"padding-left":     "36px",
			"border-radius":    "3px",
			"font-family":      "Consolas, monaco, monospace",
			"font-size":        "28px",
			"text-align":       "center",
			"letter-spacing":   "8px",
			"color":            "#555",
			"background-color": "#eee",
		},
		".vml-button-wrapper": {
			"margin":        "30px auto",
			"v-text-anchor": "middle",
			"text-align":    "center",
		},
		".invite-code-container": {
			"margin-top":    "30px",
			"margin-bottom": "30px",
		},
		".invite-code-table": {
			"padding":    "0",
			"text-align": "center",
		},
		".invite-code-cell": {
			"display":          "inline-block",
			"border-radius":    "3px",
			"font-family":      "Consolas, monaco, monospace",
			"font-size":        "28px",
			"text-align":       "center",
			"letter-spacing":   "8px",
			"color":            "#555",
			"background-color": "#eee",
			"padding":          "20px",
		},
		".button": {
			"-webkit-text-size-adjust": "none",
			"display":                  "inline-block",
			"background-color":         "#3869D4",
			"border-radius":            "3px",
			"color":                    "#ffffff !important",
			"font-size":                "15px",
			"line-height":              "45px",
			"padding":                  "0 10px",
			"mso-hide":                 "all",
			"text-align":               "center",
			"text-decoration":          "none",
		},
	}

	// Apply theme-specific CSS customizations
	if h.Theme != nil && h.Theme.Name() == "flat" {
		// Flat theme customizations
		defaultCSS["body"]["background-color"] = "#2c3e50"
		defaultCSS[".email-wrapper"]["background-color"] = "#2c3e50"
		defaultCSS[".email-footer p"]["color"] = "#eaeaea"
		defaultCSS[".button"]["background-color"] = "#00948d"
		defaultCSS[".button"]["border-radius"] = "0"
	}

	// Handle body_width override
	if e.Body.TemplateOverrides != nil {
		if bodyWidth, ok := e.Body.TemplateOverrides["body_width"].(string); ok && bodyWidth != "" {
			// Update email-body_inner and email-footer widths
			if defaultCSS[".email-body_inner"] != nil {
				defaultCSS[".email-body_inner"]["width"] = bodyWidth
			}
			if defaultCSS[".email-footer"] != nil {
				defaultCSS[".email-footer"]["width"] = bodyWidth
			}
		}
	}

	// Merge user overrides if present
	if e.Body.TemplateOverrides != nil {
		if userCSS, ok := e.Body.TemplateOverrides["css"].(map[string]map[string]interface{}); ok {
			for sel, props := range userCSS {
				if defProps, exists := defaultCSS[sel]; exists {
					for k, v := range props {
						defProps[k] = v
					}
					defaultCSS[sel] = defProps
				} else {
					defaultCSS[sel] = props
				}
			}
			e.Body.TemplateOverrides["css"] = defaultCSS
		} else {
			e.Body.TemplateOverrides["css"] = defaultCSS
		}
	} else {
		e.Body.TemplateOverrides = map[string]any{"css": defaultCSS}
	}

	return nil
}

// default values of the engine
func setDefaultHermesValues(h *Hermes) error {
	defaultTextDirection := TDLeftToRight
	defaultHermes := Hermes{
		Theme:         new(Default),
		TextDirection: defaultTextDirection,
		Product: Product{
			Name:        "Hermes",
			Copyright:   "Copyright © 2025 Hermes. All rights reserved.",
			TroubleText: "If you’re having trouble with the button '{ACTION}', copy and paste the URL below into your web browser.",
		},
	}
	// Merge the given hermes engine configuration with default one
	// Default one overrides all zero values
	err := mergo.Merge(h, defaultHermes)
	if err != nil {
		return err
	}
	if h.TextDirection != TDLeftToRight && h.TextDirection != TDRightToLeft {
		h.TextDirection = defaultTextDirection
	}

	return nil
}

// GenerateHTML generates the email body from data to an HTML Reader
// This is for modern email clients
func (h *Hermes) GenerateHTML(email Email) (string, error) {
	err := setDefaultHermesValues(h)
	if err != nil {
		return "", err
	}

	t, err := getHTMLTemplate(h.Theme)
	if err != nil {
		return "", err
	}
	return h.generateTemplate(email, t)
}

// GeneratePlainText generates the email body from data
// This is for old email clients
func (h *Hermes) GeneratePlainText(email Email) (string, error) {
	err := setDefaultHermesValues(h)
	if err != nil {
		return "", err
	}

	t, err := getPlainTextTemplate(h.Theme)
	if err != nil {
		return "", err
	}
	template, err := h.generateTemplate(email, t)
	if err != nil {
		return "", err
	}

	return html2text.FromString(template, html2text.Options{PrettyTables: true})
}

func (h *Hermes) generateTemplate(email Email, t *template.Template) (string, error) {
	err := setDefaultEmailValues(h, &email)
	if err != nil {
		return "", err
	}

	if len(email.Body.Table.Data) > 0 {
		logrus.Warn("Email.Body.Table field is deprecated, please use Email.Body.Tables instead")
		email.Body.Tables = append(email.Body.Tables, email.Body.Table)
	}

	var b bytes.Buffer
	err = t.Execute(&b, Template{*h, email})
	if err != nil {
		return "", err
	}

	res := b.String()
	if h.DisableCSSInlining {
		return res, nil
	}

	// Inlining CSS
	prem, err := premailer.NewPremailerFromString(res, premailer.NewOptions())
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

// TemplateBase returns a base template from which to parse others in
// order to provide functionality that is added by this package. It is
// the base from which raw template sources provided by a theme are
// parsed.
func TemplateBase() *template.Template {
	return template.New("hermes").Funcs(sprig.FuncMap()).Funcs(templateFuncs).Funcs(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) }, // Used for keeping comments in generated template
	})
}

func getHTMLTemplate(t Theme) (*template.Template, error) {
	if t, ok := t.(ParsedHTMLTheme); ok {
		return t.ParsedHTMLTemplate()
	}
	return TemplateBase().Parse(t.HTMLTemplate())
}

func getPlainTextTemplate(t Theme) (*template.Template, error) {
	if t, ok := t.(ParsedPlainTextTheme); ok {
		return t.ParsedPlainTextTemplate()
	}
	return TemplateBase().Parse(t.PlainTextTemplate())
}
