package examples

import (
	"github.com/olekukonko/tablewriter/tw"
)

// alignment constant used by html2text for table formatting
var alignment = tw.AlignLeft

func NoOp() string {
	return string(alignment)
}
