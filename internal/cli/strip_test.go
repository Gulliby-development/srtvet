package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const tagsSRT = `1
00:00:01,000 --> 00:00:02,000
{\an8}<i>Hello</i> <font color="#ff0000">there</font>.

2
00:00:03,000 --> 00:00:04,000
<b></b>
<U>Two</U> lines, 2 < 3 and <3 {note}

3
00:00:05,000 --> 00:00:06,000
<i>
{\i1}
</i>

4
00:00:07,000 --> 00:00:08,000
Plain <hello> <br> <s >x</S>
`

const tagsStrippedSRT = `1
00:00:01,000 --> 00:00:02,000
Hello there.

2
00:00:03,000 --> 00:00:04,000
Two lines, 2 < 3 and <3 {note}

3
00:00:05,000 --> 00:00:06,000

4
00:00:07,000 --> 00:00:08,000
Plain <hello> <br> x
`

const tagsDropEmptySortedSRT = `1
00:00:01,000 --> 00:00:02,000
Hello there.

2
00:00:03,000 --> 00:00:04,000
Two lines, 2 < 3 and <3 {note}

3
00:00:07,000 --> 00:00:08,000
Plain <hello> <br> x
`

func TestFixStripTagsExamples(t *testing.T) {
	path := writeFile(t, t.TempDir(), "tags.srt", tagsSRT)

	t.Run("worked output", func(t *testing.T) {
		wantOutput(t, run(t, "", "fix", "--strip-tags", path), 0, tagsStrippedSRT)
	})

	t.Run("drop empty and sort", func(t *testing.T) {
		wantOutput(t, run(t, "", "fix", "--drop-empty", "--strip-tags", "--sort", path), 0, tagsDropEmptySortedSRT)
	})
}

func TestFixStripTagsOptionPlacementAndRepetition(t *testing.T) {
	path := writeFile(t, t.TempDir(), "tags.srt", tagsSRT)
	tests := []struct {
		name string
		args []string
	}{
		{"before file", []string{"fix", "--strip-tags", path}},
		{"after file", []string{"fix", path, "--strip-tags"}},
		{"between options and file", []string{"fix", "--sort", "--strip-tags", path}},
		{"repeated before and after file", []string{"fix", "--strip-tags", path, "--strip-tags"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantOutput(t, run(t, "", tc.args...), 0, tagsStrippedSRT)
		})
	}
}

func TestFixWithoutStripTagsPreservesText(t *testing.T) {
	path := writeFile(t, t.TempDir(), "tags.srt", tagsSRT)
	wantOutput(t, run(t, "", "fix", path), 0, tagsSRT)
}

func TestFixStripTagsFromStdin(t *testing.T) {
	wantOutput(t, run(t, tagsSRT, "fix", "--strip-tags", "-"), 0, tagsStrippedSRT)
}

func TestFixStripTagsWithOtherTransforms(t *testing.T) {
	input := `1
00:00:03,000 --> 00:00:04,000
Third <i>later</i>

2
00:00:01,000 --> 00:00:02,000
{\an8}<b></b>

3
00:00:02,000 --> 00:00:03,000
First <u>early</u>
`
	want := `1
00:00:03,500 --> 00:00:05,500
First early

2
00:00:05,500 --> 00:00:07,500
Third later
`
	path := writeFile(t, t.TempDir(), "ordered.srt", input)
	args := []string{"fix", "--drop-empty", "--stretch", "2", "--shift", "-0.5s", "--sort", "--strip-tags", path}
	wantOutput(t, run(t, "", args...), 0, want)
}

func TestFixStripTagsOutputFileLeavesInputUnchanged(t *testing.T) {
	dir := t.TempDir()
	inputPath := writeFile(t, dir, "tags.srt", tagsSRT)
	outputPath := filepath.Join(dir, "stripped.srt")

	res := run(t, "", "fix", "--strip-tags", "--out", outputPath, inputPath)
	wantOutput(t, res, 0, "")
	if got := readFile(t, inputPath); got != tagsSRT {
		t.Errorf("input file changed\n got %q\nwant %q", got, tagsSRT)
	}
	if got := readFile(t, outputPath); got != tagsStrippedSRT {
		t.Errorf("output file\n got %q\nwant %q", got, tagsStrippedSRT)
	}
	if strings.Contains(readFile(t, outputPath), "\r") {
		t.Error("output file contains a CR byte; want LF line endings")
	}
}

func TestFixStripTagsErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"equals form is unknown", []string{"fix", "--strip-tags=1", "a.srt"}, `srtvet: unknown option "--strip-tags=1"`},
		{"not available to check", []string{"check", "--strip-tags", "a.srt"}, `srtvet: unknown option "--strip-tags"`},
		{"fix still requires a file", []string{"fix", "--strip-tags"}, "srtvet: fix needs exactly one FILE"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantError(t, run(t, "", tc.args...), tc.want)
		})
	}
}

func TestStripTagsReadmeDocumentation(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("cannot read README.md: %v", err)
	}
	readme := strings.ReplaceAll(string(data), "\r\n", "\n")

	const usageLine = "       srtvet fix [--shift DURATION] [--stretch RATIO] [--sort] [--drop-empty] [--strip-tags] [--out FILE] FILE"
	if !strings.Contains(readme, usageLine) {
		t.Errorf("README.md does not contain the fix usage line %q", usageLine)
	}

	fixingStart := strings.Index(readme, "## Fixing\n")
	tagsStart := strings.Index(readme, "## Removing formatting tags\n")
	exitStart := strings.Index(readme, "## Exit codes\n")
	developmentStart := strings.Index(readme, "## Development\n")
	if fixingStart < 0 || tagsStart < 0 || exitStart < 0 || developmentStart < 0 ||
		!(fixingStart < tagsStart && tagsStart < exitStart && exitStart < developmentStart) {
		t.Errorf("README section order is wrong: Fixing=%d, Removing formatting tags=%d, Exit codes=%d, Development=%d",
			fixingStart, tagsStart, exitStart, developmentStart)
	}

	const fixSteps = "1. `--strip-tags` removes the formatting tags from the text (see below).\n" +
		"2. `--drop-empty` removes cues that have no text lines.\n" +
		"3. `--stretch RATIO` multiplies every start and end by the ratio, rounded half\n" +
		"   up to whole milliseconds.\n" +
		"4. `--shift DURATION` adds the duration to every start and end.\n" +
		"5. `--sort` sorts the cues by start, then by end; equal cues keep their file\n" +
		"   order.\n" +
		"6. The cues are renumbered 1, 2, 3, ... in their final order."
	if fixingStart >= 0 && tagsStart > fixingStart && !strings.Contains(readme[fixingStart:tagsStart], fixSteps) {
		t.Error("README.md does not document strip-tags as fix step 1 followed by steps 2 through 6")
	}

	const removedForms = "`--strip-tags` removes opening and closing `<i>`, `<b>`, `<u>`, `<s>` and `<font>` tags, without regard to case."
	const openingAttributes = "Opening tags may have attributes after spaces or tabs."
	const overrideForms = "Override blocks such as `{\\an8}` and `{\\pos(10,20)\\i1}` are removed."
	const retainedForms = "Forms that stay include `<br>`, `<ii>`, `<fonts>`, `<i/>`, `</ i>`, `<hello>`, unclosed `<i`, `2 < 3`, `<3`, `{note}` and `{}`."
	const keptText = "Other text and spaces are kept as written. Text lines that become empty are removed; whitespace-only lines remain. `--drop-empty` removes any cue left without text."
	for _, want := range []string{removedForms, openingAttributes, overrideForms, retainedForms, keptText} {
		if !strings.Contains(readme, want) {
			t.Errorf("README.md does not contain required tag documentation %q", want)
		}
	}

	const inputExample = "Worked example, `tags.srt`:\n\n```\n" + tagsSRT + "```"
	const outputExample = "`srtvet fix --strip-tags tags.srt` prints this and exits with code 0:\n\n```\n" + tagsStrippedSRT + "```"
	for _, example := range []string{inputExample, outputExample} {
		if !strings.Contains(readme, example) {
			t.Errorf("README.md does not contain the exact tags.srt example %q", example)
		}
	}
}
