package srt

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// The three states of the parser.
const (
	stateNumber = iota // expect a cue number
	stateTiming        // expect a timing line
	stateText          // in the text
)

const blankChars = " \t"

var bom = []byte{0xEF, 0xBB, 0xBF}

// splitLines splits the data at LF, removes one CR directly before each
// LF and drops the empty piece after a final LF.
func splitLines(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	last := len(lines) - 1
	for i := 0; i < last; i++ {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}
	if lines[last] == "" {
		lines = lines[:last]
	}
	return lines
}

// parseNumber reads a cue-number line: one to nine decimal digits after
// trimming spaces and tabs.
func parseNumber(line string) (int, bool) {
	s := strings.Trim(line, blankChars)
	if len(s) < 1 || len(s) > 9 {
		return 0, false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	return n, true
}

// Parse reads SubRip data into cues. The name is used in error texts.
// Only the first error in file order is returned.
func Parse(name string, data []byte) ([]Cue, error) {
	data = bytes.TrimPrefix(data, bom)
	var cues []Cue
	var cur Cue
	state := stateNumber
	for i, line := range splitLines(data) {
		lineNo := i + 1
		blank := strings.Trim(line, blankChars) == ""
		switch state {
		case stateNumber:
			if blank {
				continue
			}
			n, ok := parseNumber(line)
			if !ok {
				return nil, fmt.Errorf("srtvet: %s:%d: expected a cue number, found \"%s\"", name, lineNo, line)
			}
			cur = Cue{Number: n, Line: lineNo}
			state = stateTiming
		case stateTiming:
			start, end, ok := ParseTiming(line)
			if !ok {
				return nil, fmt.Errorf("srtvet: %s:%d: expected a timing line, found \"%s\"", name, lineNo, line)
			}
			cur.Start, cur.End = start, end
			state = stateText
		case stateText:
			if blank {
				cues = append(cues, cur)
				state = stateNumber
				continue
			}
			if len(cur.Text) == 0 {
				cur.TextLine = lineNo
			}
			cur.Text = append(cur.Text, line)
		}
	}
	switch state {
	case stateTiming:
		return nil, fmt.Errorf("srtvet: %s:%d: cue number without a timing line", name, cur.Line)
	case stateText:
		cues = append(cues, cur)
	}
	return cues, nil
}
