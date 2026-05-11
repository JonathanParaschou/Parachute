package cli

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	kib = 1024
	mib = kib * 1024
	gib = mib * 1024
	tib = gib * 1024
)

func parseSize(input string) (uint64, error) {
	s := strings.TrimSpace(strings.ToUpper(input))
	if s == "" {
		return 0, fmt.Errorf("size is required")
	}

	multiplier := uint64(1)
	for _, suffix := range []struct {
		text string
		mul  uint64
	}{
		{"TB", tib},
		{"T", tib},
		{"GB", gib},
		{"G", gib},
		{"MB", mib},
		{"M", mib},
		{"KB", kib},
		{"K", kib},
		{"B", 1},
	} {
		if strings.HasSuffix(s, suffix.text) {
			multiplier = suffix.mul
			s = strings.TrimSpace(strings.TrimSuffix(s, suffix.text))
			break
		}
	}

	value, err := strconv.ParseFloat(s, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid size %q", input)
	}
	return uint64(value * float64(multiplier)), nil
}

func formatSize(bytes uint64) string {
	if bytes >= tib {
		return fmt.Sprintf("%.2f TB", float64(bytes)/float64(tib))
	}
	if bytes >= gib {
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gib))
	}
	if bytes >= mib {
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mib))
	}
	if bytes >= kib {
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kib))
	}
	return fmt.Sprintf("%d B", bytes)
}
