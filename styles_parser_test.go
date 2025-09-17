package hermes

import (
	"strings"
	"testing"
)

func TestParseStylesDefinitionBasic(t *testing.T) {
	css := `/* comment */
body { color: #111; background-color: #fff; }
.a, .b { font-size: 14px; }
.a { line-height: 20px; }
`
	styles := ParseStylesDefinition(css)
	if bodyProps := styles["body"]; bodyProps == nil {
		t.Fatalf("expected body selector to be parsed, got nil")
	} else if bodyProps["color"] != "#111" || bodyProps["background-color"] != "#fff" {
		t.Fatalf("expected body styles parsed; got %#v", bodyProps)
	}
	if styles[".a"]["font-size"] != "14px" || styles[".b"]["font-size"] != "14px" {
		t.Fatalf("expected shared font-size for .a and .b")
	}
	if styles[".a"]["line-height"] != "20px" {
		t.Fatalf("expected later rule to merge additional property for .a")
	}
}

func TestParseStylesDefinitionIgnoresInvalid(t *testing.T) {
	css := `div { color: red; }
span { missing colon }
/* comment */
`
	styles := ParseStylesDefinition(css)
	if styles["div"]["color"] != "red" {
		t.Fatalf("expected div color red")
	}
	if _, ok := styles["span"]; ok {
		t.Fatalf("expected span rule to be ignored due to malformed property")
	}
}

func TestParseStylesDefinitionPreservesSelectorComments(t *testing.T) {
	css := `@font-face /* v1 */ { font-family: MyFont; src: url("v1.woff2"); }
@font-face /* v2 */ { font-family: MyFont; src: url("v2.woff2"); }
`
	styles := ParseStylesDefinition(css)
	// Expect two separate keys including their trailing comment tokens
	count := 0
	for sel := range styles {
		if strings.HasPrefix(sel, "@font-face") {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 @font-face entries, got %d (%#v)", count, styles)
	}
}

func TestParseStylesDefinitionRemovesStandaloneComments(t *testing.T) {
	css := `/* This is a standalone comment */
body { color: black; }
/* Another standalone comment */
.header { background: white; }
@font-face /* inline comment */ { font-family: Test; }
`
	styles := ParseStylesDefinition(css)

	// Should have clean selectors without standalone comments
	if _, ok := styles["body"]; !ok {
		t.Error("expected body selector to be present")
	}
	if _, ok := styles[".header"]; !ok {
		t.Error("expected .header selector to be present")
	}

	// Should preserve inline comment with @font-face
	found := false
	for key := range styles {
		if strings.Contains(key, "@font-face") && strings.Contains(key, "inline comment") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected @font-face with inline comment to be preserved")
	}
}

func TestParseStylesDefinitionEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		css  string
		want map[string]map[string]any
	}{
		{
			name: "empty css",
			css:  "",
			want: map[string]map[string]any{},
		},
		{
			name: "only comments",
			css:  "/* just a comment */",
			want: map[string]map[string]any{},
		},
		{
			name: "selector with no properties",
			css:  "div { }",
			want: map[string]map[string]any{},
		},
		{
			name: "property with !important",
			css:  "p { color: red !important; }",
			want: map[string]map[string]any{
				"p": {"color": "red !important"},
			},
		},
		{
			name: "multiple selectors",
			css:  "h1, h2, h3 { margin: 0; }",
			want: map[string]map[string]any{
				"h1": {"margin": "0"},
				"h2": {"margin": "0"},
				"h3": {"margin": "0"},
			},
		},
		{
			name: "property values with quotes",
			css:  `.btn { content: "Click me"; font-family: "Arial", sans-serif; }`,
			want: map[string]map[string]any{
				".btn": {
					"content":     `"Click me"`,
					"font-family": `"Arial", sans-serif`,
				},
			},
		},
		{
			name: "complex selector with comments",
			css:  `@media /* print */ screen { color: black; }`,
			want: map[string]map[string]any{
				"@media /* print */ screen": {"color": "black"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseStylesDefinition(tt.css)
			if len(got) != len(tt.want) {
				t.Errorf("expected %d selectors, got %d", len(tt.want), len(got))
			}
			for sel, props := range tt.want {
				gotProps, ok := got[sel]
				if !ok {
					t.Errorf("missing selector %q", sel)
					continue
				}
				for prop, val := range props {
					if gotProps[prop] != val {
						t.Errorf("selector %q property %q: want %v, got %v", sel, prop, val, gotProps[prop])
					}
				}
			}
		})
	}
}

func TestParseStylesDefinitionNotificationExample(t *testing.T) {
	// Real-world notification CSS with multiple @font-face blocks
	css := `/* Notification Template CSS Overrides */
@font-face /* Clario Air */ {
  font-display: swap !important;
  font-family: Clario !important;
  font-style: normal !important;
  font-weight: 100 !important;
  src: local('Clario Air'), local('ClarioAir') !important;
}

@font-face /* Clario Bold */ {
  font-display: swap !important;
  font-family: Clario !important;
  font-style: normal !important;
  font-weight: 700 !important;
  src: local('Clario Bold'), local('ClarioBold') !important;
}

* {
  font-family: Clario, "Helvetica Neue", Helvetica, sans-serif !important;
  -webkit-box-sizing: border-box !important;
  box-sizing: border-box !important;
}

#mail-footer-link { color: blue; }
body { font-family: Clario, "Helvetica Neue", Arial, sans-serif; }
.email-body_inner { width: 900px; }`

	styles := ParseStylesDefinition(css)

	// Check distinct @font-face blocks preserved
	faceCount := 0
	for k := range styles {
		if strings.HasPrefix(k, "@font-face") {
			faceCount++
		}
	}
	if faceCount != 2 {
		t.Errorf("expected 2 @font-face entries, got %d", faceCount)
	}

	// Check universal selector
	if _, ok := styles["*"]; !ok {
		t.Error("expected universal selector '*' to be preserved")
	}

	// Check ID selector
	if styles["#mail-footer-link"]["color"] != "blue" {
		t.Error("expected #mail-footer-link color to be blue")
	}

	// Check body font-family
	if !strings.Contains(styles["body"]["font-family"].(string), "Clario") {
		t.Error("expected body font-family to contain Clario")
	}
}
