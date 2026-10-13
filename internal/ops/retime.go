package ops

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

const (
	// ratioScale is the factor applied to both parts of a ratio.
	ratioScale int64 = 1000000
	// ratioMax is the largest scaled ratio part (1000).
	ratioMax = 1000 * ratioScale
)

var (
	secondsRE   = regexp.MustCompile(`^([+-]?)([0-9]{1,5})(?:[.]([0-9]{1,3}))?s?$`)
	millisRE    = regexp.MustCompile(`^([+-]?)([0-9]{1,8})ms$`)
	ratioPartRE = regexp.MustCompile(`^([0-9]{1,4})(?:[.]([0-9]{1,6}))?$`)
)

// Ratio is a stretch factor Num/Den; both parts are scaled by 1,000,000.
type Ratio struct {
	Num, Den int64
}

// scaleDecimal returns the decimal number whole.frac multiplied by
// 10^digits as an integer; frac is padded on the right to digits digits.
// The callers pass short runs of decimal digits only.
func scaleDecimal(whole, frac string, digits int) int64 {
	n, _ := strconv.ParseInt(whole+frac+strings.Repeat("0", digits-len(frac)), 10, 64)
	return n
}

// ParseDuration parses a --shift value into whole milliseconds.
func ParseDuration(s string) (ms int64, ok bool) {
	var sign string
	if m := millisRE.FindStringSubmatch(s); m != nil {
		sign, ms = m[1], scaleDecimal(m[2], "", 0)
	} else if m := secondsRE.FindStringSubmatch(s); m != nil {
		sign, ms = m[1], scaleDecimal(m[2], m[3], 3)
	} else {
		return 0, false
	}
	if sign == "-" {
		ms = -ms
	}
	return ms, true
}

// parseRatioPart parses one side of a ratio into its scaled value.
func parseRatioPart(s string) (int64, bool) {
	m := ratioPartRE.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	v := scaleDecimal(m[1], m[2], 6)
	if v <= 0 || v > ratioMax {
		return 0, false
	}
	return v, true
}

// ParseRatio parses a --stretch value of the form X or X/Y.
func ParseRatio(s string) (Ratio, bool) {
	parts := strings.Split(s, "/")
	if len(parts) > 2 {
		return Ratio{}, false
	}
	r := Ratio{Den: ratioScale}
	var ok bool
	if r.Num, ok = parseRatioPart(parts[0]); !ok {
		return Ratio{}, false
	}
	if len(parts) == 2 {
		if r.Den, ok = parseRatioPart(parts[1]); !ok {
			return Ratio{}, false
		}
	}
	return r, true
}

// Stretch replaces every start and end t by (t*Num + Den/2) / Den,
// which rounds half up.
func Stretch(cues []srt.Cue, r Ratio) {
	for i := range cues {
		cues[i].Start = (cues[i].Start*r.Num + r.Den/2) / r.Den
		cues[i].End = (cues[i].End*r.Num + r.Den/2) / r.Den
	}
}

// Shift adds ms to every start and end.
func Shift(cues []srt.Cue, ms int64) {
	for i := range cues {
		cues[i].Start += ms
		cues[i].End += ms
	}
}

// CheckRange reports the first cue whose start is below 0 or whose end
// is above srt.MaxTime.
func CheckRange(cues []srt.Cue) error {
	for _, c := range cues {
		if c.Start < 0 || c.End > srt.MaxTime {
			return fmt.Errorf("srtvet: cue %d: time out of range after retiming", c.Number)
		}
	}
	return nil
}
