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

type StylesDefinition map[string]map[string]any

// Theme is an interface to implement when creating a new theme
type Theme interface {
	Name() string              // The name of the theme
	Styles() StylesDefinition  // The CSS styles definition for the theme
	HTMLTemplate() string      // The golang template for HTML emails
	PlainTextTemplate() string // The golang templte for plain text emails (can be basic HTML)
}

func (s StylesDefinition) MergeCSSWithTheme(theme Theme) StylesDefinition {
	themeStyles := theme.Styles()
	for sel, props := range s {
		if defProps, exists := themeStyles[sel]; exists {
			for k, v := range props {
				defProps[k] = v
			}
			themeStyles[sel] = defProps
		} else {
			themeStyles[sel] = props
		}
	}
	return themeStyles
}

// normalizeStyles attempts to coerce various accepted override map types into StylesDefinition.
// Supports:
//   - StylesDefinition (already correct)
//   - map[string]map[string]interface{} (legacy test usage)
//
// Returns nil if the input type is unsupported.
func normalizeStyles(v any) StylesDefinition {
	switch css := v.(type) {
	case StylesDefinition:
		return css
	case map[string]map[string]interface{}:
		out := StylesDefinition{}
		for sel, props := range css {
			cp := map[string]any{}
			for k, val := range props {
				cp[k] = val
			}
			out[sel] = cp
		}
		return out
	default:
		return nil
	}
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
	Title        string        // Title of the table
	Data         [][]Entry     // Contains data
	Columns      Columns       // Contains meta-data for display purpose (width, alignement)
	Class        string        // Optional CSS class applied to the wrapping table element
	TitleUnsafe  template.HTML // Optional unsafe HTML that replaces Title when set
	Footer       string        // Optional footer text rendered below the table
	FooterUnsafe template.HTML // Optional unsafe HTML footer rendered below the table (overrides Footer when set)
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

	styles := h.Theme.Styles()

	// Handle body_width override
	if e.Body.TemplateOverrides != nil {
		if bodyWidth, ok := e.Body.TemplateOverrides["body_width"].(string); ok && bodyWidth != "" {
			// Update email-body_inner and email-footer widths
			if styles[".email-body_inner"] != nil {
				styles[".email-body_inner"]["width"] = bodyWidth
			}
			if styles[".email-footer"] != nil {
				styles[".email-footer"]["width"] = bodyWidth
			}
		}
	}

	// Merge user overrides if present using helper normalization
	if e.Body.TemplateOverrides != nil {
		if raw, ok := e.Body.TemplateOverrides["css"]; ok {
			if userStyles := normalizeStyles(raw); userStyles != nil {
				styles = userStyles.MergeCSSWithTheme(h.Theme)
			}
		}
		// ensure css key reflects merged styles regardless of override presence
		e.Body.TemplateOverrides["css"] = styles
	} else {
		e.Body.TemplateOverrides = map[string]any{"css": styles}
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
