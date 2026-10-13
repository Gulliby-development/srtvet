package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Gulliby-development/srtvet/internal/ops"
	"github.com/Gulliby-development/srtvet/internal/srt"
)

const usage = "usage: srtvet check [--max-line N] FILE...\n" +
	"       srtvet fix [--shift DURATION] [--stretch RATIO] [--sort] [--drop-empty] [--strip-tags] [--out FILE] FILE\n"

// Exit codes of Run.
const (
	exitOK       = 0
	exitFindings = 1
	exitError    = 2
)

// displayPath is the path as shown in messages: every backslash is
// replaced by a slash.
func displayPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// readCues reads and parses one FILE argument; the lone - is stdin.
func readCues(arg string, stdin io.Reader) ([]srt.Cue, error) {
	display := displayPath(arg)
	var data []byte
	var err error
	if arg == "-" {
		data, err = io.ReadAll(stdin)
	} else {
		data, err = os.ReadFile(arg)
	}
	if err != nil {
		return nil, fmt.Errorf("srtvet: cannot read \"%s\"", display)
	}
	return srt.Parse(display, data)
}

// count formats a number with its noun, plural unless n is 1.
func count(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

// runCheck writes the findings and the summary into out and returns the
// exit code. Every file is read and parsed before anything is reported.
func runCheck(opts options, stdin io.Reader, out *bytes.Buffer) (int, error) {
	parsed := make([][]srt.Cue, 0, len(opts.files))
	for _, file := range opts.files {
		cues, err := readCues(file, stdin)
		if err != nil {
			return exitError, err
		}
		parsed = append(parsed, cues)
	}
	problems, files := 0, 0
	for i, cues := range parsed {
		findings := ops.Check(cues, opts.maxLine)
		if len(findings) > 0 {
			files++
		}
		problems += len(findings)
		display := displayPath(opts.files[i])
		for _, f := range findings {
			fmt.Fprintf(out, "%s:%d: %s %s\n", display, f.Line, f.Code, f.Message)
		}
	}
	if problems == 0 {
		out.WriteString("ok\n")
		return exitOK, nil
	}
	fmt.Fprintf(out, "%s in %s\n", count(problems, "problem"), count(files, "file"))
	return exitFindings, nil
}

// runFix writes the fixed file into out, or into the --out file when one
// is given; then out stays empty.
func runFix(opts options, stdin io.Reader, out *bytes.Buffer) error {
	cues, err := readCues(opts.files[0], stdin)
	if err != nil {
		return err
	}
	if opts.stripTags {
		ops.StripTags(cues)
	}
	if opts.dropEmpty {
		cues = ops.DropEmpty(cues)
	}
	if opts.hasStretch {
		ops.Stretch(cues, opts.stretch)
	}
	if opts.hasShift {
		ops.Shift(cues, opts.shift)
	}
	if err := ops.CheckRange(cues); err != nil {
		return err
	}
	if opts.sort {
		ops.Sort(cues)
	}
	ops.Renumber(cues)
	var fixed bytes.Buffer
	if err := srt.Write(&fixed, cues); err != nil {
		return err
	}
	if !opts.hasOut {
		out.Write(fixed.Bytes())
		return nil
	}
	if err := os.WriteFile(opts.out, fixed.Bytes(), 0o644); err != nil {
		return fmt.Errorf("srtvet: cannot write \"%s\"", displayPath(opts.out))
	}
	return nil
}

// Run executes srtvet with the arguments after the program name and
// returns the exit code: 0 success, 1 findings, 2 any error. An error is
// one line on stderr and nothing on stdout.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	opts, err := parseArgs(args)
	var out bytes.Buffer
	code := exitOK
	switch {
	case err != nil:
	case opts.help:
		out.WriteString(usage)
	case opts.command == cmdCheck:
		code, err = runCheck(opts, stdin, &out)
	default:
		err = runFix(opts, stdin, &out)
	}
	if err != nil {
		io.WriteString(stderr, err.Error()+"\n")
		return exitError
	}
	if _, err := stdout.Write(out.Bytes()); err != nil {
		return exitError
	}
	return code
}
