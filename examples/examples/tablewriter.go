package examples

import (
	"github.com/olekukonko/tablewriter/tw"
)

// this is temporary until https://github.com/jaytaylor/html2text/pull/68 is merged
var alignment = tw.AlignLeft

func NoOp() string {
	return string(alignment)
}
