package ops

import (
	"strings"
	"testing"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

func TestCheckAllDisabledAndCheckCompatibility(t *testing.T) {
	invalid := []srt.Cue{cue(2, 1000, 1000, 7, 0)}
	wantInvalid := []Finding{
		{7, "SV001", "cue number 2, expected 1"},
		{7, "SV002", msgSV002},
		{7, "SV005", msgSV005},
	}
	assertFindings(t, Check(invalid, 42), wantInvalid)
	assertFindings(t, CheckAll(invalid, 42, 0), wantInvalid)

	fast := []srt.Cue{cue(1, 1000, 1100, 1, 3, "xx")}
	assertFindings(t, Check(fast, 0), nil)
	assertFindings(t, CheckAll(fast, 0, 0), nil)
	assertFindings(t, CheckAll(fast, 0, 1), []Finding{
		{1, "SV007", "reading speed 20.0 characters per second, the limit is 1"},
	})
	assertFindings(t, CheckAll(nil, 42, 17), nil)
}

func TestCheckAllSpeedThresholds(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		limit int
		want  []Finding
	}{
		{"below the limit", strings.Repeat("x", 16), 17, nil},
		{"at the limit", strings.Repeat("x", 17), 17, nil},
		{"one above the limit", strings.Repeat("x", 18), 17,
			[]Finding{{1, "SV007", "reading speed 18.0 characters per second, the limit is 17"}}},
		{"100 characters at limit 100", strings.Repeat("x", 100), 100, nil},
		{"101 characters at limit 100", strings.Repeat("x", 101), 100,
			[]Finding{{1, "SV007", "reading speed 101.0 characters per second, the limit is 100"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := []srt.Cue{cue(1, 1000, 2000, 1, 3, tc.text)}
			assertFindings(t, CheckAll(cues, 0, tc.limit), tc.want)
		})
	}
}

func TestCheckAllCountsTextCodePoints(t *testing.T) {
	tests := []struct {
		name  string
		text  []string
		limit int
		want  []Finding
	}{
		{"multiple lines exclude line separators", []string{"hello", "world"}, 9,
			[]Finding{{1, "SV007", "reading speed 10.0 characters per second, the limit is 9"}}},
		{"tags count as text", []string{"<i>x</i>"}, 7,
			[]Finding{{1, "SV007", "reading speed 8.0 characters per second, the limit is 7"}}},
		{"multibyte character counts as one code point", []string{"é"}, 1, nil},
		{"emoji counts as one code point", []string{"😀"}, 1, nil},
		{"combining mark counts as a code point", []string{"e\u0301"}, 1,
			[]Finding{{1, "SV007", "reading speed 2.0 characters per second, the limit is 1"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := []srt.Cue{cue(1, 1000, 2000, 1, 3, tc.text...)}
			assertFindings(t, CheckAll(cues, 0, tc.limit), tc.want)
		})
	}
}

func TestCheckAllRoundsSpeedHalfUp(t *testing.T) {
	tests := []struct {
		name     string
		duration int64
		limit    int
		message  string
	}{
		{"one character in 32 milliseconds", 32,
			31, "reading speed 31.3 characters per second, the limit is 31"},
		{"exact half tenth rounds up", 160,
			1, "reading speed 6.3 characters per second, the limit is 1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := []srt.Cue{cue(1, 1000, 1000+tc.duration, 1, 3, "x")}
			assertFindings(t, CheckAll(cues, 0, tc.limit), []Finding{{1, "SV007", tc.message}})
		})
	}
}

func TestCheckAllSpeedEligibilityAndLine(t *testing.T) {
	tests := []struct {
		name string
		cue  srt.Cue
		want []Finding
	}{
		{"no text", cue(1, 1000, 2000, 1, 0),
			[]Finding{{1, "SV005", msgSV005}}},
		{"end equals start", cue(1, 1000, 1000, 1, 3, "too fast"),
			[]Finding{{1, "SV002", msgSV002}}},
		{"end before start", cue(1, 1000, 999, 1, 3, "too fast"),
			[]Finding{{1, "SV002", msgSV002}}},
		{"finding points at cue-number line", cue(1, 1000, 2000, 41, 43, "too fast"),
			[]Finding{{41, "SV007", "reading speed 8.0 characters per second, the limit is 1"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, CheckAll([]srt.Cue{tc.cue}, 0, 1), tc.want)
		})
	}
}

func TestCheckAllSpeedFindingOrder(t *testing.T) {
	t.Run("SV001 precedes SV007", func(t *testing.T) {
		cues := []srt.Cue{cue(2, 1000, 1100, 7, 9, "xx")}
		assertFindings(t, CheckAll(cues, 0, 1), []Finding{
			{7, "SV001", "cue number 2, expected 1"},
			{7, "SV007", "reading speed 20.0 characters per second, the limit is 1"},
		})
	})

	t.Run("all SV006 findings precede SV007", func(t *testing.T) {
		cues := []srt.Cue{cue(1, 1000, 2000, 1, 3, "abcd", "efgh")}
		assertFindings(t, CheckAll(cues, 3, 1), []Finding{
			{3, "SV006", "line 1 has 4 characters, the limit is 3"},
			{4, "SV006", "line 2 has 4 characters, the limit is 3"},
			{1, "SV007", "reading speed 8.0 characters per second, the limit is 1"},
		})
	})

	t.Run("cue order and other rules are preserved", func(t *testing.T) {
		cues := []srt.Cue{
			cue(1, 2000, 3000, 10, 12, "ab"),
			cue(2, 1000, 1100, 20, 22, "ab"),
		}
		assertFindings(t, CheckAll(cues, 0, 1), []Finding{
			{10, "SV007", "reading speed 2.0 characters per second, the limit is 1"},
			{20, "SV004", "starts at 00:00:01,000, before the previous cue starts at 00:00:02,000"},
			{20, "SV007", "reading speed 20.0 characters per second, the limit is 1"},
		})
	})
}
