// Package ops holds the operations on parsed cues: the check rules,
// retiming and reordering.
package ops

import (
	"fmt"
	"slices"
	"unicode/utf8"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

// Finding is one problem reported by Check or CheckAll.
type Finding struct {
	// Line is the line number the finding points at.
	Line int
	// Code is the rule code, SV001 to SV008.
	Code string
	// Message is the text after the code.
	Message string
}

// Check runs SV001 to SV006 and SV008 on the cues of one file. The findings
// are in cue order and, inside a cue, in code order. SV008 comes after the
// other findings of its cue. A maxLine of 0 turns the line-length rule off.
func Check(cues []srt.Cue, maxLine int) []Finding {
	return CheckAll(cues, maxLine, 0)
}

// CheckAll runs the Check rules and, when maxCPS is positive, the
// reading-speed rule SV007. Findings are in cue order and, inside a cue,
// SV007 follows SV006 and SV008 comes last. A maxLine of 0 turns the
// line-length rule off.
func CheckAll(cues []srt.Cue, maxLine, maxCPS int) []Finding {
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
		if maxLine > 0 {
			for k, line := range c.Text {
				if n := utf8.RuneCountInString(line); n > maxLine {
					add(c.TextLine+k, "SV006", fmt.Sprintf("line %d has %d characters, the limit is %d", k+1, n, maxLine))
				}
			}
		}
		if maxCPS > 0 && len(c.Text) > 0 && c.End > c.Start {
			duration := c.End - c.Start
			var characters int64
			for _, line := range c.Text {
				characters += int64(utf8.RuneCountInString(line))
			}
			if characters*1000 > int64(maxCPS)*duration {
				tenths := (characters*10000 + duration/2) / duration
				add(c.Line, "SV007", fmt.Sprintf("reading speed %d.%d characters per second, the limit is %d", tenths/10, tenths%10, maxCPS))
			}
		}
		if i > 0 && len(c.Text) > 0 {
			prev := cues[i-1]
			if c.Start == prev.Start && c.End == prev.End && slices.Equal(c.Text, prev.Text) {
				add(c.Line, "SV008", "same start, end and text as the previous cue")
			}
		}
	}
	return out
}
