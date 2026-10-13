// Package cli holds the command-line front end of srtvet: the argument
// parser and Run.
package cli

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Gulliby-development/srtvet/internal/ops"
)

const (
	cmdCheck = "check"
	cmdFix   = "fix"

	optMaxLine   = "--max-line"
	optShift     = "--shift"
	optStretch   = "--stretch"
	optOut       = "--out"
	optSort      = "--sort"
	optDropEmpty = "--drop-empty"
	optStripTags = "--strip-tags"

	// defaultMaxLine is the line-length limit without --max-line.
	defaultMaxLine = 42
	// maxMaxLine is the largest accepted --max-line value.
	maxMaxLine = 200
)

// valueOptions lists, per command, the options that take a value.
var valueOptions = map[string]map[string]bool{
	cmdCheck: {optMaxLine: true},
	cmdFix:   {optShift: true, optStretch: true, optOut: true},
}

var maxLineRE = regexp.MustCompile(`^[0-9]{1,3}$`)

// options is the parsed command line.
type options struct {
	command    string
	help       bool
	files      []string
	maxLine    int
	shift      int64
	hasShift   bool
	stretch    ops.Ratio
	hasStretch bool
	sort       bool
	dropEmpty  bool
	stripTags  bool
	out        string
	hasOut     bool
}

func invalidValue(opt, value string) error {
	return fmt.Errorf("srtvet: invalid %s \"%s\"", opt, value)
}

// parseArgs reads the arguments after the program name. Its errors are
// the complete lines to print.
func parseArgs(args []string) (options, error) {
	if len(args) == 0 {
		return options{}, fmt.Errorf("srtvet: missing command (check or fix)")
	}
	if len(args) == 1 && args[0] == "--help" {
		return options{help: true, maxLine: defaultMaxLine}, nil
	}
	cmd := args[0]
	if cmd != cmdCheck && cmd != cmdFix {
		return options{}, fmt.Errorf("srtvet: unknown command \"%s\"", cmd)
	}
	opts := options{command: cmd, maxLine: defaultMaxLine}
	// raw holds the last value given for each value option.
	raw := map[string]string{}
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if arg == "-" || !strings.HasPrefix(arg, "-") {
			opts.files = append(opts.files, arg)
			continue
		}
		name, value, hasValue := strings.Cut(arg, "=")
		switch {
		case valueOptions[cmd][name]:
			if !hasValue {
				if i+1 >= len(args) {
					return options{}, fmt.Errorf("srtvet: %s requires a value", name)
				}
				i++
				value = args[i]
			}
			raw[name] = value
		case cmd == cmdFix && arg == optSort:
			opts.sort = true
		case cmd == cmdFix && arg == optDropEmpty:
			opts.dropEmpty = true
		case cmd == cmdFix && arg == optStripTags:
			opts.stripTags = true
		default:
			return options{}, fmt.Errorf("srtvet: unknown option \"%s\"", arg)
		}
	}

	if cmd == cmdCheck {
		if v, given := raw[optMaxLine]; given {
			n, err := strconv.Atoi(v)
			if !maxLineRE.MatchString(v) || err != nil || n > maxMaxLine {
				return options{}, invalidValue(optMaxLine, v)
			}
			opts.maxLine = n
		}
		if len(opts.files) < 1 {
			return options{}, fmt.Errorf("srtvet: check needs at least one FILE")
		}
		return opts, nil
	}

	if v, given := raw[optShift]; given {
		ms, ok := ops.ParseDuration(v)
		if !ok {
			return options{}, invalidValue(optShift, v)
		}
		opts.shift, opts.hasShift = ms, true
	}
	if v, given := raw[optStretch]; given {
		r, ok := ops.ParseRatio(v)
		if !ok {
			return options{}, invalidValue(optStretch, v)
		}
		opts.stretch, opts.hasStretch = r, true
	}
	opts.out, opts.hasOut = raw[optOut]
	if len(opts.files) != 1 {
		return options{}, fmt.Errorf("srtvet: fix needs exactly one FILE")
	}
	return opts, nil
}
