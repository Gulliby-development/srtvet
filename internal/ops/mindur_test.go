package ops

import (
	"testing"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

func TestMinDuration(t *testing.T) {
	tests := []struct {
		name string
		ms   int64
		in   [][2]int64
		want [][2]int64
	}{
		{"short cue grows to the minimum", 1000, [][2]int64{{1000, 1200}}, [][2]int64{{1000, 2000}}},
		{"long enough cue stays", 1000, [][2]int64{{1000, 2000}}, [][2]int64{{1000, 2000}}},
		{"longer cue stays", 1000, [][2]int64{{1000, 5000}}, [][2]int64{{1000, 5000}}},
		{"one millisecond short", 1000, [][2]int64{{1000, 1999}}, [][2]int64{{1000, 2000}}},
		{"stops at the start of the next cue", 1000, [][2]int64{{1000, 1200}, {1800, 3000}}, [][2]int64{{1000, 1800}, {1800, 3000}}},
		{"next cue starts exactly at the minimum", 1000, [][2]int64{{1000, 1200}, {2000, 3000}}, [][2]int64{{1000, 2000}, {2000, 3000}}},
		{"next cue starts before the cue ends", 1000, [][2]int64{{1000, 1500}, {1200, 3000}}, [][2]int64{{1000, 1500}, {1200, 3000}}},
		{"next cue starts with the cue", 1000, [][2]int64{{1000, 1200}, {1000, 3000}}, [][2]int64{{1000, 2000}, {1000, 3000}}},
		{"next cue starts before the cue", 1000, [][2]int64{{1000, 1200}, {500, 900}}, [][2]int64{{1000, 2000}, {500, 1500}}},
		{"cue is never shortened", 1000, [][2]int64{{1000, 1900}, {1950, 4000}}, [][2]int64{{1000, 1950}, {1950, 4000}}},
		{"end equal to start is left alone", 1000, [][2]int64{{1000, 1000}}, [][2]int64{{1000, 1000}}},
		{"end before start is left alone", 1000, [][2]int64{{2000, 1000}}, [][2]int64{{2000, 1000}}},
		{"each cue looks at the start of the next one", 500, [][2]int64{{0, 100}, {300, 350}, {1000, 1100}}, [][2]int64{{0, 300}, {300, 800}, {1000, 1500}}},
		{"no cue", 1000, nil, nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := make([]srt.Cue, len(tc.in))
			for i, p := range tc.in {
				cues[i] = cue(i+1, p[0], p[1], 1, 3, "x")
			}
			MinDuration(cues, tc.ms)
			for i, p := range tc.want {
				if cues[i].Start != tc.in[i][0] {
					t.Errorf("cue %d start = %d, want %d", i+1, cues[i].Start, tc.in[i][0])
				}
				if cues[i].End != p[1] {
					t.Errorf("cue %d end = %d, want %d", i+1, cues[i].End, p[1])
				}
			}
		})
	}
}

func TestMinDurationKeepsTextAndNumbers(t *testing.T) {
	cues := []srt.Cue{cue(7, 0, 100, 1, 3, "one", "two"), cue(9, 5000, 5100, 6, 8)}
	MinDuration(cues, 700)
	if cues[0].Number != 7 || cues[1].Number != 9 {
		t.Errorf("numbers = %d, %d, want 7, 9", cues[0].Number, cues[1].Number)
	}
	assertTextLines(t, cues[0].Text, []string{"one", "two"})
	if len(cues[1].Text) != 0 || cues[1].End != 5700 {
		t.Errorf("empty cue = %+v, want no text and end 5700", cues[1])
	}
}
