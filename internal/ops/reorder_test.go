package ops

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

// numbersOf lists the declared numbers of the cues, for example "[1 3 5]".
func numbersOf(cues []srt.Cue) string {
	nums := make([]int, 0, len(cues))
	for _, c := range cues {
		nums = append(nums, c.Number)
	}
	return fmt.Sprint(nums)
}

// textsOf joins the text of the cues, "-" standing for a cue without text.
func textsOf(cues []srt.Cue) string {
	parts := make([]string, 0, len(cues))
	for _, c := range cues {
		if len(c.Text) == 0 {
			parts = append(parts, "-")
			continue
		}
		parts = append(parts, strings.Join(c.Text, "|"))
	}
	return strings.Join(parts, " ")
}

func TestDropEmpty(t *testing.T) {
	tests := []struct {
		name      string
		cues      []srt.Cue
		wantNums  string
		wantTexts string
	}{
		{"mixed", []srt.Cue{
			cue(1, 1000, 2000, 1, 3, "a"),
			cue(2, 3000, 4000, 5, 0),
			cue(3, 5000, 6000, 8, 10, "b", "c"),
			{Number: 4, Start: 7000, End: 8000, Text: []string{}, Line: 13},
			cue(5, 500, 600, 16, 18, "d"),
		}, "[1 3 5]", "a b|c d"},
		{"all empty", []srt.Cue{cue(1, 1000, 2000, 1, 0), cue(2, 3000, 4000, 4, 0)}, "[]", ""},
		{"none empty", []srt.Cue{cue(2, 3000, 4000, 1, 3, "x"), cue(1, 1000, 2000, 5, 7, "y")}, "[2 1]", "x y"},
		{"empty first and last", []srt.Cue{
			cue(1, 1000, 2000, 1, 0),
			cue(2, 3000, 4000, 4, 6, "x"),
			cue(3, 5000, 6000, 8, 0),
		}, "[2]", "x"},
		{"nil list", nil, "[]", ""},
		{"empty list", []srt.Cue{}, "[]", ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			beforeNums, beforeTexts := numbersOf(tc.cues), textsOf(tc.cues)
			got := DropEmpty(tc.cues)
			if n := numbersOf(got); n != tc.wantNums {
				t.Errorf("numbers = %s, want %s", n, tc.wantNums)
			}
			if s := textsOf(got); s != tc.wantTexts {
				t.Errorf("texts = %q, want %q", s, tc.wantTexts)
			}
			if numbersOf(tc.cues) != beforeNums || textsOf(tc.cues) != beforeTexts {
				t.Errorf("the input list changed: %s %q", numbersOf(tc.cues), textsOf(tc.cues))
			}
		})
	}
}

func TestSort(t *testing.T) {
	tests := []struct {
		name      string
		cues      []srt.Cue
		wantTexts string
		wantNums  string
	}{
		{"start, then end, then file order", []srt.Cue{
			cue(1, 2000, 3000, 0, 0, "a"),
			cue(2, 1000, 2500, 0, 0, "b"),
			cue(3, 1000, 2000, 0, 0, "c"),
			cue(4, 2000, 3000, 0, 0, "d"),
			cue(5, 1000, 2000, 0, 0, "e"),
			cue(6, 500, 9000, 0, 0, "f"),
		}, "f c e b a d", "[6 3 5 2 1 4]"},
		{"equal starts ordered by end", []srt.Cue{
			cue(1, 1000, 3000, 0, 0, "x"),
			cue(2, 1000, 2000, 0, 0, "y"),
		}, "y x", "[2 1]"},
		{"start wins over end", []srt.Cue{
			cue(1, 2000, 2500, 0, 0, "x"),
			cue(2, 1000, 9000, 0, 0, "y"),
		}, "y x", "[2 1]"},
		{"equal cues keep file order", []srt.Cue{
			cue(1, 1000, 2000, 0, 0, "first"),
			cue(2, 1000, 2000, 0, 0, "second"),
			cue(3, 1000, 2000, 0, 0, "third"),
		}, "first second third", "[1 2 3]"},
		{"already sorted", []srt.Cue{
			cue(1, 1000, 2000, 0, 0, "a"),
			cue(2, 3000, 4000, 0, 0),
			cue(3, 5000, 6000, 0, 0, "c"),
		}, "a - c", "[1 2 3]"},
		{"nil list", nil, "", "[]"},
		{"empty list", []srt.Cue{}, "", "[]"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Sort(tc.cues)
			if s := textsOf(tc.cues); s != tc.wantTexts {
				t.Errorf("texts = %q, want %q", s, tc.wantTexts)
			}
			if n := numbersOf(tc.cues); n != tc.wantNums {
				t.Errorf("numbers = %s, want %s", n, tc.wantNums)
			}
		})
	}
}

func TestRenumber(t *testing.T) {
	tests := []struct {
		name     string
		declared []int
		want     string
	}{
		{"mixed numbers", []int{7, 7, 0, 42, 3}, "[1 2 3 4 5]"},
		{"already numbered", []int{1, 2, 3}, "[1 2 3]"},
		{"reversed", []int{3, 2, 1}, "[1 2 3]"},
		{"one cue", []int{9}, "[1]"},
		{"empty list", nil, "[]"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cues := make([]srt.Cue, 0, len(tc.declared))
			for i, n := range tc.declared {
				cues = append(cues, cue(n, int64(i)*1000, int64(i)*1000+500, 0, 0, fmt.Sprintf("t%d", i)))
			}
			before := textsOf(cues)
			Renumber(cues)
			if n := numbersOf(cues); n != tc.want {
				t.Errorf("numbers = %s, want %s", n, tc.want)
			}
			if len(cues) != len(tc.declared) || textsOf(cues) != before {
				t.Errorf("the cues changed order or count: %q, want %q", textsOf(cues), before)
			}
		})
	}
}

func TestSortThenRenumber(t *testing.T) {
	cues := []srt.Cue{
		cue(3, 70000, 72040, 1, 3, "third"),
		cue(1, 1000, 3500, 5, 7, "first"),
		cue(2, 4000, 4500, 10, 0),
	}
	Sort(cues)
	Renumber(cues)
	if got := textsOf(cues); got != "first - third" {
		t.Errorf("texts = %q, want %q", got, "first - third")
	}
	if got := numbersOf(cues); got != "[1 2 3]" {
		t.Errorf("numbers = %s, want [1 2 3]", got)
	}
}
