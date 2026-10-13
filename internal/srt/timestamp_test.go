package srt

import "testing"

func TestMaxTime(t *testing.T) {
	if MaxTime != 359999999 {
		t.Fatalf(`MaxTime = %d, want 359999999`, MaxTime)
	}
}

func TestFormatTimestamp(t *testing.T) {
	tests := []struct {
		ms   int64
		want string
	}{
		{0, `00:00:00,000`},
		{1, `00:00:00,001`},
		{999, `00:00:00,999`},
		{1000, `00:00:01,000`},
		{59999, `00:00:59,999`},
		{60000, `00:01:00,000`},
		{72040, `00:01:12,040`},
		{3599999, `00:59:59,999`},
		{3600000, `01:00:00,000`},
		{MaxTime, `99:59:59,999`},
	}
	for _, tc := range tests {
		if got := FormatTimestamp(tc.ms); got != tc.want {
			t.Errorf(`FormatTimestamp(%d) = %q, want %q`, tc.ms, got, tc.want)
		}
	}
}

func TestParseTimingAccepted(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		start, end int64
	}{
		{`comma`, `00:00:01,000 --> 00:00:03,500`, 1000, 3500},
		{`dot and two trailing spaces`, `00:00:01.000 --> 00:00:02.000  `, 1000, 2000},
		{`a tab on each side of the arrow`, `00:00:01,000` + tab + `-->` + tab + `00:00:02,000`, 1000, 2000},
		{`mixed blanks and a trailing tab`, `00:00:00,000 ` + tab + ` -->  ` + tab + `00:59:59,999` + tab, 0, 3599999},
		{`comma then dot`, `01:02:03,004 --> 05:06:07.008`, 3723004, 18367008},
		{`hour boundary`, `00:59:59,999 --> 01:00:00,000`, 3599999, 3600000},
		{`zero on both sides`, `00:00:00,000 --> 00:00:00,000`, 0, 0},
		{`largest time on both sides`, `99:59:59,999 --> 99:59:59,999`, MaxTime, MaxTime},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start, end, ok := ParseTiming(tc.line)
			if !ok {
				t.Fatalf(`ParseTiming(%q) rejected the line`, tc.line)
			}
			if start != tc.start || end != tc.end {
				t.Errorf(`ParseTiming(%q) = %d, %d, want %d, %d`, tc.line, start, end, tc.start, tc.end)
			}
		})
	}
}

func TestParseTimingRejected(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{`minutes 60 at the start`, `00:60:00,000 --> 00:00:02,000`},
		{`minutes 60 at the end`, `00:00:01,000 --> 00:60:02,000`},
		{`seconds 60 at the start`, `00:00:60,000 --> 00:01:02,000`},
		{`seconds 60 at the end`, `00:00:01,000 --> 00:00:60,000`},
		{`short arrow`, `00:00:01,000 -> 00:00:02,000`},
		{`long arrow`, `00:00:01,000 ---> 00:00:02,000`},
		{`no arrow`, `00:00:01,000 00:00:02,000`},
		{`no space around the arrow`, `00:00:01,000-->00:00:02,000`},
		{`no space before the arrow`, `00:00:01,000--> 00:00:02,000`},
		{`no space after the arrow`, `00:00:01,000 -->00:00:02,000`},
		{`one-digit hours`, `0:00:01,000 --> 00:00:02,000`},
		{`three-digit hours`, `100:00:01,000 --> 00:00:02,000`},
		{`two-digit milliseconds at the start`, `00:00:01,00 --> 00:00:02,000`},
		{`two-digit milliseconds at the end`, `00:00:01,000 --> 00:00:02,00`},
		{`four-digit milliseconds`, `00:00:01,0000 --> 00:00:02,000`},
		{`colon before the milliseconds`, `00:00:01:000 --> 00:00:02,000`},
		{`leading space`, ` 00:00:01,000 --> 00:00:02,000`},
		{`leading tab`, tab + `00:00:01,000 --> 00:00:02,000`},
		{`trailing text`, `00:00:01,000 --> 00:00:02,000 x`},
		{`trailing CR`, `00:00:01,000 --> 00:00:02,000` + cr},
		{`only one timestamp`, `00:00:01,000 -->`},
		{`empty line`, ``},
		{`line of spaces`, `   `},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if start, end, ok := ParseTiming(tc.line); ok {
				t.Errorf(`ParseTiming(%q) = %d, %d, accepted; want it rejected`, tc.line, start, end)
			}
		})
	}
}
