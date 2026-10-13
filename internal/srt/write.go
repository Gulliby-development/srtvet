package srt

import (
	"bytes"
	"io"
	"strconv"
)

// Write emits the cues in SubRip form: LF line endings, no BOM, one blank
// line between two cues and none after the last. Zero cues give zero
// bytes. It returns the first write error.
func Write(w io.Writer, cues []Cue) error {
	if len(cues) == 0 {
		return nil
	}
	var buf bytes.Buffer
	for i, c := range cues {
		if i > 0 {
			buf.WriteByte('\n')
		}
		buf.WriteString(strconv.Itoa(c.Number))
		buf.WriteByte('\n')
		buf.WriteString(FormatTimestamp(c.Start))
		buf.WriteString(" --> ")
		buf.WriteString(FormatTimestamp(c.End))
		buf.WriteByte('\n')
		for _, line := range c.Text {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	_, err := w.Write(buf.Bytes())
	return err
}
