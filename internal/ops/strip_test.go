package ops

import (
	"testing"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

func TestStripTagsFormatting(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"lowercase i", "<i>italic</i>", "italic"},
		{"uppercase i", "<I>italic</I>", "italic"},
		{"lowercase b", "<b>bold</b>", "bold"},
		{"uppercase b", "<B>bold</B>", "bold"},
		{"lowercase u", "<u>underlined</u>", "underlined"},
		{"uppercase u", "<U>underlined</U>", "underlined"},
		{"lowercase s", "<s>struck</s>", "struck"},
		{"uppercase s", "<S>struck</S>", "struck"},
		{"lowercase font", "<font>red</font>", "red"},
		{"uppercase font", "<FONT>red</FONT>", "red"},
		{"font attributes separated by spaces", "<font color=\"#ff0000\" size=3>red</font>", "red"},
		{"font attributes separated by a tab", "<font\tcolor=\"blue\">blue</FONT>", "blue"},
		{"alignment override", "{\\an8}top", "top"},
		{"position and style override", "left{\\pos(10,20)\\i1}right", "leftright"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := []srt.Cue{cue(1, 1000, 2000, 1, 3, tc.in)}
			StripTags(cues)
			assertTextLines(t, cues[0].Text, []string{tc.want})
		})
	}
}

func TestStripTagsPreservesNearMatches(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{"br", "before<br>after"},
		{"ii", "before<ii>after"},
		{"fonts", "before<fonts>after"},
		{"space after closing slash", "before</ i>after"},
		{"self-closing tag", "before<i/>after"},
		{"other name", "before<hello>after"},
		{"unclosed i", "before<i"},
		{"unclosed i with attributes", "before<i hello"},
		{"less-than comparison", "2 < 3"},
		{"less-than followed by a digit", "<3"},
		{"brace note", "{note}"},
		{"empty braces", "{}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := []srt.Cue{cue(1, 1000, 2000, 1, 3, tc.in)}
			StripTags(cues)
			assertTextLines(t, cues[0].Text, []string{tc.in})
		})
	}
}

func TestStripTagsPreservesWhitespace(t *testing.T) {
	cues := []srt.Cue{cue(1, 1000, 2000, 1, 3, "\t <i>word</i> \t", "<b> </b>")}
	StripTags(cues)

	assertTextLines(t, cues[0].Text, []string{"\t word \t", " "})
}

func TestStripTagsDropsEmptyLinesAndPreservesCueMetadata(t *testing.T) {
	cues := []srt.Cue{
		cue(7, 1234, 5678, 11, 13, "<i></i>", "{\\an8}", "  ", "", "<b>text</b>"),
		cue(8, 2222, 3333, 17, 19, "<i>", "{\\pos(10,20)\\i1}", "</i>"),
		cue(9, 4444, 5555, 23, 0),
	}
	StripTags(cues)

	assertTextLines(t, cues[0].Text, []string{"  ", "text"})
	assertTextLines(t, cues[1].Text, []string{})
	assertTextLines(t, cues[2].Text, []string{})

	wantMetadata := []srt.Cue{
		cue(7, 1234, 5678, 11, 13),
		cue(8, 2222, 3333, 17, 19),
		cue(9, 4444, 5555, 23, 0),
	}
	for i, got := range cues {
		want := wantMetadata[i]
		if got.Number != want.Number || got.Start != want.Start || got.End != want.End || got.Line != want.Line || got.TextLine != want.TextLine {
			t.Errorf("cue %d metadata = number %d, times %d-%d, line %d, text line %d; want number %d, times %d-%d, line %d, text line %d",
				i, got.Number, got.Start, got.End, got.Line, got.TextLine,
				want.Number, want.Start, want.End, want.Line, want.TextLine)
		}
	}
}

func TestStripTagsEmptyCueLists(t *testing.T) {
	tests := []struct {
		name string
		cues []srt.Cue
	}{
		{"nil", nil},
		{"empty", []srt.Cue{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			StripTags(tc.cues)
			if len(tc.cues) != 0 {
				t.Errorf("cue count = %d, want 0", len(tc.cues))
			}
		})
	}
}

func assertTextLines(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("text lines = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("text line %d = %q, want %q", i, got[i], want[i])
		}
	}
}
