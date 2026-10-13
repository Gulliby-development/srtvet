package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReadmeContents checks that README.md at the repository root holds the
// usage lines, the worked examples and the required headings. go test runs in
// the package directory, so the README is two directories up.
func TestReadmeContents(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "README.md"))
	if err != nil {
		t.Fatalf("cannot read README.md: %v", err)
	}
	readme := strings.ReplaceAll(string(data), "\r\n", "\n")

	tests := []struct {
		name string
		want string
	}{
		{"usage line check", "usage: srtvet check [--max-line N] FILE..."},
		{"usage line fix", "       srtvet fix [--shift DURATION] [--stretch RATIO] [--sort] [--drop-empty] [--strip-tags] [--out FILE] FILE"},

		{"check output SV003", "movie.srt:5: SV003 starts at 00:00:03,000, before the previous cue ends at 00:00:03,500"},
		{"check output SV001", "movie.srt:9: SV001 cue number 4, expected 3"},
		{"check output SV002", "movie.srt:9: SV002 end is not after start"},
		{"check output SV005", "movie.srt:9: SV005 no text"},
		{"check output SV004", "movie.srt:12: SV004 starts at 00:00:04,000, before the previous cue starts at 00:00:06,000"},
		{"check output SV006", "movie.srt:14: SV006 line 1 has 78 characters, the limit is 42"},
		{"check output summary", "6 problems in 1 file"},

		{"fix output first timing", "00:00:00,543 --> 00:00:03,149"},
		{"fix output first text", "<i>Hello there.</i>"},
		{"fix output second text", "Second line."},
		{"fix output second timing", "00:01:12,490 --> 00:01:14,617"},
		{"fix output third text", "Third, but listed first."},

		{"heading install", "## Install"},
		{"heading usage", "## Usage"},
		{"heading checks", "## Checks"},
		{"heading fixing", "## Fixing"},
		{"heading exit codes", "## Exit codes"},
		{"heading development", "## Development"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(readme, tt.want) {
				t.Errorf("README.md does not contain %q", tt.want)
			}
		})
	}
}
