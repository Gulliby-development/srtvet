package ops

import (
	"testing"

	"github.com/Gulliby-development/srtvet/internal/srt"
)

const duplicateMessage = "same start, end and text as the previous cue"

func TestCheckDuplicateCue(t *testing.T) {
	tests := []struct {
		name    string
		cues    []srt.Cue
		maxLine int
		want    []Finding
	}{
		{
			name: "same start, end and text",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello."),
				cue(2, 1000, 2000, 5, 7, "Hello."),
			},
			want: []Finding{
				{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"},
				{5, "SV008", duplicateMessage},
			},
		},
		{
			name: "three identical cues in a row",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello."),
				cue(2, 1000, 2000, 5, 7, "Hello."),
				cue(3, 1000, 2000, 9, 11, "Hello."),
			},
			want: []Finding{
				{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"},
				{5, "SV008", duplicateMessage},
				{9, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"},
				{9, "SV008", duplicateMessage},
			},
		},
		{
			name: "both text lines are equal",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello.", "World."),
				cue(2, 1000, 2000, 5, 7, "Hello.", "World."),
			},
			want: []Finding{
				{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"},
				{5, "SV008", duplicateMessage},
			},
		},
		{
			name: "different second text line",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello.", "World."),
				cue(2, 1000, 2000, 5, 7, "Hello.", "world."),
			},
			want: []Finding{{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"}},
		},
		{
			name: "different number of text lines",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello.", "World."),
				cue(2, 1000, 2000, 5, 7, "Hello."),
			},
			want: []Finding{{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"}},
		},
		{
			name: "case difference",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello."),
				cue(2, 1000, 2000, 5, 7, "hello."),
			},
			want: []Finding{{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"}},
		},
		{
			name: "trailing space difference",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello."),
				cue(2, 1000, 2000, 5, 7, "Hello. "),
			},
			want: []Finding{{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"}},
		},
		{
			name: "different end",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello."),
				cue(2, 1000, 2500, 5, 7, "Hello."),
			},
			want: []Finding{{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"}},
		},
		{
			name: "different start",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello."),
				cue(2, 1500, 2500, 5, 7, "Hello."),
			},
			want: []Finding{{5, "SV003", "starts at 00:00:01,500, before the previous cue ends at 00:00:02,000"}},
		},
		{
			name: "same text with later times",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello."),
				cue(2, 3000, 4000, 5, 7, "Hello."),
			},
		},
		{
			name: "copy that is not adjacent",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "Hello."),
				cue(2, 3000, 4000, 5, 7, "Different."),
				cue(3, 1000, 2000, 9, 11, "Hello."),
			},
			want: []Finding{{9, "SV004", "starts at 00:00:01,000, before the previous cue starts at 00:00:03,000"}},
		},
		{
			name: "two cues without text",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 0),
				cue(2, 1000, 2000, 5, 0),
			},
			want: []Finding{
				{1, "SV005", msgSV005},
				{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"},
				{5, "SV005", msgSV005},
			},
		},
		{
			name: "single cue",
			cues: []srt.Cue{cue(1, 1000, 2000, 1, 3, "Hello.")},
		},
		{
			name: "invalid-duration copies",
			cues: []srt.Cue{
				cue(1, 1000, 1000, 1, 3, "Hello."),
				cue(2, 1000, 1000, 5, 7, "Hello."),
			},
			want: []Finding{
				{1, "SV002", msgSV002},
				{5, "SV002", msgSV002},
				{5, "SV008", duplicateMessage},
			},
		},
		{
			name: "max line does not disable SV008",
			cues: []srt.Cue{
				cue(1, 1000, 2000, 1, 3, "long"),
				cue(2, 1000, 2000, 5, 7, "long"),
			},
			maxLine: 2,
			want: []Finding{
				{3, "SV006", "line 1 has 4 characters, the limit is 2"},
				{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"},
				{7, "SV006", "line 1 has 4 characters, the limit is 2"},
				{5, "SV008", duplicateMessage},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assertFindings(t, Check(tc.cues, tc.maxLine), tc.want)
		})
	}
}

func TestCheckAllDuplicateAfterSV007(t *testing.T) {
	cues := []srt.Cue{
		cue(1, 1000, 2000, 1, 3, "long"),
		cue(2, 1000, 2000, 5, 7, "long"),
	}
	want := []Finding{
		{3, "SV006", "line 1 has 4 characters, the limit is 2"},
		{1, "SV007", "reading speed 4.0 characters per second, the limit is 2"},
		{5, "SV003", "starts at 00:00:01,000, before the previous cue ends at 00:00:02,000"},
		{7, "SV006", "line 1 has 4 characters, the limit is 2"},
		{5, "SV007", "reading speed 4.0 characters per second, the limit is 2"},
		{5, "SV008", duplicateMessage},
	}
	assertFindings(t, CheckAll(cues, 2, 2), want)
}
