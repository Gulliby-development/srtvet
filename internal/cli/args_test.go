package cli

import (
	"reflect"
	"testing"

	"github.com/Gulliby-development/srtvet/internal/ops"
)

func TestParseArgsHelp(t *testing.T) {
	opts, err := parseArgs([]string{"--help"})
	if err != nil {
		t.Fatalf("parseArgs(--help) error = %v, want nil", err)
	}
	if !opts.help {
		t.Errorf("parseArgs(--help) help = false, want true")
	}
}

func TestParseArgsAccepted(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want options
	}{
		{"check one file", []string{"check", "a.srt"},
			options{command: "check", files: []string{"a.srt"}, maxLine: 42}},
		{"max-line zero", []string{"check", "--max-line", "0", "a.srt"},
			options{command: "check", files: []string{"a.srt"}, maxLine: 0}},
		{"max-line leading zero", []string{"check", "--max-line", "042", "a.srt"},
			options{command: "check", files: []string{"a.srt"}, maxLine: 42}},
		{"max-line equals form", []string{"check", "--max-line=007", "a.srt"},
			options{command: "check", files: []string{"a.srt"}, maxLine: 7}},
		{"max-line largest", []string{"check", "--max-line", "200", "a.srt"},
			options{command: "check", files: []string{"a.srt"}, maxLine: 200}},
		{"option among files and stdin", []string{"check", "a.srt", "--max-line", "80", "b.srt", "-"},
			options{command: "check", files: []string{"a.srt", "b.srt", "-"}, maxLine: 80}},
		{"repeated max-line keeps last", []string{"check", "--max-line", "5", "--max-line=9", "a.srt"},
			options{command: "check", files: []string{"a.srt"}, maxLine: 9}},
		{"repeated invalid then valid", []string{"check", "--max-line", "abc", "--max-line", "9", "a.srt"},
			options{command: "check", files: []string{"a.srt"}, maxLine: 9}},
		{"fix one file", []string{"fix", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}}},
		{"fix stdin", []string{"fix", "-"},
			options{command: "fix", files: []string{"-"}}},
		{"shift separate value", []string{"fix", "--shift", "1s", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, shift: 1000, hasShift: true}},
		{"shift equals form", []string{"fix", "--shift=1s", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, shift: 1000, hasShift: true}},
		{"shift value starting with dash", []string{"fix", "--shift", "-0.5s", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, shift: -500, hasShift: true}},
		{"shift of zero is given", []string{"fix", "--shift", "0", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, shift: 0, hasShift: true}},
		{"option after the file", []string{"fix", "a.srt", "--shift", "1s"},
			options{command: "fix", files: []string{"a.srt"}, shift: 1000, hasShift: true}},
		{"repeated shift keeps last", []string{"fix", "--shift", "5s", "--shift", "1s", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, shift: 1000, hasShift: true}},
		{"stretch ratio", []string{"fix", "--stretch", "25/23.976", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, stretch: ops.Ratio{Num: 25000000, Den: 23976000}, hasStretch: true}},
		{"stretch largest plain", []string{"fix", "--stretch=1000", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, stretch: ops.Ratio{Num: 1000000000, Den: 1000000}, hasStretch: true}},
		{"sort only", []string{"fix", "--sort", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, sort: true}},
		{"drop-empty only", []string{"fix", "a.srt", "--drop-empty"},
			options{command: "fix", files: []string{"a.srt"}, dropEmpty: true}},
		{"out separate value", []string{"fix", "--out", "o.srt", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, out: "o.srt", hasOut: true}},
		{"out equals form", []string{"fix", "--out=o.srt", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, out: "o.srt", hasOut: true}},
		{"out takes the next argument whatever it is", []string{"fix", "--out", "--sort", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, out: "--sort", hasOut: true}},
		{"repeated out keeps last", []string{"fix", "--out", "x.srt", "a.srt", "--out=y.srt"},
			options{command: "fix", files: []string{"a.srt"}, out: "y.srt", hasOut: true}},
		{"all fix options", []string{"fix", "--stretch", "24/25", "--shift", "-250ms", "--sort", "--drop-empty", "--out", "o.srt", "a.srt"},
			options{command: "fix", files: []string{"a.srt"}, shift: -250, hasShift: true,
				stretch: ops.Ratio{Num: 24000000, Den: 25000000}, hasStretch: true,
				sort: true, dropEmpty: true, out: "o.srt", hasOut: true}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseArgs(tc.args)
			if err != nil {
				t.Fatalf("parseArgs(%q) error = %v, want nil", tc.args, err)
			}
			want := tc.want
			if want.command == "fix" {
				// The line limit has no meaning for fix.
				want.maxLine = got.maxLine
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("parseArgs(%q)\n got %+v\nwant %+v", tc.args, got, want)
			}
		})
	}
}

func TestParseArgsErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no argument", nil, `srtvet: missing command (check or fix)`},
		{"unknown command", []string{"frobnicate"}, `srtvet: unknown command "frobnicate"`},
		{"unknown command with dash", []string{"-x"}, `srtvet: unknown command "-x"`},
		{"help with another argument", []string{"--help", "check"}, `srtvet: unknown command "--help"`},
		{"help twice", []string{"--help", "--help"}, `srtvet: unknown command "--help"`},
		{"unknown command before a file", []string{"Check", "a.srt"}, `srtvet: unknown command "Check"`},

		{"max-line without value", []string{"check", "a.srt", "--max-line"}, `srtvet: --max-line requires a value`},
		{"shift without value", []string{"fix", "a.srt", "--shift"}, `srtvet: --shift requires a value`},
		{"stretch without value", []string{"fix", "a.srt", "--stretch"}, `srtvet: --stretch requires a value`},
		{"out without value", []string{"fix", "a.srt", "--out"}, `srtvet: --out requires a value`},
		{"missing value before value checks", []string{"fix", "--shift", "x", "--out"}, `srtvet: --out requires a value`},

		{"unknown option after a file", []string{"check", "a.srt", "--bogus"}, `srtvet: unknown option "--bogus"`},
		{"help after a command", []string{"check", "a.srt", "--help"}, `srtvet: unknown option "--help"`},
		{"help after fix", []string{"fix", "--help"}, `srtvet: unknown option "--help"`},
		{"double dash", []string{"fix", "a.srt", "--"}, `srtvet: unknown option "--"`},
		{"short option", []string{"check", "-x", "a.srt"}, `srtvet: unknown option "-x"`},
		{"fix option for check", []string{"check", "--shift", "1s", "a.srt"}, `srtvet: unknown option "--shift"`},
		{"fix option for check, equals form", []string{"check", "--shift=1s", "a.srt"}, `srtvet: unknown option "--shift=1s"`},
		{"sort for check", []string{"check", "--sort", "a.srt"}, `srtvet: unknown option "--sort"`},
		{"drop-empty for check", []string{"check", "--drop-empty", "a.srt"}, `srtvet: unknown option "--drop-empty"`},
		{"out for check", []string{"check", "--out", "o.srt", "a.srt"}, `srtvet: unknown option "--out"`},
		{"check option for fix", []string{"fix", "--max-line", "5", "a.srt"}, `srtvet: unknown option "--max-line"`},
		{"sort with a value", []string{"fix", "--sort=x", "a.srt"}, `srtvet: unknown option "--sort=x"`},
		{"drop-empty with a value", []string{"fix", "--drop-empty=x", "a.srt"}, `srtvet: unknown option "--drop-empty=x"`},
		{"unknown with a value", []string{"fix", "--foo=bar", "a.srt"}, `srtvet: unknown option "--foo=bar"`},
		{"unknown option before value checks", []string{"check", "--max-line", "999", "--bogus"}, `srtvet: unknown option "--bogus"`},

		{"max-line 201", []string{"check", "--max-line", "201", "a.srt"}, `srtvet: invalid --max-line "201"`},
		{"max-line four digits", []string{"check", "--max-line=1000", "a.srt"}, `srtvet: invalid --max-line "1000"`},
		{"max-line four digits with zeros", []string{"check", "--max-line", "0042", "a.srt"}, `srtvet: invalid --max-line "0042"`},
		{"max-line negative", []string{"check", "--max-line", "-1", "a.srt"}, `srtvet: invalid --max-line "-1"`},
		{"max-line with plus", []string{"check", "--max-line", "+5", "a.srt"}, `srtvet: invalid --max-line "+5"`},
		{"max-line letters", []string{"check", "--max-line", "abc", "a.srt"}, `srtvet: invalid --max-line "abc"`},
		{"max-line empty after equals", []string{"check", "--max-line=", "a.srt"}, `srtvet: invalid --max-line ""`},
		{"max-line empty argument", []string{"check", "--max-line", "", "a.srt"}, `srtvet: invalid --max-line ""`},
		{"max-line checked before the file count", []string{"check", "--max-line", "201"}, `srtvet: invalid --max-line "201"`},
		{"check without a file", []string{"check"}, `srtvet: check needs at least one FILE`},
		{"check with only an option", []string{"check", "--max-line", "10"}, `srtvet: check needs at least one FILE`},

		{"fix without a file", []string{"fix"}, `srtvet: fix needs exactly one FILE`},
		{"fix with two files", []string{"fix", "a.srt", "b.srt"}, `srtvet: fix needs exactly one FILE`},
		{"fix with only options", []string{"fix", "--sort", "--shift", "1s"}, `srtvet: fix needs exactly one FILE`},
		{"shift checked before the file count", []string{"fix", "--shift", ".5"}, `srtvet: invalid --shift ".5"`},
		{"shift fraction with ms", []string{"fix", "--shift", "1.5ms", "a.srt"}, `srtvet: invalid --shift "1.5ms"`},
		{"shift empty after equals", []string{"fix", "--shift=", "a.srt"}, `srtvet: invalid --shift ""`},
		{"shift with a space", []string{"fix", "--shift", "1 s", "a.srt"}, `srtvet: invalid --shift "1 s"`},
		{"shift checked before stretch", []string{"fix", "--shift", "x", "--stretch", "0"}, `srtvet: invalid --shift "x"`},
		{"shift checked before stretch whatever the order", []string{"fix", "--stretch", "0", "--shift", "x", "a.srt"}, `srtvet: invalid --shift "x"`},
		{"stretch checked before the file count", []string{"fix", "--shift", "1s", "--stretch", "0"}, `srtvet: invalid --stretch "0"`},
		{"stretch above 1000", []string{"fix", "--stretch", "1000.000001", "a.srt"}, `srtvet: invalid --stretch "1000.000001"`},
		{"stretch zero with fraction", []string{"fix", "--stretch", "0.000000", "a.srt"}, `srtvet: invalid --stretch "0.000000"`},
		{"stretch empty after equals", []string{"fix", "--stretch=", "a.srt"}, `srtvet: invalid --stretch ""`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseArgs(tc.args)
			if err == nil {
				t.Fatalf("parseArgs(%q) error = nil, want %q", tc.args, tc.want)
			}
			if err.Error() != tc.want {
				t.Errorf("parseArgs(%q) error\n got %q\nwant %q", tc.args, err.Error(), tc.want)
			}
		})
	}
}
