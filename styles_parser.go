package hermes

import (
	"regexp"
	"strings"
)

// ParseStylesDefinition parses a raw CSS string into a StylesDefinition.
// It supports a very small subset of CSS adequate for simple selector { prop: value; } rules.
// - Ignores comments (/* ... */) and empty lines
// - Does not support nested rules, media queries, or at-rules (they should be injected separately)
// - Multiple selectors separated by commas are split and each receives the full property set
// Consumers can use this to transform their custom CSS overrides into a StylesDefinition for merging.
func ParseStylesDefinition(css string) StylesDefinition {
	styles := StylesDefinition{}
	// Keep selector comments intact; they may disambiguate duplicate selectors for consumer maps.
	// We'll only remove comments inside declaration bodies to avoid treating them as properties.
	blockRE := regexp.MustCompile(`(?s)([^{}]+)\{([^{}]+)\}`)
	matches := blockRE.FindAllStringSubmatch(css, -1)
	for _, m := range matches {
		selectorPart := strings.TrimSpace(m[1])
		propsPart := strings.TrimSpace(m[2])
		if selectorPart == "" || propsPart == "" {
			continue
		}
		selectors := strings.Split(selectorPart, ",")
		// Strip comments from declaration body only
		commentRE := regexp.MustCompile(`(?s)/\*.*?\*/`)
		propsPartNoComments := commentRE.ReplaceAllString(propsPart, "")
		decls := strings.Split(propsPartNoComments, ";")
		props := map[string]any{}
		for _, d := range decls {
			d = strings.TrimSpace(d)
			if d == "" {
				continue
			}
			if colon := strings.Index(d, ":"); colon != -1 {
				key := strings.TrimSpace(d[:colon])
				val := strings.TrimSpace(d[colon+1:])
				if key != "" && val != "" {
					props[key] = val
				}
			}
		}
		if len(props) == 0 {
			continue
		}
		for _, sel := range selectors {
			s := strings.TrimSpace(sel)
			if s == "" {
				continue
			}
			if existing, ok := styles[s]; ok {
				for k, v := range props {
					existing[k] = v
				}
			} else {
				cp := map[string]any{}
				for k, v := range props {
					cp[k] = v
				}
				styles[s] = cp
			}
		}
	}
	return styles
}
