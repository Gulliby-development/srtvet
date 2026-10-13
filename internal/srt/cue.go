// Package srt holds the SubRip subtitle model, its parser and its writer.
package srt

// Cue is one subtitle entry of a SubRip file.
type Cue struct {
	// Number is the declared cue number.
	Number int
	// Start and End are times in whole milliseconds.
	Start, End int64
	// Text holds the text lines exactly as written.
	Text []string
	// Line is the line number of the cue-number line.
	Line int
	// TextLine is the line number of the first text line, 0 for a cue
	// without text.
	TextLine int
}
