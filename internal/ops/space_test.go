package ops

// BEGIN acceptance rows imports: cases from the spec's acceptance table. Edit acceptanceActual, not this block.
import (
	acceptancejson "encoding/json"
	acceptancereflect "reflect"
	acceptancetesting "testing"
)

// END acceptance rows imports

import (
	"fmt"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

// BEGIN acceptance rows: cases from the spec's acceptance table. Edit acceptanceActual, not this block.
func TestAcceptanceRowsSpace(t *acceptancetesting.T) {
	var rows []struct {
		ID       string `json:"id"`
		Input    any    `json:"input"`
		Expected any    `json:"expected"`
	}
	if err := acceptancejson.Unmarshal([]byte("[{\"id\":\"AC-001:1\",\"input\":\"Hello world\",\"expected\":\"Hello world\"},{\"id\":\"AC-001:2\",\"input\":\"a\\u00a0b\",\"expected\":\"a b\"},{\"id\":\"AC-001:3\",\"input\":\"a\\u00a0\\u00a0b\",\"expected\":\"a b\"},{\"id\":\"AC-001:4\",\"input\":\"a \\u00a0b\",\"expected\":\"a b\"},{\"id\":\"AC-001:5\",\"input\":\"a\\u00a0 b\",\"expected\":\"a b\"},{\"id\":\"AC-001:6\",\"input\":\"a  b\",\"expected\":\"a  b\"},{\"id\":\"AC-001:7\",\"input\":\"\\u00a0a\",\"expected\":\"a\"},{\"id\":\"AC-001:8\",\"input\":\"a\\u00a0\",\"expected\":\"a\"},{\"id\":\"AC-001:9\",\"input\":\"a \",\"expected\":\"a \"},{\"id\":\"AC-001:10\",\"input\":\"a \\u00a0\",\"expected\":\"a\"},{\"id\":\"AC-001:11\",\"input\":\" \\u00a0a\",\"expected\":\"a\"},{\"id\":\"AC-001:12\",\"input\":\"\\u00a0\\u3000\",\"expected\":\"\"},{\"id\":\"AC-001:13\",\"input\":\"   \",\"expected\":\"   \"},{\"id\":\"AC-001:14\",\"input\":\"\",\"expected\":\"\"},{\"id\":\"AC-001:15\",\"input\":\"a\\u200bb\",\"expected\":\"ab\"},{\"id\":\"AC-001:16\",\"input\":\"a \\u200bb\",\"expected\":\"a b\"},{\"id\":\"AC-001:17\",\"input\":\"a \\u200b b\",\"expected\":\"a  b\"},{\"id\":\"AC-001:18\",\"input\":\"a\\u00a0\\u200b\\u00a0b\",\"expected\":\"a b\"},{\"id\":\"AC-001:19\",\"input\":\"\\ufeff\\u200b\\u00a0 x\",\"expected\":\"x\"},{\"id\":\"AC-001:20\",\"input\":\"\\ufeffHello\",\"expected\":\"Hello\"},{\"id\":\"AC-001:21\",\"input\":\"Hello\\ufeff\",\"expected\":\"Hello\"},{\"id\":\"AC-001:22\",\"input\":\"co\\u00adoperate\",\"expected\":\"cooperate\"},{\"id\":\"AC-001:23\",\"input\":\"a\\u2060b\",\"expected\":\"ab\"},{\"id\":\"AC-001:24\",\"input\":\"a\\u0085b\",\"expected\":\"a b\"},{\"id\":\"AC-001:25\",\"input\":\"a\\u2028b\",\"expected\":\"a b\"},{\"id\":\"AC-001:26\",\"input\":\"a\\u2029b\",\"expected\":\"a b\"},{\"id\":\"AC-001:27\",\"input\":\"a\\rb\",\"expected\":\"a b\"},{\"id\":\"AC-001:28\",\"input\":\"a\\r\",\"expected\":\"a\"},{\"id\":\"AC-001:29\",\"input\":\"a\\u3000b\",\"expected\":\"a b\"},{\"id\":\"AC-001:30\",\"input\":\"a\\u2000b\\u2009c\\u200ad\",\"expected\":\"a b c d\"},{\"id\":\"AC-001:31\",\"input\":\"a\\u202fb\\u205fc\",\"expected\":\"a b c\"},{\"id\":\"AC-001:32\",\"input\":\"a\\u200cb\",\"expected\":\"a\\u200cb\"},{\"id\":\"AC-001:33\",\"input\":\"👨\\u200d👩\",\"expected\":\"👨\\u200d👩\"},{\"id\":\"AC-001:34\",\"input\":\"a\\u200eb\\u202eb\",\"expected\":\"a\\u200eb\\u202eb\"},{\"id\":\"AC-001:35\",\"input\":\"a‰b⁞b、b\",\"expected\":\"a‰b⁞b、b\"},{\"id\":\"AC-001:36\",\"input\":\"¡\\u009f\\u0086¬®\",\"expected\":\"¡\\u009f\\u0086¬®\"},{\"id\":\"AC-001:37\",\"input\":\"x\\u180e y\\u1680z\",\"expected\":\"x\\u180e y\\u1680z\"},{\"id\":\"AC-001:38\",\"input\":\"a\\tb\",\"expected\":\"a\\tb\"},{\"id\":\"AC-001:39\",\"input\":\"a\\t\\u00a0b\",\"expected\":\"a\\t b\"},{\"id\":\"AC-002:1\",\"input\":[97,194,160,255,194,160,98],\"expected\":[97,32,255,32,98]},{\"id\":\"AC-002:2\",\"input\":[194,160,255,194,160],\"expected\":[255]},{\"id\":\"AC-002:3\",\"input\":[97,192,175,98],\"expected\":[97,192,175,98]},{\"id\":\"AC-002:4\",\"input\":[97,239,191,189,98],\"expected\":[97,239,191,189,98]}]"), &rows); err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		row := row
		t.Run(row.ID, func(t *acceptancetesting.T) {
			got, err := acceptanceActual(row.Input)
			if err != nil {
				t.Fatal(err)
			}
			text, err := acceptancejson.Marshal(got)
			if err != nil {
				t.Fatal(err)
			}
			var value any
			if err := acceptancejson.Unmarshal(text, &value); err != nil {
				t.Fatal(err)
			}
			if !acceptancereflect.DeepEqual(value, row.Expected) {
				want, _ := acceptancejson.Marshal(row.Expected)
				t.Fatalf("got %s, want %s", text, want)
			}
		})
	}
}

// END acceptance rows

func acceptanceActual(input any) (any, error) {
	switch value := input.(type) {
	case string:
		return CleanSpace(value), nil
	case []any:
		bytes := make([]byte, len(value))
		for i, item := range value {
			number, ok := item.(float64)
			if !ok || number < 0 || number > 255 || number != float64(byte(number)) {
				return nil, fmt.Errorf("byte at index %d is not an integer from 0 through 255: %v", i, item)
			}
			bytes[i] = byte(number)
		}
		cleaned := []byte(CleanSpace(string(bytes)))
		actual := make([]int, len(cleaned))
		for i, value := range cleaned {
			actual[i] = int(value)
		}
		return actual, nil
	default:
		return nil, fmt.Errorf("unsupported acceptance input type %T", input)
	}
}

func TestCleanSpaceOddSpaceMembersAndBoundaries(t *acceptancetesting.T) {
	oddSpaces := []rune{
		'\u00a0', '\u2000', '\u2001', '\u2002', '\u2003', '\u2004',
		'\u2005', '\u2006', '\u2007', '\u2008', '\u2009', '\u200a',
		'\u202f', '\u205f', '\u3000', '\u0085', '\u2028', '\u2029', '\r',
	}
	for _, odd := range oddSpaces {
		odd := odd
		t.Run(fmt.Sprintf("U+%04X", odd), func(t *acceptancetesting.T) {
			if got := CleanSpace(string(odd) + "x"); got != "x" {
				t.Errorf("leading odd space = %q, want %q", got, "x")
			}
			if got := CleanSpace("x" + string(odd) + "y"); got != "x y" {
				t.Errorf("interior odd space = %q, want %q", got, "x y")
			}
			if got := CleanSpace("x" + string(odd)); got != "x" {
				t.Errorf("trailing odd space = %q, want %q", got, "x")
			}
		})
	}
}

func TestCleanSpacePreservesNonmembersAndRunBoundaries(t *acceptancetesting.T) {
	preserved := []rune{
		'\t', '\u200c', '\u200d', '\u200e', '\u1680', '\u180e',
		'\u0084', '\u0086', '\u00a9', '\u1fff', '\u202e', '\u205e', '\u2fff', '\u3001',
	}
	for _, character := range preserved {
		character := character
		t.Run(fmt.Sprintf("U+%04X", character), func(t *acceptancetesting.T) {
			input := "left\u00a0" + string(character) + "\u00a0right"
			want := "left " + string(character) + " right"
			if got := CleanSpace(input); got != want {
				t.Errorf("CleanSpace(%q) = %q, want %q", input, got, want)
			}
		})
	}

	for _, input := range []string{"a   b", "   ", "a ", " a"} {
		if got := CleanSpace(input); got != input {
			t.Errorf("plain-space-only run changed: CleanSpace(%q) = %q", input, got)
		}
	}
}

func TestCleanSpaceDeletesBeforeEvaluatingRuns(t *acceptancetesting.T) {
	deleted := []rune{'\u00ad', '\u200b', '\u2060', '\ufeff'}
	for _, character := range deleted {
		character := character
		t.Run(fmt.Sprintf("U+%04X", character), func(t *acceptancetesting.T) {
			if got := CleanSpace(string(character)); got != "" {
				t.Errorf("deleted character = %q, want empty", got)
			}
			if got := CleanSpace("a\u00a0" + string(character) + " \u00a0b"); got != "a b" {
				t.Errorf("deleted character did not join adjacent run: got %q, want %q", got, "a b")
			}
			if got := CleanSpace(string(character) + "\u00a0x"); got != "x" {
				t.Errorf("deleted character at leading boundary = %q, want %q", got, "x")
			}
			if got := CleanSpace("x\u00a0" + string(character)); got != "x" {
				t.Errorf("deleted character at trailing boundary = %q, want %q", got, "x")
			}
		})
	}
}

func TestCleanSpaceLinesFiltersOnlyEmptyLinesAndPreservesCues(t *acceptancetesting.T) {
	cues := []srt.Cue{
		cue(7, 1234, 5678, 11, 13, "\u00a0", "   ", "", "a\u00a0b", "\u200b"),
		cue(8, 2222, 3333, 17, 19, "\ufeff"),
		cue(9, 4444, 5555, 23, 0),
	}
	want := []srt.Cue{
		cue(7, 1234, 5678, 11, 13),
		cue(8, 2222, 3333, 17, 19),
		cue(9, 4444, 5555, 23, 0),
	}

	CleanSpaceLines(cues)

	if len(cues) != len(want) {
		t.Fatalf("cue count = %d, want %d", len(cues), len(want))
	}
	assertTextLines(t, cues[0].Text, []string{"   ", "a b"})
	assertTextLines(t, cues[1].Text, []string{})
	assertTextLines(t, cues[2].Text, []string{})
	for i := range cues {
		gotMetadata, wantMetadata := cues[i], want[i]
		gotMetadata.Text = nil
		wantMetadata.Text = nil
		if !acceptancereflect.DeepEqual(gotMetadata, wantMetadata) {
			t.Errorf("cue %d metadata = %+v, want %+v", i, gotMetadata, wantMetadata)
		}
	}
}

func TestCleanSpaceLinesAcceptsNilAndEmptyCueCollections(t *acceptancetesting.T) {
	for _, cues := range [][]srt.Cue{nil, {}} {
		CleanSpaceLines(cues)
		if len(cues) != 0 {
			t.Errorf("cue count after cleaning = %d, want 0", len(cues))
		}
	}
}
