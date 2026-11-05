package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const minDurSRT = "1\n" +
	"00:00:01,000 --> 00:00:01,200\n" +
	"Hi\n" +
	"\n" +
	"2\n" +
	"00:00:01,800 --> 00:00:03,000\n" +
	"There\n" +
	"\n" +
	"3\n" +
	"00:00:05,000 --> 00:00:05,100\n" +
	"End\n"

func minDurOut(end1, end2, end3 string) string {
	return "1\n" +
		"00:00:01,000 --> " + end1 + "\n" +
		"Hi\n" +
		"\n" +
		"2\n" +
		"00:00:01,800 --> " + end2 + "\n" +
		"There\n" +
		"\n" +
		"3\n" +
		"00:00:05,000 --> " + end3 + "\n" +
		"End\n"
}

func TestFixMinDurationExample(t *testing.T) {
	path := writeFile(t, t.TempDir(), "short.srt", minDurSRT)
	want := minDurOut("00:00:01,800", "00:00:03,000", "00:00:06,000")
	wantOutput(t, run(t, "", "fix", "--min-duration", "1s", path), 0, want)
}

func TestFixMinDurationForms(t *testing.T) {
	path := writeFile(t, t.TempDir(), "short.srt", minDurSRT)
	tests := []struct {
		name  string
		args  []string
		want  string
		stdin string
	}{
		{"milliseconds", []string{"--min-duration", "500ms"}, minDurOut("00:00:01,500", "00:00:03,000", "00:00:05,500"), ""},
		{"fraction of a second", []string{"--min-duration", "0.5"}, minDurOut("00:00:01,500", "00:00:03,000", "00:00:05,500"), ""},
		{"seconds with a suffix", []string{"--min-duration", "1.5s"}, minDurOut("00:00:01,800", "00:00:03,300", "00:00:06,500"), ""},
		{"largest value", []string{"--min-duration", "60"}, minDurOut("00:00:01,800", "00:00:05,000", "00:01:05,000"), ""},
		{"equals form", []string{"--min-duration=500ms"}, minDurOut("00:00:01,500", "00:00:03,000", "00:00:05,500"), ""},
		{"last value wins", []string{"--min-duration", "60", "--min-duration=500ms"}, minDurOut("00:00:01,500", "00:00:03,000", "00:00:05,500"), ""},
		{"smallest value", []string{"--min-duration", "1ms"}, minDurSRT, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			args := append([]string{"fix"}, tc.args...)
			args = append(args, path)
			wantOutput(t, run(t, tc.stdin, args...), 0, tc.want)
		})
	}
}

func TestFixMinDurationRunsAfterShiftAndBeforeSort(t *testing.T) {
	dir := t.TempDir()
	path := writeFile(t, dir, "short.srt", minDurSRT)
	wantShifted := "1\n" +
		"00:00:02,000 --> 00:00:02,800\n" +
		"Hi\n" +
		"\n" +
		"2\n" +
		"00:00:02,800 --> 00:00:04,000\n" +
		"There\n" +
		"\n" +
		"3\n" +
		"00:00:06,000 --> 00:00:07,000\n" +
		"End\n"
	wantOutput(t, run(t, "", "fix", "--shift", "1s", "--min-duration", "1s", path), 0, wantShifted)

	unsorted := writeFile(t, dir, "unsorted.srt", "3\n00:00:09,000 --> 00:00:09,100\nLast\n\n"+
		"1\n00:00:01,000 --> 00:00:01,200\nHi\n\n"+
		"2\n00:00:01,800 --> 00:00:03,000\nThere\n")
	wantSorted := "1\n" +
		"00:00:01,000 --> 00:00:01,800\n" +
		"Hi\n" +
		"\n" +
		"2\n" +
		"00:00:01,800 --> 00:00:03,000\n" +
		"There\n" +
		"\n" +
		"3\n" +
		"00:00:09,000 --> 00:00:10,000\n" +
		"Last\n"
	wantOutput(t, run(t, "", "fix", "--min-duration", "1s", "--sort", unsorted), 0, wantSorted)
}

func TestFixMinDurationOutAndStdin(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.srt")
	wantOutput(t, run(t, minDurSRT, "fix", "--min-duration", "500ms", "--out", out, "-"), 0, "")
	if got, want := readFile(t, out), minDurOut("00:00:01,500", "00:00:03,000", "00:00:05,500"); got != want {
		t.Errorf("--out file\n got %q\nwant %q", got, want)
	}
}

func TestFixMinDurationRangeError(t *testing.T) {
	path := writeFile(t, t.TempDir(), "late.srt", "1\n99:59:59,500 --> 99:59:59,600\nx\n")
	wantError(t, run(t, "", "fix", "--min-duration", "1s", path), "srtvet: cue 1: time out of range after retiming")
	end := writeFile(t, t.TempDir(), "end.srt", "1\n99:59:59,000 --> 99:59:59,100\nx\n")
	wantOutput(t, run(t, "", "fix", "--min-duration", "900ms", end), 0, "1\n99:59:59,000 --> 99:59:59,900\nx\n")
}

func TestMinDurationOptionErrors(t *testing.T) {
	path := writeFile(t, t.TempDir(), "short.srt", minDurSRT)
	for _, value := range []string{"0", "0s", "0ms", "-1s", "+1s", "61", "60001ms", "60.001", "1.5ms", ".5", "x", "", "1 s", "1e3"} {
		t.Run("value "+value, func(t *testing.T) {
			wantError(t, run(t, "", "fix", "--min-duration", value, path), `srtvet: invalid --min-duration "`+value+`"`)
		})
	}
	wantError(t, run(t, "", "fix", path, "--min-duration"), "srtvet: --min-duration requires a value")
	wantError(t, run(t, "", "check", "--min-duration", "1s", path), `srtvet: unknown option "--min-duration"`)
}

func TestMinDurationReadmeDocumentation(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("cannot read README.md: %v", err)
	}
	readme := strings.ReplaceAll(string(data), "\r\n", "\n")
	for _, want := range []string{
		"## Minimum duration\n",
		"`--min-duration DURATION`",
		"5. `--min-duration DURATION` lengthens the cues that last less than the duration",
		"6. `--sort` sorts the cues by start, then by end",
		"7. The cues are renumbered 1, 2, 3, ... in their final order.",
	} {
		if !strings.Contains(readme, want) {
			t.Errorf("README.md does not contain %q", want)
		}
	}
	if strings.Index(readme, "## Fixing\n") > strings.Index(readme, "## Minimum duration\n") ||
		strings.Index(readme, "## Minimum duration\n") > strings.Index(readme, "## Removing formatting tags\n") {
		t.Error("the section Minimum duration is not between Fixing and Removing formatting tags")
	}
}
