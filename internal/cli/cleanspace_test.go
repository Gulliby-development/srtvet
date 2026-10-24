package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const cleanSpaceExampleInput = "1\n" +
	"00:00:01,000 --> 00:00:02,000\n" +
	"\ufeffHello\u00a0\u00a0there,\n" +
	"you \u200bknow\u3000\n" +
	"\n" +
	"2\n" +
	"00:00:03,000 --> 00:00:04,000\n" +
	"\u00a0\n" +
	"\n" +
	"3\n" +
	"00:00:05,000 --> 00:00:06,000\n" +
	"a\u2028b\n" +
	"x \u00a0 y\n" +
	"z  w\n" +
	"p\t\u00a0q\n"

const cleanSpaceExampleOutput = "1\n" +
	"00:00:01,000 --> 00:00:02,000\n" +
	"Hello there,\n" +
	"you know\n" +
	"\n" +
	"2\n" +
	"00:00:03,000 --> 00:00:04,000\n" +
	"\n" +
	"3\n" +
	"00:00:05,000 --> 00:00:06,000\n" +
	"a b\n" +
	"x y\n" +
	"z  w\n" +
	"p\t q\n"

func TestFixCleanSpaceExampleFromFileAndStdin(t *testing.T) {
	path := writeFile(t, t.TempDir(), "odd.srt", cleanSpaceExampleInput)

	t.Run("file", func(t *testing.T) {
		wantOutput(t, run(t, "", "fix", "--clean-space", path), 0, cleanSpaceExampleOutput)
	})
	t.Run("standard input", func(t *testing.T) {
		wantOutput(t, run(t, cleanSpaceExampleInput, "fix", "--clean-space", "-"), 0, cleanSpaceExampleOutput)
	})
}

func TestFixCleanSpaceFlagPlacementRepetitionAndOutput(t *testing.T) {
	dir := t.TempDir()
	inputPath := writeFile(t, dir, "odd.srt", cleanSpaceExampleInput)
	tests := []struct {
		name string
		args []string
	}{
		{"before file", []string{"fix", "--clean-space", inputPath}},
		{"after file", []string{"fix", inputPath, "--clean-space"}},
		{"between existing options", []string{"fix", "--sort", "--clean-space", inputPath}},
		{"repeated before and after file", []string{"fix", "--clean-space", inputPath, "--clean-space"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wantOutput(t, run(t, "", tc.args...), 0, cleanSpaceExampleOutput)
		})
	}

	t.Run("out file", func(t *testing.T) {
		outputPath := filepath.Join(dir, "cleaned.srt")
		wantOutput(t, run(t, "", "fix", "--clean-space", "--out", outputPath, inputPath), 0, "")
		if got := readFile(t, outputPath); got != cleanSpaceExampleOutput {
			t.Errorf("output file\n got %q\nwant %q", got, cleanSpaceExampleOutput)
		}
	})
	t.Run("omitting the option preserves text", func(t *testing.T) {
		wantOutput(t, run(t, "", "fix", inputPath), 0, cleanSpaceExampleInput)
	})
}

func TestFixCleanSpaceRunsBeforeStripTags(t *testing.T) {
	input := "1\n00:00:01,000 --> 00:00:02,000\n<b>\u00a0x\u00a0</b>\n"
	want := "1\n00:00:01,000 --> 00:00:02,000\n x \n"
	path := writeFile(t, t.TempDir(), "tags.srt", input)

	for _, args := range [][]string{
		{"fix", "--clean-space", "--strip-tags", path},
		{"fix", "--strip-tags", "--clean-space", path},
	} {
		wantOutput(t, run(t, "", args...), 0, want)
	}
}

func TestFixCleanSpaceRetainsEmptyCuesUnlessDropped(t *testing.T) {
	input := "1\n00:00:01,000 --> 00:00:02,000\n\u00a0\n\n" +
		"2\n00:00:03,000 --> 00:00:04,000\n\n" +
		"3\n00:00:05,000 --> 00:00:06,000\nstay\n"
	path := writeFile(t, t.TempDir(), "empty-lines.srt", input)
	withoutDrop := "1\n00:00:01,000 --> 00:00:02,000\n\n" +
		"2\n00:00:03,000 --> 00:00:04,000\n\n" +
		"3\n00:00:05,000 --> 00:00:06,000\nstay\n"
	wantOutput(t, run(t, "", "fix", "--clean-space", path), 0, withoutDrop)
}

func TestFixCleanSpaceDropEmptyAndOtherTransforms(t *testing.T) {
	input := "1\n00:00:05,000 --> 00:00:06,000\nlater\n\n" +
		"2\n00:00:01,000 --> 00:00:02,000\n\u00a0\n\n" +
		"3\n00:00:03,000 --> 00:00:04,000\n\n" +
		"4\n00:00:01,000 --> 00:00:02,000\na\u00a0b\n"
	want := "1\n00:00:03,000 --> 00:00:05,000\na b\n\n" +
		"2\n00:00:11,000 --> 00:00:13,000\nlater\n"
	path := writeFile(t, t.TempDir(), "drop-empty.srt", input)

	wantOutput(t, run(t, "", "fix", "--clean-space", "--drop-empty", "--stretch", "2", "--shift", "1s", "--sort", path), 0, want)
}

func TestCleanSpaceOptionErrorsAndHelp(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"equals form for fix", []string{"fix", "--clean-space=1", "input.srt"}, "srtvet: unknown option \"--clean-space=1\""},
		{"option for check", []string{"check", "--clean-space", "input.srt"}, "srtvet: unknown option \"--clean-space\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wantError(t, run(t, "", tc.args...), tc.want)
		})
	}

	wantOutput(t, run(t, "", "--help"), 0, usageText)
}

func TestCleanSpaceReadmeDocumentation(t *testing.T) {
	root := cleanSpaceProjectRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "README.md"))
	if err != nil {
		t.Fatalf("cannot read README.md: %v", err)
	}
	readme := strings.ReplaceAll(string(data), "\r\n", "\n")

	fence := strings.Repeat(string(rune(96)), 3)
	usageFence := fence + "\n" +
		"usage: srtvet check [--max-line N] [--max-cps N] FILE...\n" +
		"       srtvet fix [--shift DURATION] [--stretch RATIO] [--sort] [--drop-empty] [--strip-tags] [--clean-space] [--out FILE] FILE\n" +
		fence
	usageStart := strings.Index(readme, "## Usage\n")
	if usageStart < 0 || !strings.Contains(readme[usageStart:], usageFence) {
		t.Error("README.md Usage fence does not contain the exact updated usage lines")
	}

	const fixSteps = "1. ~--strip-tags~ removes the formatting tags from the text (see below).\n" +
		"2. ~--drop-empty~ removes cues that have no text lines.\n" +
		"3. ~--stretch RATIO~ multiplies every start and end by the ratio, rounded half\n" +
		"   up to whole milliseconds.\n" +
		"4. ~--shift DURATION~ adds the duration to every start and end.\n" +
		"5. ~--sort~ sorts the cues by start, then by end; equal cues keep their file\n" +
		"   order.\n" +
		"6. The cues are renumbered 1, 2, 3, ... in their final order."
	const cleanSpaceNote = "~--clean-space~ is not a numbered step: it runs first, before step 1 (see below)."
	markdownTick := string(rune(96))
	if !strings.Contains(readme, strings.ReplaceAll(fixSteps, "~", markdownTick)+"\n\n"+strings.ReplaceAll(cleanSpaceNote, "~", markdownTick)+"\n\n") {
		t.Error("README.md does not keep the six fix steps followed by the exact clean-space paragraph")
	}

	expectedSectionTemplate := "## Cleaning spaces\n\n" +
		"~--clean-space~ cleans the text lines of every cue (not the timing lines) before any other step of ~fix~. It first deletes the invisible characters U+00AD, U+200B, U+2060 and U+FEFF. A *space run* is a maximal sequence of plain spaces (U+0020) and odd spaces; the odd spaces are U+00A0, U+2000 to U+200A, U+202F, U+205F, U+3000, U+0085, U+2028, U+2029 and the carriage return (U+000D). A run that holds at least one odd space becomes one plain space, or nothing when it touches the start or the end of the line; a run of plain spaces only stays as it is. Every other character, a TAB included, ends a run and stays. Text lines that become empty are removed; ~--drop-empty~ removes any cue left without text.\n\n" +
		"Worked example, ~odd.srt~ (each ~<U+XXXX>~ stands for that one character, ~<TAB>~ for a TAB):\n\n" +
		fence + "\n" +
		"1\n00:00:01,000 --> 00:00:02,000\n<U+FEFF>Hello<U+00A0><U+00A0>there,\nyou <U+200B>know<U+3000>\n\n" +
		"2\n00:00:03,000 --> 00:00:04,000\n<U+00A0>\n\n" +
		"3\n00:00:05,000 --> 00:00:06,000\na<U+2028>b\nx <U+00A0> y\nz  w\np<TAB><U+00A0>q\n" +
		fence + "\n\n" +
		"~srtvet fix --clean-space odd.srt~ prints this and exits with code 0:\n\n" +
		fence + "\n" +
		"1\n00:00:01,000 --> 00:00:02,000\nHello there,\nyou know\n\n" +
		"2\n00:00:03,000 --> 00:00:04,000\n\n" +
		"3\n00:00:05,000 --> 00:00:06,000\na b\nx y\nz  w\np<TAB> q\n" +
		fence + "\n\n"
	expectedSection := strings.ReplaceAll(expectedSectionTemplate, "~", markdownTick)
	cleanStart := strings.Index(readme, "## Cleaning spaces\n")
	tagsStart := strings.Index(readme, "## Removing formatting tags\n")
	exitStart := strings.Index(readme, "## Exit codes\n")
	if cleanStart < 0 || tagsStart < 0 || exitStart < 0 || !(tagsStart < cleanStart && cleanStart < exitStart) {
		t.Fatalf("README section order is wrong: Removing formatting tags=%d, Cleaning spaces=%d, Exit codes=%d", tagsStart, cleanStart, exitStart)
	}
	if got := readme[cleanStart:exitStart]; got != expectedSection {
		index := 0
		for index < len(got) && index < len(expectedSection) && got[index] == expectedSection[index] {
			index++
		}
		t.Errorf("README Cleaning spaces section differs at byte %d (got length %d, want length %d): got byte %q, want byte %q", index, len(got), len(expectedSection), got[index:min(index+1, len(got))], expectedSection[index:min(index+1, len(expectedSection))])
	}
}

func cleanSpaceProjectRoot(t *testing.T) string {
	t.Helper()
	_, sourceFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate the test source file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(sourceFile), "..", ".."))
}

func TestCleanSpaceProcessPreservesExactBytes(t *testing.T) {
	root := cleanSpaceProjectRoot(t)
	binary := buildCleanSpaceBinary(t, root)

	const cueHeader = "1\n00:00:01,000 --> 00:00:02,000\n"
	makeCue := func(text []byte) []byte {
		input := append([]byte(cueHeader), text...)
		return append(input, '\n')
	}
	makeOutput := func(text []byte) []byte {
		return makeCue(text)
	}

	for _, tc := range []struct {
		name string
		text []byte
		want []byte
	}{
		{"NBSP runs around invalid byte", []byte{97, 194, 160, 255, 194, 160, 98}, []byte{97, 32, 255, 32, 98}},
		{"boundary NBSP runs around invalid byte", []byte{194, 160, 255, 194, 160}, []byte{255}},
		{"invalid overlong bytes are preserved", []byte{97, 192, 175, 98}, []byte{97, 192, 175, 98}},
		{"valid replacement character is preserved", []byte{97, 239, 191, 189, 98}, []byte{97, 239, 191, 189, 98}},
	} {
		t.Run("stdin/"+tc.name, func(t *testing.T) {
			code, stdout, stderr := runCleanSpaceBinary(t, binary, makeCue(tc.text), "fix", "--clean-space", "-")
			if code != 0 || len(stderr) != 0 || !bytes.Equal(stdout, makeOutput(tc.want)) {
				t.Errorf("process result = (code %d, stdout %q, stderr %q), want (code 0, stdout %q, empty stderr)",
					code, stdout, stderr, makeOutput(tc.want))
			}
		})
	}

	t.Run("cue-text BOM through stdin", func(t *testing.T) {
		input := makeCue([]byte{0xef, 0xbb, 0xbf, 'H', 'e', 'l', 'l', 'o', 0xc2, 0xa0, 'x'})
		want := makeOutput([]byte("Hello x"))
		code, stdout, stderr := runCleanSpaceBinary(t, binary, input, "fix", "--clean-space", "-")
		if code != 0 || len(stderr) != 0 || !bytes.Equal(stdout, want) {
			t.Errorf("process result = (code %d, stdout %q, stderr %q), want (code 0, stdout %q, empty stderr)",
				code, stdout, stderr, want)
		}
	})

	t.Run("cue-text BOM and invalid UTF-8 through file", func(t *testing.T) {
		input := makeCue([]byte{0xef, 0xbb, 0xbf, 'a', 0xc2, 0xa0, 0xff, 0xc2, 0xa0, 'b'})
		path := filepath.Join(t.TempDir(), "raw.srt")
		if err := os.WriteFile(path, input, 0o644); err != nil {
			t.Fatalf("write raw subtitle file: %v", err)
		}
		want := makeOutput([]byte{'a', ' ', 0xff, ' ', 'b'})
		code, stdout, stderr := runCleanSpaceBinary(t, binary, nil, "fix", "--clean-space", path)
		if code != 0 || len(stderr) != 0 || !bytes.Equal(stdout, want) {
			t.Errorf("process result = (code %d, stdout %q, stderr %q), want (code 0, stdout %q, empty stderr)",
				code, stdout, stderr, want)
		}
	})
}

func buildCleanSpaceBinary(t *testing.T, root string) string {
	t.Helper()
	name := "srtvet"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	binary := filepath.Join(t.TempDir(), name)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", binary, ".")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			t.Fatalf("building srtvet timed out: %v", ctx.Err())
		}
		t.Fatalf("building srtvet: %v\n%s", err, output)
	}
	return binary
}

func runCleanSpaceBinary(t *testing.T, binary string, stdin []byte, args ...string) (int, []byte, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return 0, stdout.Bytes(), stderr.Bytes()
	}
	if ctx.Err() != nil {
		t.Fatalf("running srtvet timed out: %v", ctx.Err())
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode(), stdout.Bytes(), stderr.Bytes()
	}
	t.Fatalf("running srtvet: %v", err)
	return -1, stdout.Bytes(), stderr.Bytes()
}
