package ops

import (
	"regexp"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

var formattingTagRE = regexp.MustCompile(`(?i)</?(?:i|b|u|s|font)(?:[ \t][^<>]*)?>|\{\\[^}]*\}`)

// StripTags removes supported formatting tags and override blocks from cue
// text, dropping only lines that become empty.
func StripTags(cues []srt.Cue) {
	for i := range cues {
		text := make([]string, 0, len(cues[i].Text))
		for _, line := range cues[i].Text {
			line = formattingTagRE.ReplaceAllString(line, "")
			if line != "" {
				text = append(text, line)
			}
		}
		cues[i].Text = text
	}
}
