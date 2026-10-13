package srt

import (
	"strings"
	"testing"
)

const (
	lf  = "\n"
	cr  = "\r"
	tab = "\t"

	timingA  = `00:00:01,000 --> 00:00:02,000`
	timingB  = `00:00:03,000 --> 00:00:04,000`
	longLine = `Out of order, and this line is far too long to read comfortably on any screen!`
)

// unix ends every line with LF.
func unix(lines ...string) string {
	return strings.Join(lines, lf) + lf
}

// dos ends every line with CRLF.
func dos(lines ...string) string {
	return strings.Join(lines, cr+lf) + cr + lf
}

// movie is the movie.srt example: its cues start on lines 1, 5, 9 and 12.
var movie = unix(
	`1`, `00:00:01,000 --> 00:00:03,500`, `Hello there.`, ``,
	`2`, `00:00:03,000 --> 00:00:05,000`, `<i>Overlap with the first cue.</i>`, ``,
	`4`, `00:00:06,000 --> 00:00:06,000`, ``,
	`5`, `00:00:04,000 --> 00:00:04,500`, longLine, `Short.`,
)

func compareCues(t *testing.T, got, want []Cue) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf(`got %d cues, want %d: %+v`, len(got), len(want), got)
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.Number != w.Number || g.Start != w.Start || g.End != w.End {
			t.Errorf(`cue %d: number %d from %d to %d, want number %d from %d to %d`, i, g.Number, g.Start, g.End, w.Number, w.Start, w.End)
		}
		if g.Line != w.Line || g.TextLine != w.TextLine {
			t.Errorf(`cue %d: Line %d, TextLine %d, want Line %d, TextLine %d`, i, g.Line, g.TextLine, w.Line, w.TextLine)
		}
		if len(g.Text) != len(w.Text) {
			t.Errorf(`cue %d: text %q, want %q`, i, g.Text, w.Text)
			continue
		}
		for j := range w.Text {
			if g.Text[j] != w.Text[j] {
				t.Errorf(`cue %d: text line %d is %q, want %q`, i, j+1, g.Text[j], w.Text[j])
			}
		}
	}
}

func TestParse(t *testing.T) {
	hi := []Cue{{Number: 1, Start: 1000, End: 2000, Text: []string{`hi`}, Line: 1, TextLine: 3}}
	noText := []Cue{{Number: 1, Start: 1000, End: 2000, Line: 1}}
	tests := []struct {
		name string
		data string
		want []Cue
	}{
		{`BOM and CRLF`, string(bom) + dos(`1`, `00:00:01.000 --> 00:00:02.000  `, `hi`), hi},
		{`LF without BOM`, unix(`1`, `00:00:01.000 --> 00:00:02.000  `, `hi`), hi},
		{`BOM and LF`, string(bom) + unix(`1`, timingA, `hi`), hi},
		{`missing final LF`, `1` + lf + timingA + lf + `hi`, hi},
		{`CRLF and missing final CRLF`, `1` + cr + lf + timingA + cr + lf + `hi`, hi},
		{`final CR without LF stays`, `1` + cr + lf + timingA + cr + lf + `hi` + cr, []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`hi` + cr}, Line: 1, TextLine: 3},
		}},
		{`extra blank lines`, unix(``, ``, `1`, timingA, `a`, ``, ``, ``, `2`, timingB, `b`, ``, ``), []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`a`}, Line: 3, TextLine: 5},
			{Number: 2, Start: 3000, End: 4000, Text: []string{`b`}, Line: 9, TextLine: 11},
		}},
		{`whitespace-only separator lines`, unix(`1`, timingA, `a`, ` `+tab+` `, tab, `2`, timingB, `b`, `  `), []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`a`}, Line: 1, TextLine: 3},
			{Number: 2, Start: 3000, End: 4000, Text: []string{`b`}, Line: 6, TextLine: 8},
		}},
		{`blank lines with CRLF`, dos(``, `1`, timingA, `a`, ``, `2`, timingB, `b`), []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`a`}, Line: 2, TextLine: 4},
			{Number: 2, Start: 3000, End: 4000, Text: []string{`b`}, Line: 6, TextLine: 8},
		}},
		{`CR inside a text line`, unix(`1`, timingA, `a`+cr+`b`), []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`a` + cr + `b`}, Line: 1, TextLine: 3},
		}},
		{`only one CR is removed`, dos(`1`, timingA, `a`+cr), []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`a` + cr}, Line: 1, TextLine: 3},
		}},
		{`cue without text before another`, unix(`1`, timingA, ``, `2`, timingB, `b`), []Cue{
			{Number: 1, Start: 1000, End: 2000, Line: 1, TextLine: 0},
			{Number: 2, Start: 3000, End: 4000, Text: []string{`b`}, Line: 4, TextLine: 6},
		}},
		{`cue without text at the end`, unix(`1`, timingA), noText},
		{`cue without text and no final LF`, `1` + lf + timingA, noText},
		{`text like a number and a timing line`, unix(`1`, timingA, `42`, timingB), []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`42`, timingB}, Line: 1, TextLine: 3},
		}},
		{`spaces kept in text lines`, unix(`1`, timingA, `  lead`, `trail  `, tab+`both `+tab), []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`  lead`, `trail  `, tab + `both ` + tab}, Line: 1, TextLine: 3},
		}},
		{`empty input`, ``, nil},
		{`BOM only`, string(bom), nil},
		{`blank lines only`, unix(``, ` `, tab), nil},
		{`CRLF blank lines only`, dos(``, ``), nil},
		{`movie.srt`, movie, []Cue{
			{Number: 1, Start: 1000, End: 3500, Text: []string{`Hello there.`}, Line: 1, TextLine: 3},
			{Number: 2, Start: 3000, End: 5000, Text: []string{`<i>Overlap with the first cue.</i>`}, Line: 5, TextLine: 7},
			{Number: 4, Start: 6000, End: 6000, Line: 9, TextLine: 0},
			{Number: 5, Start: 4000, End: 4500, Text: []string{longLine, `Short.`}, Line: 12, TextLine: 14},
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(`f.srt`, []byte(tc.data))
			if err != nil {
				t.Fatalf(`unexpected error: %v`, err)
			}
			compareCues(t, got, tc.want)
		})
	}
}

func TestParseCueNumber(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{`7`, 7},
		{`0042`, 42},
		{` ` + tab + `7` + tab + ` `, 7},
		{`123456789`, 123456789},
		{`000000001`, 1},
		{`0`, 0},
	}
	for _, tc := range tests {
		t.Run(tc.line, func(t *testing.T) {
			got, err := Parse(`f.srt`, []byte(unix(tc.line, timingA, `x`)))
			if err != nil {
				t.Fatalf(`unexpected error: %v`, err)
			}
			compareCues(t, got, []Cue{
				{Number: tc.want, Start: 1000, End: 2000, Text: []string{`x`}, Line: 1, TextLine: 3},
			})
		})
	}
}

func TestParseCueNumberInvalid(t *testing.T) {
	lines := []string{`x`, `1234567890`, `+7`, `1a`, `-1`, `1 2`, ` x `, tab + `7a`}
	for _, line := range lines {
		want := `srtvet: a.srt:1: expected a cue number, found "` + line + `"`
		for _, data := range []string{unix(line, timingA, `x`), dos(line, timingA, `x`), line} {
			cues, err := Parse(`a.srt`, []byte(data))
			if err == nil {
				t.Errorf(`Parse(%q): no error, got %d cues`, data, len(cues))
				continue
			}
			if err.Error() != want {
				t.Errorf(`Parse(%q): error %q, want %q`, data, err.Error(), want)
			}
		}
	}
}

func TestParseErrors(t *testing.T) {
	const noTiming = `srtvet: bad.srt:1: cue number without a timing line`
	tests := []struct {
		name string
		file string
		data string
		want string
	}{
		{`not a number`, `bad.srt`, `x`, `srtvet: bad.srt:1: expected a cue number, found "x"`},
		{`number at the end`, `bad.srt`, `1`, noTiming},
		{`number and LF at the end`, `bad.srt`, unix(`1`), noTiming},
		{`number and CRLF at the end`, `bad.srt`, dos(`1`), noTiming},
		{`short arrow`, `bad.srt`, `1` + lf + `00:00:01,000 -> 00:00:02,000`, `srtvet: bad.srt:2: expected a timing line, found "00:00:01,000 -> 00:00:02,000"`},
		{`second cue without a timing line`, `bad.srt`, unix(`1`, timingA, `hi`, ``, ``, `2`, `bad`), `srtvet: bad.srt:7: expected a timing line, found "bad"`},
		{`blank line after the number`, `bad.srt`, unix(`1`, ``, timingA), `srtvet: bad.srt:2: expected a timing line, found ""`},
		{`spaces after the number`, `bad.srt`, unix(`1`, `  `), `srtvet: bad.srt:2: expected a timing line, found "  "`},
		{`timing line with a leading space`, `bad.srt`, unix(`1`, ` `+timingA), `srtvet: bad.srt:2: expected a timing line, found " ` + timingA + `"`},
		{`timing error with CRLF`, `bad.srt`, dos(`1`, `bad`), `srtvet: bad.srt:2: expected a timing line, found "bad"`},
		{`last number without a timing line`, `bad.srt`, unix(`1`, timingA, `hi`, ``, `2`), `srtvet: bad.srt:5: cue number without a timing line`},
		{`second number invalid`, `bad.srt`, unix(`1`, timingA, `hi`, ``, `two`), `srtvet: bad.srt:5: expected a cue number, found "two"`},
		{`first error only`, `bad.srt`, unix(`x`, `1`, `bad`), `srtvet: bad.srt:1: expected a cue number, found "x"`},
		{`timing error before a later number error`, `bad.srt`, unix(`1`, `bad`, ``, `y`), `srtvet: bad.srt:2: expected a timing line, found "bad"`},
		{`BOM is not part of the line`, `bad.srt`, string(bom) + `x`, `srtvet: bad.srt:1: expected a cue number, found "x"`},
		{`blank lines are counted`, `bad.srt`, unix(``, ``, `x`), `srtvet: bad.srt:3: expected a cue number, found "x"`},
		{`name as given`, `dir/sub.srt`, `x`, `srtvet: dir/sub.srt:1: expected a cue number, found "x"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues, err := Parse(tc.file, []byte(tc.data))
			if err == nil {
				t.Fatalf(`no error, got %d cues`, len(cues))
			}
			if err.Error() != tc.want {
				t.Errorf(`error = %q, want %q`, err.Error(), tc.want)
			}
		})
	}
}
