package services

import (
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	tests := []struct {
		port int
		want string
	}{
		// known ports
		{21, "ftp"},
		{22, "ssh"},
		{80, "http"},
		{139, "netbios-ssn"},
		{443, "https"},
		{445, "microsoft-ds"},
		{3389, "rdp"},
		{8080, "http-proxy"},

		// unknown ports
		{1, "-"},
		{9999, "-"},
		{65535, "-"},
	}

	for _, tc := range tests {
		if got := Name(tc.port); got != tc.want {
			t.Errorf("Name(%d) = %q, want %q", tc.port, got, tc.want)
		}
	}
}

// Every name in the table should be non-empty and free of stray spaces or commas,
// which catches typos like "netbios-ssn," that a single-row test could miss.
func TestCommonTableIsClean(t *testing.T) {
	for port, name := range common {
		if name == "" || strings.ContainsAny(name, " ,\t") {
			t.Errorf("common[%d] = %q: empty or contains a space/comma", port, name)
		}
	}
}
