package hermes

import (
	"fmt"
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
	if styles["body"]["color"] != "#111" || styles["body"]["background-color"] != "#fff" {
		t.Fatalf("expected body styles parsed; got %#v", styles["body"])
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
	fmt.Printf("Parsed styles: %#v\n", styles)
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
