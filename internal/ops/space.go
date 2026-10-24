package ops

import (
	"strings"
	"unicode/utf8"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

// CleanSpace removes invisible formatting characters and normalizes runs of
// odd spaces in a subtitle text line.
func CleanSpace(line string) string {
	line = deleteInvisibleSpaceCharacters(line)

	var cleaned strings.Builder
	runStart := -1
	hasOddSpace := false

	flushRun := func(end int, atLineEnd bool) {
		if runStart < 0 {
			return
		}
		if hasOddSpace {
			if runStart > 0 && !atLineEnd {
				cleaned.WriteByte(' ')
			}
		} else {
			cleaned.WriteString(line[runStart:end])
		}
		runStart = -1
		hasOddSpace = false
	}

	for i := 0; i < len(line); {
		r, size := utf8.DecodeRuneInString(line[i:])
		if isSpaceRunRune(r) {
			if runStart < 0 {
				runStart = i
			}
			if r != ' ' {
				hasOddSpace = true
			}
			i += size
			continue
		}

		flushRun(i, false)
		cleaned.WriteString(line[i : i+size])
		i += size
	}

	flushRun(len(line), true)
	return cleaned.String()
}

func deleteInvisibleSpaceCharacters(line string) string {
	var cleaned strings.Builder
	segmentStart := 0

	for i := 0; i < len(line); {
		r, size := utf8.DecodeRuneInString(line[i:])
		if isDeletedSpaceRune(r) {
			cleaned.WriteString(line[segmentStart:i])
			i += size
			segmentStart = i
			continue
		}
		i += size
	}

	if segmentStart == 0 {
		return line
	}
	cleaned.WriteString(line[segmentStart:])
	return cleaned.String()
}

func isDeletedSpaceRune(r rune) bool {
	switch r {
	case '\u00ad', '\u200b', '\u2060', '\ufeff':
		return true
	default:
		return false
	}
}

func isSpaceRunRune(r rune) bool {
	if r == ' ' || (r >= '\u2000' && r <= '\u200a') {
		return true
	}
	switch r {
	case '\u00a0', '\u202f', '\u205f', '\u3000', '\u0085', '\u2028', '\u2029', '\r':
		return true
	default:
		return false
	}
}

// CleanSpaceLines cleans every cue text line and drops lines that become empty.
func CleanSpaceLines(cues []srt.Cue) {
	for i := range cues {
		text := make([]string, 0, len(cues[i].Text))
		for _, line := range cues[i].Text {
			line = CleanSpace(line)
			if line != "" {
				text = append(text, line)
			}
		}
		cues[i].Text = text
	}
}
