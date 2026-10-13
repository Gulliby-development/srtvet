<p align="center"><img src="docs/logo.svg" width="96" alt="" /></p>

# srtvet

`srtvet` is an offline command-line checker and fixer for SubRip (`.srt`)
subtitle files. `srtvet check` reports broken cues, `srtvet fix` rewrites a
file into a clean one: it can move all timings, convert between frame rates
(for example `25/23.976`), sort the cues, drop empty ones and renumber them.

It is for people who download, edit or convert subtitles. It needs no network,
uses only the Go standard library and never changes its input file.

[![CI](https://github.com/Gulliby-development/srtvet/actions/workflows/ci.yml/badge.svg)](https://github.com/Gulliby-development/srtvet/actions/workflows/ci.yml) [![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

## Install

With a Go toolchain:

```
go install github.com/Gulliby-development/srtvet@latest
```

Or build the binary from a checkout of this repository:

```
go build
```

## Usage

```
usage: srtvet check [--max-line N] FILE...
       srtvet fix [--shift DURATION] [--stretch RATIO] [--sort] [--drop-empty] [--strip-tags] [--out FILE] FILE
```

`srtvet --help` (as the only argument) prints these two lines.

- The first argument is the command, `check` or `fix`.
- After the command the arguments are read left to right; options may stand
  among the FILEs.
- An option that takes a value accepts both `--opt VALUE` and `--opt=VALUE`.
  It takes the next argument whatever it looks like, so `--shift -0.5s` works.
- A repeated option keeps its last value.
- A FILE is any argument that does not start with `-`. The lone `-` means
  standard input.
- In messages a FILE path is shown as given, with every `\` replaced by `/`.

Examples:

```
srtvet check movie.srt
srtvet check --max-line 0 movie.srt other.srt
srtvet check - < movie.srt
srtvet fix --shift -0.5s clip.srt
srtvet fix --shift=1500ms --sort --out fixed.srt clip.srt
srtvet fix --stretch 25/23.976 clip.srt
```

The input is UTF-8 with an optional BOM and LF or CRLF line endings. Every
line the program prints ends with LF on every operating system.

## Checks

`srtvet check` reads and parses every FILE first, then prints one line per
finding as `FILE:LINE: CODE MESSAGE` and a summary line. LINE is the line of
the cue-number line, except for SV006, where it is the line of the long text
line. "Previous" is the cue before in the file.

| Code | Message | Reported when |
| --- | --- | --- |
| SV001 | `cue number N, expected M` | the declared number is not the previous cue's number plus 1 (1 for the first cue) |
| SV002 | `end is not after start` | the end is less than or equal to the start |
| SV003 | `starts at T1, before the previous cue ends at T2` | the start is not before the previous start but is before the previous end |
| SV004 | `starts at T1, before the previous cue starts at T2` | the start is before the previous start |
| SV005 | `no text` | the cue has no text lines |
| SV006 | `line K has C characters, the limit is L` | text line K of the cue has more than L Unicode code points |

A cue gets SV004 or SV003, never both. Tags such as `<i>` count like any other
text for SV006.

`--max-line N` sets the limit L of SV006: one to three digits, 0 to 200, the
default is 42, and 0 turns SV006 off.

The summary line is `N problems in M files` (M counts the files with at least
one finding), or the single line `ok` when there is no finding.

Worked example, `movie.srt`:

```
1
00:00:01,000 --> 00:00:03,500
Hello there.

2
00:00:03,000 --> 00:00:05,000
<i>Overlap with the first cue.</i>

4
00:00:06,000 --> 00:00:06,000

5
00:00:04,000 --> 00:00:04,500
Out of order, and this line is far too long to read comfortably on any screen!
Short.
```

`srtvet check movie.srt` prints this and exits with code 1:

```
movie.srt:5: SV003 starts at 00:00:03,000, before the previous cue ends at 00:00:03,500
movie.srt:9: SV001 cue number 4, expected 3
movie.srt:9: SV002 end is not after start
movie.srt:9: SV005 no text
movie.srt:12: SV004 starts at 00:00:04,000, before the previous cue starts at 00:00:06,000
movie.srt:14: SV006 line 1 has 78 characters, the limit is 42
6 problems in 1 file
```

`srtvet check --max-line 0 movie.srt` prints the same without the SV006 line
and ends with `5 problems in 1 file`.

## Fixing

`srtvet fix` reads exactly one FILE and writes the cleaned file to standard
output, or to the file named by `--out FILE`. The steps run in this order, each
only when its option is given, except the last, which always runs:

1. `--strip-tags` removes the formatting tags from the text (see below).
2. `--drop-empty` removes cues that have no text lines.
3. `--stretch RATIO` multiplies every start and end by the ratio, rounded half
   up to whole milliseconds.
4. `--shift DURATION` adds the duration to every start and end.
5. `--sort` sorts the cues by start, then by end; equal cues keep their file
   order.
6. The cues are renumbered 1, 2, 3, ... in their final order.

After stretching and shifting, a cue that would start below `00:00:00,000` or
end above `99:59:59,999` stops the run with
`srtvet: cue N: time out of range after retiming` (N is the cue's declared
number) and nothing is written.

`--shift DURATION` accepts:

- seconds: an optional `+` or `-`, one to five digits, optionally a dot and one
  to three digits, optionally followed by `s`: `+1.5s`, `-0.5s`, `2`, `0.040`,
  `-12s`;
- milliseconds: an optional `+` or `-`, one to eight digits and `ms`: `-250ms`,
  `1500ms`.

Forms such as `.5`, `1.5ms` or a value with a space are invalid.

`--stretch RATIO` accepts `X` or `X/Y`, where each of X and Y is one to four
digits, optionally a dot and one to six digits, greater than 0 and at most
1000: `1.001`, `25/23.976`, `24/25`. A plain `X` means `X/1`.

`--out FILE` writes the result to FILE instead of standard output. The file is
created or overwritten only after everything else succeeded, and it may name
the input file.

The output has no BOM and LF line endings: for each cue its number, the timing
line `HH:MM:SS,mmm --> HH:MM:SS,mmm` and its text lines exactly as read, with
one blank line between two cues. `fix` does not report the problems of `check`.

Worked example, `clip.srt`:

```
3
00:01:10,000 --> 00:01:12,040
Third, but listed first.

1
00:00:01,000 --> 00:00:03,500
<i>Hello there.</i>
Second line.

2
00:00:04,000 --> 00:00:04,500

```

`srtvet fix --stretch 25/23.976 --shift -0.5s --sort --drop-empty clip.srt`
prints this and exits with code 0 (the empty cue is gone, the times are
stretched and then moved by half a second, the cues are sorted and renumbered):

```
1
00:00:00,543 --> 00:00:03,149
<i>Hello there.</i>
Second line.

2
00:01:12,490 --> 00:01:14,617
Third, but listed first.
```

## Removing formatting tags

`--strip-tags` removes opening and closing `<i>`, `<b>`, `<u>`, `<s>` and `<font>` tags, without regard to case. Opening tags may have attributes after spaces or tabs. Override blocks such as `{\an8}` and `{\pos(10,20)\i1}` are removed.

Forms that stay include `<br>`, `<ii>`, `<fonts>`, `<i/>`, `</ i>`, `<hello>`, unclosed `<i`, `2 < 3`, `<3`, `{note}` and `{}`. Other text and spaces are kept as written. Text lines that become empty are removed; whitespace-only lines remain. `--drop-empty` removes any cue left without text.

Worked example, `tags.srt`:

```
1
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
```

`srtvet fix --strip-tags tags.srt` prints this and exits with code 0:

```
1
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
```

## Exit codes

| Code | Meaning |
| --- | --- |
| 0 | success: `check` printed `ok`, `fix` wrote its output, or `--help` |
| 1 | `check` reported at least one finding |
| 2 | any error (usage, invalid value, unreadable or unparsable file, time out of range, unwritable output): one line starting with `srtvet: ` on standard error and nothing on standard output |

## Development

The project uses the Go standard library only. Run the checks from the
repository root:

```
go vet ./...
go test ./...
```

Format the code with `go fmt ./...`.

## License

MIT, see [LICENSE](LICENSE).

---

Built with BAKED, an open-source coding agent (bench loop round 14).
