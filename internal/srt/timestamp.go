package srt

import (
	"fmt"
	"regexp"
	"strconv"
)

// MaxTime is the largest time in milliseconds (99:59:59,999).
const MaxTime int64 = 359999999

var timingRE = regexp.MustCompile(`^([0-9]{2}):([0-5][0-9]):([0-5][0-9])[,.]([0-9]{3})[ \t]+-->[ \t]+([0-9]{2}):([0-5][0-9]):([0-5][0-9])[,.]([0-9]{3})[ \t]*$`)

// FormatTimestamp formats a millisecond value as HH:MM:SS,mmm.
func FormatTimestamp(ms int64) string {
	h := ms / 3600000
	m := ms / 60000 % 60
	s := ms / 1000 % 60
	return fmt.Sprintf("%02d:%02d:%02d,%03d", h, m, s, ms%1000)
}

// ParseTiming parses a timing line such as
// "00:00:01,000 --> 00:00:03,500" into start and end milliseconds.
func ParseTiming(line string) (start, end int64, ok bool) {
	m := timingRE.FindStringSubmatch(line)
	if m == nil {
		return 0, 0, false
	}
	return fieldsToMillis(m[1:5]), fieldsToMillis(m[5:9]), true
}

// fieldsToMillis turns the matched hours, minutes, seconds and
// milliseconds digits into one millisecond value.
func fieldsToMillis(f []string) int64 {
	var v [4]int64
	for i, s := range f {
		// The pattern guarantees short runs of decimal digits.
		n, _ := strconv.ParseInt(s, 10, 64)
		v[i] = n
	}
	return v[0]*3600000 + v[1]*60000 + v[2]*1000 + v[3]
}
