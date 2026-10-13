package ops

import (
	"testing"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

func mustRatio(t *testing.T, s string) Ratio {
	t.Helper()
	r, ok := ParseRatio(s)
	if !ok {
		t.Fatalf("ParseRatio(%q) was rejected", s)
	}
	return r
}

func TestParseDurationAccepted(t *testing.T) {
	tests := []struct {
		in   string
		want int64
	}{
		{"+1.5s", 1500},
		{"-0.5s", -500},
		{"2", 2000},
		{"0.040", 40},
		{"-12s", -12000},
		{"1.5", 1500},
		{"99999.999s", 99999999},
		{"-250ms", -250},
		{"1500ms", 1500},
		{"99999999ms", 99999999},
		{"0", 0},
		{"+0ms", 0},
		{"0.04", 40},
		{"1.05", 1050},
		{"-99999.999", -99999999},
		{"-99999999ms", -99999999},
		{"1s", 1000},
		{"1ms", 1},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, ok := ParseDuration(tc.in)
			if !ok {
				t.Fatalf("ParseDuration(%q) was rejected, want %d", tc.in, tc.want)
			}
			if got != tc.want {
				t.Errorf("ParseDuration(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseDurationRejected(t *testing.T) {
	tests := []string{
		".5", "1.5ms", "1 s", "", "100000", "1.2345", "123456789ms", "1.", "s", "ms",
		"+", "1S", "1,5", "abc", " 1", "1 ", "1.s", "--1", "1sms", "1mss", "100000s",
	}
	for _, in := range tests {
		t.Run(in, func(t *testing.T) {
			if got, ok := ParseDuration(in); ok {
				t.Errorf("ParseDuration(%q) = %d, want it rejected", in, got)
			}
		})
	}
}

func TestParseRatioAccepted(t *testing.T) {
	tests := []struct {
		in       string
		num, den int64
	}{
		{"1.001", 1001000, 1000000},
		{"25/23.976", 25000000, 23976000},
		{"24/25", 24000000, 25000000},
		{"1000", 1000000000, 1000000},
		{"1000/1000", 1000000000, 1000000000},
		{"0.000001", 1, 1000000},
		{"1/0.000001", 1000000, 1},
		{"1", 1000000, 1000000},
		{"1000.000000", 1000000000, 1000000},
		{"999.999999/2", 999999999, 2000000},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, ok := ParseRatio(tc.in)
			if !ok {
				t.Fatalf("ParseRatio(%q) was rejected", tc.in)
			}
			if got.Num != tc.num || got.Den != tc.den {
				t.Errorf("ParseRatio(%q) = %d/%d, want %d/%d", tc.in, got.Num, got.Den, tc.num, tc.den)
			}
		})
	}
}

func TestParseRatioRejected(t *testing.T) {
	tests := []string{
		"0", "0.000000", "1000.000001", "1/0", "12345", "1.1234567", "/2", "1/", "1/2/3",
		"-1", "+1", ".5", "1.", "", "1001", "1/1000.000001", "1/0.000000", "1 /2", "1/-2",
		"abc", "1,5", "/",
	}
	for _, in := range tests {
		t.Run(in, func(t *testing.T) {
			if got, ok := ParseRatio(in); ok {
				t.Errorf("ParseRatio(%q) = %d/%d, want it rejected", in, got.Num, got.Den)
			}
		})
	}
}

func TestStretch(t *testing.T) {
	tests := []struct {
		ratio string
		in    []int64
		want  []int64
	}{
		{"1/2", []int64{1, 2, 3}, []int64{1, 1, 2}},
		{"3/2", []int64{1}, []int64{2}},
		{"25/23.976", []int64{1000, 3500, 70000, 72040}, []int64{1043, 3649, 72990, 75117}},
		{"1", []int64{0, 1, 359999999}, []int64{0, 1, 359999999}},
		{"2", []int64{0, 1000}, []int64{0, 2000}},
		{"1/3", []int64{1, 2, 4}, []int64{0, 1, 1}},
	}
	for _, tc := range tests {
		t.Run(tc.ratio, func(t *testing.T) {
			r := mustRatio(t, tc.ratio)
			n := len(tc.in)
			// Each cue starts at one value and ends at the mirrored one, so
			// Start and End are checked with different numbers.
			cues := make([]srt.Cue, n)
			for i := range cues {
				cues[i] = cue(i+1, tc.in[i], tc.in[n-1-i], 0, 0, "x")
			}
			Stretch(cues, r)
			for i, c := range cues {
				if c.Start != tc.want[i] {
					t.Errorf("start %d stretched by %s = %d, want %d", tc.in[i], tc.ratio, c.Start, tc.want[i])
				}
				if c.End != tc.want[n-1-i] {
					t.Errorf("end %d stretched by %s = %d, want %d", tc.in[n-1-i], tc.ratio, c.End, tc.want[n-1-i])
				}
				if c.Number != i+1 {
					t.Errorf("cue %d: number changed to %d", i+1, c.Number)
				}
			}
		})
	}
}

func TestShift(t *testing.T) {
	tests := []struct {
		name                 string
		ms                   int64
		wantStart1, wantEnd1 int64
		wantStart2, wantEnd2 int64
	}{
		{"zero changes nothing", 0, 1000, 3500, 70000, 72040},
		{"forward", 500, 1500, 4000, 70500, 72540},
		{"backward", -250, 750, 3250, 69750, 71790},
		{"below zero is not clamped", -2000, -1000, 1500, 68000, 70040},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := []srt.Cue{cue(1, 1000, 3500, 1, 3, "a"), cue(2, 70000, 72040, 5, 7, "b")}
			Shift(cues, tc.ms)
			if cues[0].Start != tc.wantStart1 || cues[0].End != tc.wantEnd1 {
				t.Errorf("first cue = %d to %d, want %d to %d", cues[0].Start, cues[0].End, tc.wantStart1, tc.wantEnd1)
			}
			if cues[1].Start != tc.wantStart2 || cues[1].End != tc.wantEnd2 {
				t.Errorf("second cue = %d to %d, want %d to %d", cues[1].Start, cues[1].End, tc.wantStart2, tc.wantEnd2)
			}
		})
	}
}

func TestStretchThenShift(t *testing.T) {
	cues := []srt.Cue{cue(1, 1000, 2000, 1, 3, "a")}
	Stretch(cues, mustRatio(t, "2"))
	Shift(cues, 500)
	if cues[0].Start != 2500 {
		t.Errorf("start = %d, want 2500 (stretch, then shift)", cues[0].Start)
	}
	if cues[0].End != 4500 {
		t.Errorf("end = %d, want 4500 (stretch, then shift)", cues[0].End)
	}
}

func TestCheckRange(t *testing.T) {
	const errCue1 = "srtvet: cue 1: time out of range after retiming"
	const errCue7 = "srtvet: cue 7: time out of range after retiming"
	const errCue8 = "srtvet: cue 8: time out of range after retiming"
	const errCue9 = "srtvet: cue 9: time out of range after retiming"
	tests := []struct {
		name    string
		cues    []srt.Cue
		shift   int64
		wantErr string
	}{
		{"second cue starts below zero", []srt.Cue{
			cue(3, 70000, 72040, 1, 3, "a"),
			cue(1, 1000, 3500, 5, 7, "b"),
		}, -2000, errCue1},
		{"end above the maximum", []srt.Cue{cue(9, 1000, 359999999, 1, 3, "a")}, 1, errCue9},
		{"start at zero passes", []srt.Cue{cue(9, 0, 1000, 1, 3, "a")}, 0, ""},
		{"end reaches the maximum", []srt.Cue{cue(9, 1000, 359999998, 1, 3, "a")}, 1, ""},
		{"start one below zero", []srt.Cue{cue(9, 0, 1000, 1, 3, "a")}, -1, errCue9},
		{"first offender is a start", []srt.Cue{
			cue(4, 0, 1000, 1, 3, "a"),
			cue(7, -1, 1000, 5, 7, "b"),
			cue(8, 1000, 360000000, 9, 11, "c"),
		}, 0, errCue7},
		{"first offender is an end", []srt.Cue{
			cue(4, 0, 1000, 1, 3, "a"),
			cue(8, 1000, 360000000, 5, 7, "b"),
			cue(7, -1, 1000, 9, 11, "c"),
		}, 0, errCue8},
		{"whole range in one cue", []srt.Cue{cue(1, 0, 359999999, 1, 3, "a")}, 0, ""},
		{"no cues", nil, 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Shift(tc.cues, tc.shift)
			err := CheckRange(tc.cues)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckRange = %q, want no error", err.Error())
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckRange = nil, want %q", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Errorf("CheckRange = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestCheckRangeBoundaryValues(t *testing.T) {
	if srt.MaxTime != 359999999 {
		t.Fatalf("srt.MaxTime = %d, want 359999999", srt.MaxTime)
	}
	low := []srt.Cue{cue(1, 0, 1000, 1, 3, "a")}
	Shift(low, 0)
	if err := CheckRange(low); err != nil || low[0].Start != 0 {
		t.Errorf("start = %d, err = %v; want start 0 and no error", low[0].Start, err)
	}
	high := []srt.Cue{cue(1, 1000, 359999998, 1, 3, "a")}
	Shift(high, 1)
	if err := CheckRange(high); err != nil || high[0].End != 359999999 {
		t.Errorf("end = %d, err = %v; want end 359999999 and no error", high[0].End, err)
	}
}
