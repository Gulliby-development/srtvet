package ops

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

// cue builds a cue for the tests of this package.
func cue(number int, start, end int64, line, textLine int, text ...string) srt.Cue {
	return srt.Cue{Number: number, Start: start, End: end, Text: text, Line: line, TextLine: textLine}
}

func assertFindings(t *testing.T, got, want []Finding) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d findings %+v, want %d findings %+v", len(got), got, len(want), want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("finding %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}

const (
	msgSV002 = "end is not after start"
	msgSV005 = "no text"
)

func TestCheckNumbering(t *testing.T) {
	tests := []struct {
		name string
		cues []srt.Cue
		want []Finding
	}{
		{"1 2 5 6 reports only the 5", []srt.Cue{
			cue(1, 1000, 2000, 1, 3, "a"),
			cue(2, 3000, 4000, 5, 7, "b"),
			cue(5, 5000, 6000, 9, 11, "c"),
			cue(6, 7000, 8000, 13, 15, "d"),
		}, []Finding{{9, "SV001", "cue number 5, expected 3"}}},
		{"first cue numbered 2", []srt.Cue{
			cue(2, 1000, 2000, 1, 3, "a"),
		}, []Finding{{1, "SV001", "cue number 2, expected 1"}}},
		{"first cue numbered 0", []srt.Cue{
			cue(0, 1000, 2000, 4, 6, "a"),
		}, []Finding{{4, "SV001", "cue number 0, expected 1"}}},
		{"repeated number", []srt.Cue{
			cue(1, 1000, 2000, 1, 3, "a"),
			cue(1, 3000, 4000, 5, 7, "b"),
		}, []Finding{{5, "SV001", "cue number 1, expected 2"}}},
		{"1 2 3 is clean", []srt.Cue{
			cue(1, 1000, 2000, 1, 3, "a"),
			cue(2, 3000, 4000, 5, 7, "b"),
			cue(3, 5000, 6000, 9, 11, "c"),
		}, nil},
		{"no cues", nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, Check(tc.cues, 42), tc.want)
		})
	}
}

func TestCheckEndAndText(t *testing.T) {
	tests := []struct {
		name string
		cues []srt.Cue
		want []Finding
	}{
		{"end equal to start", []srt.Cue{cue(1, 1000, 1000, 1, 3, "a")},
			[]Finding{{1, "SV002", msgSV002}}},
		{"end less than start", []srt.Cue{cue(1, 1000, 999, 1, 3, "a")},
			[]Finding{{1, "SV002", msgSV002}}},
		{"end one millisecond after start", []srt.Cue{cue(1, 1000, 1001, 1, 3, "a")}, nil},
		{"zero start and end", []srt.Cue{cue(1, 0, 0, 1, 3, "a")},
			[]Finding{{1, "SV002", msgSV002}}},
		{"no text", []srt.Cue{cue(1, 1000, 2000, 1, 0)},
			[]Finding{{1, "SV005", msgSV005}}},
		{"empty text slice", []srt.Cue{{Number: 1, Start: 1000, End: 2000, Text: []string{}, Line: 1}},
			[]Finding{{1, "SV005", msgSV005}}},
		{"with text", []srt.Cue{cue(1, 1000, 2000, 1, 3, "a")}, nil},
		{"bad end and no text in code order", []srt.Cue{cue(1, 1000, 1000, 1, 0)},
			[]Finding{{1, "SV002", msgSV002}, {1, "SV005", msgSV005}}},
		{"only the second cue has no text", []srt.Cue{
			cue(1, 1000, 2000, 1, 3, "a"),
			cue(2, 3000, 4000, 5, 0),
			cue(3, 5000, 6000, 8, 10, "c"),
		}, []Finding{{5, "SV005", msgSV005}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, Check(tc.cues, 42), tc.want)
		})
	}
}

func TestCheckOverlap(t *testing.T) {
	tests := []struct {
		name  string
		start int64
		want  []Finding
	}{
		{"inside the previous cue", 3000,
			[]Finding{{5, "SV003", "starts at 00:00:03,000, before the previous cue ends at 00:00:03,500"}}},
		{"equal starts", 1000,
			[]Finding{{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:03,500"}}},
		{"one millisecond before the previous end", 3499,
			[]Finding{{5, "SV003", "starts at 00:00:03,499, before the previous cue ends at 00:00:03,500"}}},
		{"at the previous end", 3500, nil},
		{"after the previous end", 3501, nil},
		{"one millisecond before the previous start", 999,
			[]Finding{{5, "SV004", "starts at 00:00:00,999, before the previous cue starts at 00:00:01,000"}}},
		{"at zero", 0,
			[]Finding{{5, "SV004", "starts at 00:00:00,000, before the previous cue starts at 00:00:01,000"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := []srt.Cue{
				cue(1, 1000, 3500, 1, 3, "a"),
				cue(2, tc.start, tc.start+500, 5, 7, "b"),
			}
			got := Check(cues, 42)
			assertFindings(t, got, tc.want)
			if len(got) > 1 {
				t.Errorf("a cue got more than one of SV003 and SV004: %+v", got)
			}
		})
	}
}

func TestCheckFirstCueHasNoPrevious(t *testing.T) {
	for _, start := range []int64{0, 1000, 359999000} {
		got := Check([]srt.Cue{cue(1, start, start+500, 1, 3, "a")}, 42)
		assertFindings(t, got, nil)
	}
}

func TestCheckLineLength(t *testing.T) {
	const ten = "h\u00e9llo w\u00f6rl"
	const eleven = "h\u00e9llo w\u00f6rld"
	if utf8.RuneCountInString(ten) != 10 || len(ten) <= 10 {
		t.Fatalf("test line %q must have 10 code points and more bytes", ten)
	}
	if utf8.RuneCountInString(eleven) != 11 || len(eleven) <= 11 {
		t.Fatalf("test line %q must have 11 code points and more bytes", eleven)
	}
	tests := []struct {
		name    string
		maxLine int
		text    []string
		want    []Finding
	}{
		{"exactly the limit", 10, []string{ten}, nil},
		{"one above the limit", 10, []string{eleven},
			[]Finding{{3, "SV006", "line 1 has 11 characters, the limit is 10"}}},
		{"second line too long", 10, []string{ten, eleven},
			[]Finding{{4, "SV006", "line 2 has 11 characters, the limit is 10"}}},
		{"several long lines in text-line order", 10, []string{eleven, "ok", eleven + "!!"},
			[]Finding{
				{3, "SV006", "line 1 has 11 characters, the limit is 10"},
				{5, "SV006", "line 3 has 13 characters, the limit is 10"},
			}},
		{"limit 0 is off", 0, []string{eleven, strings.Repeat("\u00e9", 500)}, nil},
		{"limit 1 with one multi-byte character", 1, []string{"\u00e9"}, nil},
		{"limit 1 with two characters", 1, []string{"ab"},
			[]Finding{{3, "SV006", "line 1 has 2 characters, the limit is 1"}}},
		{"default 42 at the limit", 42, []string{strings.Repeat("\u00e9", 42)}, nil},
		{"default 42 one above", 42, []string{strings.Repeat("\u00e9", 43)},
			[]Finding{{3, "SV006", "line 1 has 43 characters, the limit is 42"}}},
		{"tags count like text", 10, []string{"<i>12345</i>"},
			[]Finding{{3, "SV006", "line 1 has 12 characters, the limit is 10"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := []srt.Cue{cue(1, 1000, 2000, 1, 3, tc.text...)}
			assertFindings(t, Check(cues, tc.maxLine), tc.want)
		})
	}
}

func TestCheckMovie(t *testing.T) {
	movie := []srt.Cue{
		cue(1, 1000, 3500, 1, 3, "Hello there."),
		cue(2, 3000, 5000, 5, 7, "<i>Overlap with the first cue.</i>"),
		cue(4, 6000, 6000, 9, 0),
		cue(5, 4000, 4500, 12, 14,
			"Out of order, and this line is far too long to read comfortably on any screen!",
			"Short."),
	}
	five := []Finding{
		{5, "SV003", "starts at 00:00:03,000, before the previous cue ends at 00:00:03,500"},
		{9, "SV001", "cue number 4, expected 3"},
		{9, "SV002", msgSV002},
		{9, "SV005", msgSV005},
		{12, "SV004", "starts at 00:00:04,000, before the previous cue starts at 00:00:06,000"},
	}
	six := append(append([]Finding{}, five...),
		Finding{14, "SV006", "line 1 has 78 characters, the limit is 42"})
	tests := []struct {
		name    string
		maxLine int
		want    []Finding
	}{
		{"default limit", 42, six},
		{"limit 0", 0, five},
		{"limit 80", 80, five},
		{"limit 78", 78, five},
		{"limit 77", 77, append(append([]Finding{}, five...),
			Finding{14, "SV006", "line 1 has 78 characters, the limit is 77"})},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, Check(movie, tc.maxLine), tc.want)
		})
	}
}
