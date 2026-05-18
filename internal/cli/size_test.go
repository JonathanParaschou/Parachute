package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseSize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint64
	}{
		{name: "bytes", input: "512", want: 512},
		{name: "kilobytes", input: "2KB", want: 2 * kib},
		{name: "gigabytes", input: "1.5GB", want: uint64(1.5 * float64(gib))},
		{name: "terabytes", input: "1TB", want: tib},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSize(tt.input)
			if err != nil {
				t.Fatalf("parseSize returned error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("parseSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseStorageAddArgsAcceptsLimitAfterPath(t *testing.T) {
	path, limit, err := parseStorageAddArgs([]string{"/data", "--limit", "500GB"})
	if err != nil {
		t.Fatalf("parseStorageAddArgs returned error: %v", err)
	}
	if path != "/data" {
		t.Fatalf("path = %q, want /data", path)
	}
	if limit != "500GB" {
		t.Fatalf("limit = %q, want 500GB", limit)
	}
}

func TestRemoteStatusUsage(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := Run([]string{"remote"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run returned %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "usage: parachute remote status") {
		t.Fatalf("stderr = %q", stderr.String())
	}
}
