package ports

import (
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    []int
		wantErr bool
	}{
		// valid input
		{"single", "80", []int{80}, false},
		{"range", "1-5", []int{1, 2, 3, 4, 5}, false},
		{"duplicates", "22,80,22", []int{22, 80}, false},
		{"sorted", "443,22,80", []int{22, 80, 443}, false},
		{"spaces", " 443 , 80 ", []int{80, 443}, false},
		{"overlap", "1-5,3-7", []int{1, 2, 3, 4, 5, 6, 7}, false},
		{"single-port range", "80-80", []int{80}, false},
		{"lowest port", "1", []int{1}, false},
		{"highest port", "65535", []int{65535}, false},
		{"spaced range", "1 - 3", []int{1, 2, 3}, false},

		// invalid input
		{"bad text", "abc", nil, true},
		{"zero", "0", nil, true},
		{"too big", "70000", nil, true},
		{"negative", "-5", nil, true},
		{"reversed", "100-50", nil, true},
		{"double dash", "1--5", nil, true},
		{"open-ended range", "5-", nil, true},
		{"empty part", "22,,80", nil, true},
		{"trailing comma", "22,", nil, true},
		{"empty", "", nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.spec)

			if tc.wantErr {
				if err == nil {
					t.Errorf("Parse(%q) expected an error, got %v", tc.spec, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse(%q) unexpected error: %v", tc.spec, err)
			}
			if !slices.Equal(got, tc.want) {
				t.Errorf("Parse(%q) = %v, want %v", tc.spec, got, tc.want)
			}
		})
	}
}

func TestParseFullRange(t *testing.T) {
	got, err := Parse("1-65535")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 65535 || got[0] != 1 || got[len(got)-1] != 65535 {
		t.Errorf("got %d ports from %d to %d, want 65535 ports from 1 to 65535",
			len(got), got[0], got[len(got)-1])
	}
}
