package ops

import (
	"sort"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

// DropEmpty returns a new slice without the cues that have no text
// lines; the order of the others is kept.
func DropEmpty(cues []srt.Cue) []srt.Cue {
	out := make([]srt.Cue, 0, len(cues))
	for _, c := range cues {
		if len(c.Text) > 0 {
			out = append(out, c)
		}
	}
	return out
}

// Sort orders the cues by start, then by end; equal cues keep their
// order.
func Sort(cues []srt.Cue) {
	sort.SliceStable(cues, func(i, j int) bool {
		if cues[i].Start != cues[j].Start {
			return cues[i].Start < cues[j].Start
		}
		return cues[i].End < cues[j].End
	})
}

// Renumber sets the cue numbers to 1, 2, 3, ... in slice order.
func Renumber(cues []srt.Cue) {
	for i := range cues {
		cues[i].Number = i + 1
	}
}
