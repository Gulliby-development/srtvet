package cli

import (
	"bytes"
	"strings"
	"testing"
)

const duplicateSRT = "1\n00:00:01,000 --> 00:00:02,000\nHello.\n\n" +
	"2\n00:00:01,000 --> 00:00:02,000\nHello.\n\n" +
	"3\n00:00:03,000 --> 00:00:04,000\nHello.\n"

func TestRunDuplicateCue(t *testing.T) {
	tests := []struct {
		name string
		args []string
		code int
		want string
	}{
		{
			name: "check reports adjacent duplicate",
			args: []string{"check", "-"},
			code: 1,
			want: "-:5: SV003 starts at 00:00:01,000, before the previous cue ends at 00:00:02,000\n" +
				"-:5: SV008 same start, end and text as the previous cue\n" +
				"2 problems in 1 file\n",
		},
		{
			name: "fix preserves the subtitle content",
			args: []string{"fix", "-"},
			code: 0,
			want: duplicateSRT,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(tc.args, strings.NewReader(duplicateSRT), &stdout, &stderr)
			if code != tc.code {
				t.Errorf("exit code = %d, want %d", code, tc.code)
			}
			if got := stdout.String(); got != tc.want {
				t.Errorf("stdout\n got %q\nwant %q", got, tc.want)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty", got)
			}
			if tc.args[0] == "fix" && strings.Contains(stdout.String(), "SV008") {
				t.Errorf("fix output contains SV008: %q", stdout.String())
			}
		})
	}
}
