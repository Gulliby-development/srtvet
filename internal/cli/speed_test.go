package cli

import (
	"strings"
	"testing"
)

const fastSpeedSRT = "1\n" +
	"00:00:01,000 --> 00:00:02,000\n" +
	"Hello there, how are you?\n" +
	"\n" +
	"2\n" +
	"00:00:02,500 --> 00:00:04,500\n" +
	"Short.\n" +
	"\n" +
	"3\n" +
	"00:00:05,000 --> 00:00:05,400\n" +
	"Go now!\n" +
	"\n" +
	"4\n" +
	"00:00:06,000 --> 00:00:07,000\n" +
	"<i>Two lines</i>\n" +
	"of fast text here\n" +
	"\n" +
	"5\n" +
	"00:00:08,000 --> 00:00:09,000\n" +
	"Seventeen chars!!\n"

func TestParseArgsMaxCPSAccepted(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		maxCPS  int
		files   []string
		maxLine int
	}{
		{"default is off", []string{"check", "a.srt"}, 0, []string{"a.srt"}, 42},
		{"separate value", []string{"check", "--max-cps", "17", "a.srt"}, 17, []string{"a.srt"}, 42},
		{"equals form", []string{"check", "--max-cps=17", "a.srt"}, 17, []string{"a.srt"}, 42},
		{"zero", []string{"check", "--max-cps", "0", "a.srt"}, 0, []string{"a.srt"}, 42},
		{"zero with leading zero", []string{"check", "--max-cps=00", "a.srt"}, 0, []string{"a.srt"}, 42},
		{"leading zeros", []string{"check", "--max-cps", "017", "a.srt"}, 17, []string{"a.srt"}, 42},
		{"maximum", []string{"check", "--max-cps=100", "a.srt"}, 100, []string{"a.srt"}, 42},
		{"before files", []string{"check", "--max-cps", "17", "a.srt", "b.srt"}, 17, []string{"a.srt", "b.srt"}, 42},
		{"between files", []string{"check", "a.srt", "--max-cps=17", "b.srt"}, 17, []string{"a.srt", "b.srt"}, 42},
		{"after files", []string{"check", "a.srt", "b.srt", "--max-cps", "17"}, 17, []string{"a.srt", "b.srt"}, 42},
		{"repeated keeps last", []string{"check", "--max-cps", "5", "--max-cps=17", "a.srt"}, 17, []string{"a.srt"}, 42},
		{"invalid earlier value is replaced", []string{"check", "--max-cps", "abc", "--max-cps", "17", "a.srt"}, 17, []string{"a.srt"}, 42},

		{"max-line and max-cps", []string{"check", "--max-cps", "17", "--max-line", "0", "a.srt"}, 17, []string{"a.srt"}, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts, err := parseArgs(tc.args)
			if err != nil {
				t.Fatalf("parseArgs(%q) error = %v, want nil", tc.args, err)
			}
			if opts.maxCPS != tc.maxCPS {
				t.Errorf("maxCPS = %d, want %d", opts.maxCPS, tc.maxCPS)
			}
			if opts.maxLine != tc.maxLine {
				t.Errorf("maxLine = %d, want %d", opts.maxLine, tc.maxLine)
			}
			if len(opts.files) != len(tc.files) {
				t.Fatalf("files = %q, want %q", opts.files, tc.files)
			}
			for i := range tc.files {
				if opts.files[i] != tc.files[i] {
					t.Errorf("files[%d] = %q, want %q", i, opts.files[i], tc.files[i])
				}
			}
		})
	}
}

func TestParseArgsMaxCPSErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"above maximum", []string{"check", "--max-cps", "101", "a.srt"}, `srtvet: invalid --max-cps "101"`},
		{"empty equals value", []string{"check", "--max-cps=", "a.srt"}, `srtvet: invalid --max-cps ""`},
		{"empty separate value", []string{"check", "--max-cps", "", "a.srt"}, `srtvet: invalid --max-cps ""`},
		{"plus sign", []string{"check", "--max-cps", "+1", "a.srt"}, `srtvet: invalid --max-cps "+1"`},
		{"minus sign", []string{"check", "--max-cps", "-1", "a.srt"}, `srtvet: invalid --max-cps "-1"`},
		{"decimal", []string{"check", "--max-cps", "1.0", "a.srt"}, `srtvet: invalid --max-cps "1.0"`},
		{"whitespace", []string{"check", "--max-cps", "1 7", "a.srt"}, `srtvet: invalid --max-cps "1 7"`},
		{"four digits", []string{"check", "--max-cps", "0000", "a.srt"}, `srtvet: invalid --max-cps "0000"`},
		{"non-ASCII digits", []string{"check", "--max-cps", "１７", "a.srt"}, `srtvet: invalid --max-cps "１７"`},
		{"letters", []string{"check", "--max-cps", "abc", "a.srt"}, `srtvet: invalid --max-cps "abc"`},
		{"missing final value", []string{"check", "a.srt", "--max-cps"}, `srtvet: --max-cps requires a value`},
		{"option-like value is consumed", []string{"check", "--max-cps", "--max-line", "5", "a.srt"}, `srtvet: invalid --max-cps "--max-line"`},
		{"fix rejects max-cps", []string{"fix", "--max-cps", "5", "a.srt"}, `srtvet: unknown option "--max-cps"`},
		{"last repeated value is validated", []string{"check", "--max-cps", "17", "--max-cps", "101", "a.srt"}, `srtvet: invalid --max-cps "101"`},
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

func TestRunMaxCPSErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"above maximum", []string{"check", "--max-cps", "101", "a.srt"}, `srtvet: invalid --max-cps "101"`},
		{"empty equals value", []string{"check", "--max-cps=", "a.srt"}, `srtvet: invalid --max-cps ""`},
		{"empty separate value", []string{"check", "--max-cps", "", "a.srt"}, `srtvet: invalid --max-cps ""`},
		{"plus sign", []string{"check", "--max-cps", "+1", "a.srt"}, `srtvet: invalid --max-cps "+1"`},
		{"minus sign", []string{"check", "--max-cps", "-1", "a.srt"}, `srtvet: invalid --max-cps "-1"`},
		{"decimal", []string{"check", "--max-cps", "1.0", "a.srt"}, `srtvet: invalid --max-cps "1.0"`},
		{"whitespace", []string{"check", "--max-cps", "1 7", "a.srt"}, `srtvet: invalid --max-cps "1 7"`},
		{"four digits", []string{"check", "--max-cps", "0000", "a.srt"}, `srtvet: invalid --max-cps "0000"`},
		{"non-ASCII digits", []string{"check", "--max-cps", "１７", "a.srt"}, `srtvet: invalid --max-cps "１７"`},
		{"letters", []string{"check", "--max-cps", "abc", "a.srt"}, `srtvet: invalid --max-cps "abc"`},
		{"missing final value", []string{"check", "a.srt", "--max-cps"}, `srtvet: --max-cps requires a value`},
		{"option-like value is consumed", []string{"check", "--max-cps", "--max-line", "5", "a.srt"}, `srtvet: invalid --max-cps "--max-line"`},
		{"fix rejects max-cps", []string{"fix", "--max-cps", "5", "a.srt"}, `srtvet: unknown option "--max-cps"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantError(t, run(t, "", tc.args...), tc.want)
		})
	}
}

func TestRunMaxCPSValidationOrder(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"max-line before max-cps", []string{"check", "--max-line", "201", "--max-cps", "101"}, `srtvet: invalid --max-line "201"`},
		{"max-cps before file count", []string{"check", "--max-line", "200", "--max-cps", "101"}, `srtvet: invalid --max-cps "101"`},
		{"file count after valid options", []string{"check", "--max-line", "200", "--max-cps", "100"}, "srtvet: check needs at least one FILE"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantError(t, run(t, "", tc.args...), tc.want)
		})
	}
}

func TestRunMaxCPSSuppliedStdinExample(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		code   int
		stdout string
	}{
		{"limit 17", []string{"check", "--max-cps", "17", "-"}, 1,
			"-:1: SV007 reading speed 25.0 characters per second, the limit is 17\n" +
				"-:9: SV007 reading speed 17.5 characters per second, the limit is 17\n" +
				"-:13: SV007 reading speed 33.0 characters per second, the limit is 17\n" +
				"3 problems in 1 file\n"},
		{"limit 25 equals form", []string{"check", "--max-cps=25", "-"}, 1,
			"-:13: SV007 reading speed 33.0 characters per second, the limit is 25\n" +
				"1 problem in 1 file\n"},
		{"option omitted", []string{"check", "-"}, 0, "ok\n"},
		{"limit zero", []string{"check", "--max-cps", "0", "-"}, 0, "ok\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantOutput(t, run(t, fastSpeedSRT, tc.args...), tc.code, tc.stdout)
		})
	}
}

func TestRunMaxCPSFileInputAndOrder(t *testing.T) {
	dir := t.TempDir()
	second := writeFile(t, dir, "second.srt", "1\n00:00:01,000 --> 00:00:02,000\nABCD\n")
	first := writeFile(t, dir, "first.srt", "1\n00:00:01,000 --> 00:00:02,000\nAB\n")
	want := shown(second) + ":1: SV007 reading speed 4.0 characters per second, the limit is 1\n" +
		shown(first) + ":1: SV007 reading speed 2.0 characters per second, the limit is 1\n" +
		"2 problems in 2 files\n"
	wantOutput(t, run(t, "", "check", second, "--max-cps=1", first), 1, want)
}

func TestRunMaxCPSMaximumBoundaryAndZeroMaxLine(t *testing.T) {
	dir := t.TempDir()
	atLimit := writeFile(t, dir, "at.srt", "1\n00:00:01,000 --> 00:00:02,000\n"+strings.Repeat("x", 100)+"\n")
	aboveLimit := writeFile(t, dir, "above.srt", "1\n00:00:01,000 --> 00:00:02,000\n"+strings.Repeat("x", 101)+"\n")

	wantOutput(t, run(t, "", "check", "--max-line", "0", "--max-cps", "100", atLimit), 0, "ok\n")
	wantOutput(t, run(t, "", "check", "--max-line=0", "--max-cps=100", aboveLimit), 1,
		shown(aboveLimit)+":1: SV007 reading speed 101.0 characters per second, the limit is 100\n"+
			"1 problem in 1 file\n")
}
