package srt

import (
	"bytes"
	"errors"
	"testing"
)

// failWriter rejects every write with its error.
type failWriter struct {
	err error
}

func (f failWriter) Write([]byte) (int, error) {
	return 0, f.err
}

func TestWrite(t *testing.T) {
	tests := []struct {
		name string
		cues []Cue
		want string
	}{
		{`three cues, the middle one without text`, []Cue{
			{Number: 1, Start: 1000, End: 3500, Text: []string{`<i>Hello there.</i>`, `Second line.`}},
			{Number: 2, Start: 4000, End: 4500},
			{Number: 3, Start: 70000, End: 72040, Text: []string{`Third, but listed first.`}},
		}, unix(
			`1`, `00:00:01,000 --> 00:00:03,500`, `<i>Hello there.</i>`, `Second line.`, ``,
			`2`, `00:00:04,000 --> 00:00:04,500`, ``,
			`3`, `00:01:10,000 --> 00:01:12,040`, `Third, but listed first.`,
		)},
		{`no cues`, nil, ``},
		{`empty slice`, []Cue{}, ``},
		{`one cue`, []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`hi`}},
		}, unix(`1`, timingA, `hi`)},
		{`one cue without text, the time limits`, []Cue{
			{Number: 42, Start: 0, End: MaxTime},
		}, unix(`42`, `00:00:00,000 --> 99:59:59,999`)},
		{`last cue without text`, []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`a`}},
			{Number: 2, Start: 3000, End: 4000},
		}, unix(`1`, timingA, `a`, ``, `2`, timingB)},
		{`text lines exactly as stored`, []Cue{
			{Number: 1, Start: 1000, End: 2000, Text: []string{`  lead`, `trail  `, `42`, timingB}},
		}, unix(`1`, timingA, `  lead`, `trail  `, `42`, timingB)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := Write(&buf, tc.cues); err != nil {
				t.Fatalf(`unexpected error: %v`, err)
			}
			if got := buf.String(); got != tc.want {
				t.Errorf(`output is %q, want %q`, got, tc.want)
			}
			if bytes.Contains(buf.Bytes(), []byte(cr)) {
				t.Errorf(`output holds a CR: %q`, buf.String())
			}
			if bytes.HasPrefix(buf.Bytes(), bom) {
				t.Errorf(`output starts with a BOM: %q`, buf.String())
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	want := errors.New(`disk full`)
	err := Write(failWriter{err: want}, []Cue{{Number: 1, Start: 1000, End: 2000, Text: []string{`hi`}}})
	if !errors.Is(err, want) {
		t.Fatalf(`Write returned %v, want %v`, err, want)
	}
}
