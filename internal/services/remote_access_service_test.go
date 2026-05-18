package services

import "testing"

func TestFirstNonEmptyLine(t *testing.T) {
	got := firstNonEmptyLine("\n\n100.64.0.10\n100.64.0.11\n")
	if got != "100.64.0.10" {
		t.Fatalf("firstNonEmptyLine returned %q", got)
	}
}

func TestDashboardURL(t *testing.T) {
	got := dashboardURL("100.64.0.10")
	if got != "http://100.64.0.10:8080" {
		t.Fatalf("dashboardURL returned %q", got)
	}
}
