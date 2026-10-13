// Package ops holds the operations on parsed cues: the check rules,
// retiming and reordering.
package ops

import (
	"fmt"
	"unicode/utf8"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

// Finding is one problem reported by Check.
type Finding struct {
	// Line is the line number the finding points at.
	Line int
	// Code is the rule code, SV001 to SV006.
	Code string
	// Message is the text after the code.
	Message string
}

// Check runs the six rules on the cues of one file. The findings are in
// cue order and, inside a cue, in code order SV001 to SV006. A maxLine
// of 0 turns the line-length rule off.
func Check(cues []srt.Cue, maxLine int) []Finding {
	var out []Finding
	for i, c := range cues {
		add := func(line int, code, message string) {
			out = append(out, Finding{Line: line, Code: code, Message: message})
		}
		want := 1
		if i > 0 {
			want = cues[i-1].Number + 1
		}
		if c.Number != want {
			add(c.Line, "SV001", fmt.Sprintf("cue number %d, expected %d", c.Number, want))
		}
		if c.End <= c.Start {
			add(c.Line, "SV002", "end is not after start")
		}
		if i > 0 {
			prev := cues[i-1]
			if c.Start < prev.Start {
				add(c.Line, "SV004", fmt.Sprintf("starts at %s, before the previous cue starts at %s",
					srt.FormatTimestamp(c.Start), srt.FormatTimestamp(prev.Start)))
			} else if c.Start < prev.End {
				add(c.Line, "SV003", fmt.Sprintf("starts at %s, before the previous cue ends at %s",
					srt.FormatTimestamp(c.Start), srt.FormatTimestamp(prev.End)))
			}
		}
		if len(c.Text) == 0 {
			add(c.Line, "SV005", "no text")
		}
		if maxLine == 0 {
			continue
		}
		for k, line := range c.Text {
			if n := utf8.RuneCountInString(line); n > maxLine {
				add(c.TextLine+k, "SV006", fmt.Sprintf("line %d has %d characters, the limit is %d", k+1, n, maxLine))
			}
		}
	}
	return out
}
