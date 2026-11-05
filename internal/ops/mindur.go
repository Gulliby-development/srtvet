package ops

import "github.com/Gulliby-development/srtvet/internal/srt"

// MinDuration lengthens every cue that lasts less than ms milliseconds so
// that it lasts ms, but its end never passes the start of the next cue when
// that start is later than the cue's own start. A cue is never shortened,
// and a cue whose end is not after its start is left as it is.
func MinDuration(cues []srt.Cue, ms int64) {
	for i := range cues {
		if cues[i].End <= cues[i].Start || cues[i].End-cues[i].Start >= ms {
			continue
		}
		newEnd := cues[i].Start + ms
		if i+1 < len(cues) && cues[i+1].Start > cues[i].Start && newEnd > cues[i+1].Start {
			newEnd = cues[i+1].Start
		}
		if newEnd > cues[i].End {
			cues[i].End = newEnd
		}
	}
}
