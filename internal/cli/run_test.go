package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const firstCueSRT = "1\n" +
	"00:00:01,000 --> 00:00:03,500\n" +
	"Hello there.\n"

const movieSRT = firstCueSRT +
	"\n" +
	"2\n" +
	"00:00:03,000 --> 00:00:05,000\n" +
	"<i>Overlap with the first cue.</i>\n" +
	"\n" +
	"4\n" +
	"00:00:06,000 --> 00:00:06,000\n" +
	"\n" +
	"5\n" +
	"00:00:04,000 --> 00:00:04,500\n" +
	"Out of order, and this line is far too long to read comfortably on any screen!\n" +
	"Short.\n"

const clipSRT = "3\n" +
	"00:01:10,000 --> 00:01:12,040\n" +
	"Third, but listed first.\n" +
	"\n" +
	"1\n" +
	"00:00:01,000 --> 00:00:03,500\n" +
	"<i>Hello there.</i>\n" +
	"Second line.\n" +
	"\n" +
	"2\n" +
	"00:00:04,000 --> 00:00:04,500\n" +
	"\n"

// clipFixed is the output of
// fix --stretch 25/23.976 --shift -0.5s --sort --drop-empty clip.srt.
const clipFixed = "1\n" +
	"00:00:00,543 --> 00:00:03,149\n" +
	"<i>Hello there.</i>\n" +
	"Second line.\n" +
	"\n" +
	"2\n" +
	"00:01:12,490 --> 00:01:14,617\n" +
	"Third, but listed first.\n"

// oneFindingSRT has exactly one finding: SV001 on line 1.
const oneFindingSRT = "2\n" +
	"00:00:01,000 --> 00:00:02,000\n" +
	"hi\n"

const usageText = "usage: srtvet check [--max-line N] FILE...\n" +
	"       srtvet fix [--shift DURATION] [--stretch RATIO] [--sort] [--drop-empty] [--out FILE] FILE\n"

var clipFixArgs = []string{"fix", "--stretch", "25/23.976", "--shift", "-0.5s", "--sort", "--drop-empty"}

type result struct {
	code   int
	stdout string
	stderr string
}

// run calls Run with injected streams and checks what holds for every
// call: no CR byte in either output and a final LF on a non-empty output.
func run(t *testing.T, stdin string, args ...string) result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, strings.NewReader(stdin), &stdout, &stderr)
	res := result{code: code, stdout: stdout.String(), stderr: stderr.String()}
	streams := []struct{ name, text string }{{"stdout", res.stdout}, {"stderr", res.stderr}}
	for _, s := range streams {
		if strings.Contains(s.text, "\r") {
			t.Errorf("Run(%q) %s contains a CR byte: %q", args, s.name, s.text)
		}
		if s.text != "" && !strings.HasSuffix(s.text, "\n") {
			t.Errorf("Run(%q) %s does not end with LF: %q", args, s.name, s.text)
		}
	}
	return res
}

// wantError checks the shape of every error: exit code 2, an empty stdout
// and exactly one stderr line, which starts with "srtvet: ".
func wantError(t *testing.T, res result, line string) {
	t.Helper()
	if res.code != 2 {
		t.Errorf("exit code = %d, want 2", res.code)
	}
	if res.stdout != "" {
		t.Errorf("stdout = %q, want empty", res.stdout)
	}
	if !strings.HasPrefix(res.stderr, "srtvet: ") {
		t.Errorf("stderr = %q, want the prefix %q", res.stderr, "srtvet: ")
	}
	if n := strings.Count(res.stderr, "\n"); n != 1 {
		t.Errorf("stderr has %d lines, want 1: %q", n, res.stderr)
	}
	if res.stderr != line+"\n" {
		t.Errorf("stderr\n got %q\nwant %q", res.stderr, line+"\n")
	}
}

// wantOutput checks a run without an error: the exit code, the exact
// stdout and an empty stderr.
func wantOutput(t *testing.T, res result, code int, stdout string) {
	t.Helper()
	if res.code != code {
		t.Errorf("exit code = %d, want %d", res.code, code)
	}
	if res.stdout != stdout {
		t.Errorf("stdout\n got %q\nwant %q", res.stdout, stdout)
	}
	if res.stderr != "" {
		t.Errorf("stderr = %q, want empty", res.stderr)
	}
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

// shown is the path as the program shows it in messages.
func shown(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// movieFindings is the worked check output of movie.srt without the
// summary line, for the display path p.
func movieFindings(p string, withSV006 bool) string {
	out := p + ":5: SV003 starts at 00:00:03,000, before the previous cue ends at 00:00:03,500\n" +
		p + ":9: SV001 cue number 4, expected 3\n" +
		p + ":9: SV002 end is not after start\n" +
		p + ":9: SV005 no text\n" +
		p + ":12: SV004 starts at 00:00:04,000, before the previous cue starts at 00:00:06,000\n"
	if withSV006 {
		out += p + ":14: SV006 line 1 has 78 characters, the limit is 42\n"
	}
	return out
}

func TestCheckMovie(t *testing.T) {
	path := writeFile(t, t.TempDir(), "movie.srt", movieSRT)
	res := run(t, "", "check", path)
	wantOutput(t, res, 1, movieFindings(shown(path), true)+"6 problems in 1 file\n")
}

func TestCheckMaxLine(t *testing.T) {
	dir := t.TempDir()
	movie := writeFile(t, dir, "movie.srt", movieSRT)
	p := shown(movie)
	without := movieFindings(p, false) + "5 problems in 1 file\n"
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"limit 0 turns the rule off", []string{"check", "--max-line", "0", movie}, without},
		{"limit 80", []string{"check", "--max-line", "80", movie}, without},
		{"limit 0, equals form", []string{"check", "--max-line=0", movie}, without},
		{"option after the file", []string{"check", movie, "--max-line", "0"}, without},
		{"limit equal to the line length", []string{"check", "--max-line", "78", movie}, without},
		{"limit one below the line length", []string{"check", "--max-line", "77", movie},
			movieFindings(p, false) + p + ":14: SV006 line 1 has 78 characters, the limit is 77\n" + "6 problems in 1 file\n"},
		{"repeated option keeps the last value", []string{"check", "--max-line", "0", "--max-line", "42", movie},
			movieFindings(p, true) + "6 problems in 1 file\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantOutput(t, run(t, "", tc.args...), 1, tc.want)
		})
	}
}

func TestCheckOK(t *testing.T) {
	dir := t.TempDir()
	one := writeFile(t, dir, "one.srt", firstCueSRT)
	empty := writeFile(t, dir, "empty.srt", "")
	tests := []struct {
		name string
		args []string
	}{
		{"one cue", []string{"check", one}},
		{"empty file", []string{"check", empty}},
		{"two clean files", []string{"check", one, empty}},
		{"largest limit", []string{"check", "--max-line", "200", one}},
		{"limit with a leading zero", []string{"check", "--max-line", "042", one}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantOutput(t, run(t, "", tc.args...), 0, "ok\n")
		})
	}
}

func TestCheckDefaultLimit(t *testing.T) {
	dir := t.TempDir()
	head := "1\n00:00:01,000 --> 00:00:02,000\n"
	atLimit := writeFile(t, dir, "at.srt", head+strings.Repeat("\u00e9", 42)+"\n")
	over := writeFile(t, dir, "over.srt", head+strings.Repeat("\u00e9", 43)+"\n")

	wantOutput(t, run(t, "", "check", atLimit), 0, "ok\n")
	wantOutput(t, run(t, "", "check", over), 1,
		shown(over)+":3: SV006 line 1 has 43 characters, the limit is 42\n"+"1 problem in 1 file\n")
}

func TestCheckSeveralFiles(t *testing.T) {
	dir := t.TempDir()
	movie := writeFile(t, dir, "movie.srt", movieSRT)
	second := writeFile(t, dir, "second.srt", oneFindingSRT)
	clean := writeFile(t, dir, "clean.srt", firstCueSRT)
	secondFinding := shown(second) + ":1: SV001 cue number 2, expected 1\n"
	combined := movieFindings(shown(movie), true) + secondFinding + "7 problems in 2 files\n"
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"two files", []string{"check", movie, second}, combined},
		{"three files, the clean one is not counted", []string{"check", movie, clean, second}, combined},
		{"argument order decides", []string{"check", second, movie},
			secondFinding + movieFindings(shown(movie), true) + "7 problems in 2 files\n"},
		{"the second file alone", []string{"check", second}, secondFinding + "1 problem in 1 file\n"},
		{"the same file twice", []string{"check", second, second}, secondFinding + secondFinding + "2 problems in 2 files\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantOutput(t, run(t, "", tc.args...), 1, tc.want)
		})
	}
}

func TestHelp(t *testing.T) {
	wantOutput(t, run(t, "", "--help"), 0, usageText)
}

func TestCommandErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"no argument", nil, `srtvet: missing command (check or fix)`},
		{"unknown command", []string{"frobnicate"}, `srtvet: unknown command "frobnicate"`},
		{"command with a dash", []string{"-x"}, `srtvet: unknown command "-x"`},
		{"help with another argument", []string{"--help", "check"}, `srtvet: unknown command "--help"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantError(t, run(t, "", tc.args...), tc.want)
		})
	}
}

func TestFixShiftForms(t *testing.T) {
	path := writeFile(t, t.TempDir(), "clip.srt", clipSRT)
	plus := "1\n" +
		"00:01:11,000 --> 00:01:13,040\n" +
		"Third, but listed first.\n" +
		"\n" +
		"2\n" +
		"00:00:02,000 --> 00:00:04,500\n" +
		"<i>Hello there.</i>\n" +
		"Second line.\n" +
		"\n" +
		"3\n" +
		"00:00:05,000 --> 00:00:05,500\n"
	minus := "1\n" +
		"00:01:09,000 --> 00:01:11,040\n" +
		"Third, but listed first.\n" +
		"\n" +
		"2\n" +
		"00:00:00,000 --> 00:00:02,500\n" +
		"<i>Hello there.</i>\n" +
		"Second line.\n" +
		"\n" +
		"3\n" +
		"00:00:03,000 --> 00:00:03,500\n"
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"separate value", []string{"fix", "--shift", "1s", path}, plus},
		{"equals form", []string{"fix", "--shift=1s", path}, plus},
		{"option after the file", []string{"fix", path, "--shift", "1s"}, plus},
		{"value that starts with a dash", []string{"fix", "--shift", "-1s", path}, minus},
		{"repeated option keeps the last value", []string{"fix", "--shift", "5s", "--shift", "1s", path}, plus},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantOutput(t, run(t, "", tc.args...), 0, tc.want)
		})
	}
}

func TestOptionErrors(t *testing.T) {
	f := writeFile(t, t.TempDir(), "f.srt", firstCueSRT)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"unknown option after a FILE", []string{"check", f, "--bogus"}, `srtvet: unknown option "--bogus"`},
		{"help after a command", []string{"check", f, "--help"}, `srtvet: unknown option "--help"`},
		{"double dash", []string{"fix", f, "--"}, `srtvet: unknown option "--"`},
		{"short option", []string{"check", "-x", f}, `srtvet: unknown option "-x"`},
		{"shift without a value", []string{"fix", f, "--shift"}, `srtvet: --shift requires a value`},
		{"stretch without a value", []string{"fix", f, "--stretch"}, `srtvet: --stretch requires a value`},
		{"out without a value", []string{"fix", f, "--out"}, `srtvet: --out requires a value`},
		{"max-line without a value", []string{"check", f, "--max-line"}, `srtvet: --max-line requires a value`},
		{"fix option for check", []string{"check", "--shift", "1s", f}, `srtvet: unknown option "--shift"`},
		{"sort with a value", []string{"fix", "--sort=x", f}, `srtvet: unknown option "--sort=x"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantError(t, run(t, "", tc.args...), tc.want)
		})
	}
}

func TestValueAndCountErrors(t *testing.T) {
	dir := t.TempDir()
	f := writeFile(t, dir, "f.srt", firstCueSRT)
	g := writeFile(t, dir, "g.srt", firstCueSRT)
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"check without a FILE", []string{"check"}, `srtvet: check needs at least one FILE`},
		{"max-line before the FILE count", []string{"check", "--max-line", "201"}, `srtvet: invalid --max-line "201"`},
		{"max-line letters", []string{"check", "--max-line", "abc", f}, `srtvet: invalid --max-line "abc"`},
		{"max-line four digits", []string{"check", "--max-line=1000", f}, `srtvet: invalid --max-line "1000"`},
		{"max-line negative", []string{"check", "--max-line", "-1", f}, `srtvet: invalid --max-line "-1"`},
		{"max-line empty", []string{"check", "--max-line=", f}, `srtvet: invalid --max-line ""`},
		{"fix without a FILE", []string{"fix"}, `srtvet: fix needs exactly one FILE`},
		{"fix with two FILEs", []string{"fix", f, g}, `srtvet: fix needs exactly one FILE`},
		{"shift before the FILE count", []string{"fix", "--shift", ".5"}, `srtvet: invalid --shift ".5"`},
		{"shift fraction with ms", []string{"fix", "--shift", "1.5ms", f}, `srtvet: invalid --shift "1.5ms"`},
		{"shift before stretch", []string{"fix", "--shift", "x", "--stretch", "0"}, `srtvet: invalid --shift "x"`},
		{"stretch before the FILE count", []string{"fix", "--shift", "1s", "--stretch", "0"}, `srtvet: invalid --stretch "0"`},
		{"stretch above 1000", []string{"fix", "--stretch", "1000.000001", f}, `srtvet: invalid --stretch "1000.000001"`},
		{"stretch zero", []string{"fix", "--stretch", "0.000000", f}, `srtvet: invalid --stretch "0.000000"`},
		{"shift empty", []string{"fix", "--shift=", f}, `srtvet: invalid --shift ""`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantError(t, run(t, "", tc.args...), tc.want)
		})
	}

	t.Run("stretch 1000 is accepted", func(t *testing.T) {
		want := "1\n" +
			"00:16:40,000 --> 00:58:20,000\n" +
			"Hello there.\n"
		wantOutput(t, run(t, "", "fix", "--stretch", "1000", f), 0, want)
	})
}

func TestInputErrors(t *testing.T) {
	dir := t.TempDir()
	good := writeFile(t, dir, "good.srt", firstCueSRT)
	movie := writeFile(t, dir, "movie.srt", movieSRT)
	bad := writeFile(t, dir, "bad.srt", "x")
	missing := filepath.Join(dir, "missing.srt")
	backslash := dir + `\no-such-dir\missing.srt`
	parseLine := `srtvet: ` + shown(bad) + `:1: expected a cue number, found "x"`
	readLine := `srtvet: cannot read "` + shown(missing) + `"`
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"parse failure after a good file", []string{"check", good, bad}, parseLine},
		{"parse failure after a file with findings", []string{"check", movie, bad}, parseLine},
		{"parse failure before a good file", []string{"check", bad, good}, parseLine},
		{"fix on a file that does not parse", []string{"fix", bad}, parseLine},
		{"check on a missing file", []string{"check", missing}, readLine},
		{"fix on a missing file", []string{"fix", missing}, readLine},
		{"missing file after a good one", []string{"check", good, missing}, readLine},
		{"missing file after a file with findings", []string{"check", movie, missing}, readLine},
		{"first failing file wins", []string{"check", missing, bad}, readLine},
		{"backslash in the argument, check", []string{"check", backslash},
			`srtvet: cannot read "` + strings.ReplaceAll(backslash, "\\", "/") + `"`},
		{"backslash in the argument, fix", []string{"fix", backslash},
			`srtvet: cannot read "` + strings.ReplaceAll(backslash, "\\", "/") + `"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := run(t, "", tc.args...)
			wantError(t, res, tc.want)
			if strings.Contains(res.stderr, "\\") {
				t.Errorf("stderr contains a backslash: %q", res.stderr)
			}
		})
	}
}

func TestParseErrorLines(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"not a cue number", "x", `:1: expected a cue number, found "x"`},
		{"cue number at the end", "1", `:1: cue number without a timing line`},
		{"short arrow", "1\n00:00:01,000 -> 00:00:02,000\n", `:2: expected a timing line, found "00:00:01,000 -> 00:00:02,000"`},
		{"second cue is bad", firstCueSRT + "\n\n2\nbad\n", `:7: expected a timing line, found "bad"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bad := writeFile(t, dir, "bad.srt", tc.content)
			want := "srtvet: " + shown(bad) + tc.want
			wantError(t, run(t, "", "check", bad), want)
			wantError(t, run(t, "", "fix", bad), want)
		})
	}
}

func TestStdin(t *testing.T) {
	t.Run("check", func(t *testing.T) {
		res := run(t, movieSRT, "check", "-")
		wantOutput(t, res, 1, movieFindings("-", true)+"6 problems in 1 file\n")
	})
	t.Run("fix", func(t *testing.T) {
		args := append(append([]string{}, clipFixArgs...), "-")
		wantOutput(t, run(t, clipSRT, args...), 0, clipFixed)
	})
	t.Run("parse error", func(t *testing.T) {
		wantError(t, run(t, "x", "check", "-"), `srtvet: -:1: expected a cue number, found "x"`)
	})
}

func TestFixClip(t *testing.T) {
	path := writeFile(t, t.TempDir(), "clip.srt", clipSRT)
	args := append(append([]string{}, clipFixArgs...), path)
	wantOutput(t, run(t, "", args...), 0, clipFixed)
}

func TestFixShiftSortAndRange(t *testing.T) {
	path := writeFile(t, t.TempDir(), "clip.srt", clipSRT)

	want := "1\n" +
		"00:00:02,500 --> 00:00:05,000\n" +
		"<i>Hello there.</i>\n" +
		"Second line.\n" +
		"\n" +
		"2\n" +
		"00:00:05,500 --> 00:00:06,000\n" +
		"\n" +
		"3\n" +
		"00:01:11,500 --> 00:01:13,540\n" +
		"Third, but listed first.\n"
	wantOutput(t, run(t, "", "fix", "--shift", "1.5", "--sort", path), 0, want)

	wantError(t, run(t, "", "fix", "--shift", "-2s", path), `srtvet: cue 1: time out of range after retiming`)

	// The cue declared 1 starts at 1000 ms: a shift of -1s puts it at 0,
	// which is in range, and one millisecond more is not.
	if res := run(t, "", "fix", "--shift", "-1000ms", path); res.code != 0 || res.stderr != "" {
		t.Errorf("shift -1000ms: code %d, stderr %q, want 0 and empty", res.code, res.stderr)
	}
	wantError(t, run(t, "", "fix", "--shift", "-1001ms", path), `srtvet: cue 1: time out of range after retiming`)
}

func TestFixOut(t *testing.T) {
	t.Run("writes the file and nothing on stdout", func(t *testing.T) {
		dir := t.TempDir()
		in := writeFile(t, dir, "clip.srt", clipSRT)
		out := filepath.Join(dir, "out.srt")
		args := append(append([]string{}, clipFixArgs...), "--out", out, in)
		wantOutput(t, run(t, "", args...), 0, "")
		if got := readFile(t, out); got != clipFixed {
			t.Errorf("out file\n got %q\nwant %q", got, clipFixed)
		}
		if got := readFile(t, in); got != clipSRT {
			t.Errorf("input file changed: %q", got)
		}
	})

	t.Run("overwrites an existing file", func(t *testing.T) {
		dir := t.TempDir()
		in := writeFile(t, dir, "clip.srt", clipSRT)
		out := writeFile(t, dir, "out.srt", strings.Repeat("old content\n", 100))
		args := append(append([]string{}, clipFixArgs...), "--out="+out, in)
		wantOutput(t, run(t, "", args...), 0, "")
		if got := readFile(t, out); got != clipFixed {
			t.Errorf("out file\n got %q\nwant %q", got, clipFixed)
		}
	})

	t.Run("out names the input file", func(t *testing.T) {
		in := writeFile(t, t.TempDir(), "clip.srt", clipSRT)
		args := append(append([]string{}, clipFixArgs...), "--out", in, in)
		wantOutput(t, run(t, "", args...), 0, "")
		if got := readFile(t, in); got != clipFixed {
			t.Errorf("file\n got %q\nwant %q", got, clipFixed)
		}
	})

	t.Run("out cannot be written", func(t *testing.T) {
		dir := t.TempDir()
		in := writeFile(t, dir, "clip.srt", clipSRT)
		out := filepath.Join(dir, "no-such-dir", "out.srt")
		res := run(t, "", "fix", "--out", out, in)
		wantError(t, res, `srtvet: cannot write "`+shown(out)+`"`)
		if got := readFile(t, in); got != clipSRT {
			t.Errorf("input file changed: %q", got)
		}
	})

	t.Run("range error creates no file", func(t *testing.T) {
		dir := t.TempDir()
		in := writeFile(t, dir, "clip.srt", clipSRT)
		out := filepath.Join(dir, "out2.srt")
		res := run(t, "", "fix", "--shift", "-2s", "--out", out, in)
		wantError(t, res, `srtvet: cue 1: time out of range after retiming`)
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Errorf("out file exists after a range error (stat error %v)", err)
		}
	})

	t.Run("range error keeps an input named by out", func(t *testing.T) {
		in := writeFile(t, t.TempDir(), "clip.srt", clipSRT)
		res := run(t, "", "fix", "--shift", "-2s", "--out", in, in)
		wantError(t, res, `srtvet: cue 1: time out of range after retiming`)
		if got := readFile(t, in); got != clipSRT {
			t.Errorf("input file changed: %q", got)
		}
	})
}

func TestInputIsNeverChanged(t *testing.T) {
	dir := t.TempDir()
	movie := writeFile(t, dir, "movie.srt", movieSRT)
	clip := writeFile(t, dir, "clip.srt", clipSRT)

	if res := run(t, "", "check", movie, clip); res.code != 1 {
		t.Errorf("check exit code = %d, want 1", res.code)
	}
	if res := run(t, "", "fix", "--sort", "--drop-empty", "--shift", "1s", clip); res.code != 0 {
		t.Errorf("fix exit code = %d, want 0", res.code)
	}
	if res := run(t, "", "fix", movie); res.code != 0 {
		t.Errorf("fix exit code = %d, want 0", res.code)
	}

	if got := readFile(t, movie); got != movieSRT {
		t.Errorf("movie.srt changed: %q", got)
	}
	if got := readFile(t, clip); got != clipSRT {
		t.Errorf("clip.srt changed: %q", got)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	if got := strings.Join(names, " "); got != "clip.srt movie.srt" {
		t.Errorf("directory holds %q, want only clip.srt and movie.srt", got)
	}
}

func TestFixFormats(t *testing.T) {
	dir := t.TempDir()

	t.Run("BOM and CRLF", func(t *testing.T) {
		path := writeFile(t, dir, "bom.srt", "\xEF\xBB\xBF1\r\n00:00:01.000 --> 00:00:02.000  \r\nhi\r\n")
		wantOutput(t, run(t, "", "fix", path), 0, "1\n00:00:01,000 --> 00:00:02,000\nhi\n")
		wantOutput(t, run(t, "", "check", path), 0, "ok\n")
	})

	t.Run("empty file", func(t *testing.T) {
		path := writeFile(t, dir, "empty.srt", "")
		wantOutput(t, run(t, "", "fix", path), 0, "")
	})

	t.Run("only blank lines", func(t *testing.T) {
		path := writeFile(t, dir, "blank.srt", "\n \t\n\n")
		wantOutput(t, run(t, "", "fix", "--sort", "--drop-empty", path), 0, "")
	})

	t.Run("movie keeps the zero-length cue in file order", func(t *testing.T) {
		path := writeFile(t, dir, "movie.srt", movieSRT)
		want := firstCueSRT +
			"\n" +
			"2\n" +
			"00:00:03,000 --> 00:00:05,000\n" +
			"<i>Overlap with the first cue.</i>\n" +
			"\n" +
			"3\n" +
			"00:00:06,000 --> 00:00:06,000\n" +
			"\n" +
			"4\n" +
			"00:00:04,000 --> 00:00:04,500\n" +
			"Out of order, and this line is far too long to read comfortably on any screen!\n" +
			"Short.\n"
		wantOutput(t, run(t, "", "fix", path), 0, want)
	})

	t.Run("drop-empty alone keeps the file order", func(t *testing.T) {
		path := writeFile(t, dir, "clip.srt", clipSRT)
		want := "1\n" +
			"00:01:10,000 --> 00:01:12,040\n" +
			"Third, but listed first.\n" +
			"\n" +
			"2\n" +
			"00:00:01,000 --> 00:00:03,500\n" +
			"<i>Hello there.</i>\n" +
			"Second line.\n"
		wantOutput(t, run(t, "", "fix", "--drop-empty", path), 0, want)
	})
}
