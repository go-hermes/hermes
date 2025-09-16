package examples

import (
	"github.com/olekukonko/tablewriter"
)

// this is temporary until https://github.com/jaytaylor/html2text/pull/68 is merged
var alignment = tablewriter.ALIGN_LEFT

func NoOp() int {
	return alignment
}
