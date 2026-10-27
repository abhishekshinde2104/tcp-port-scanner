package targets

import (
	"slices"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		want    []string
		wantErr bool
	}{
		{"single host", "scanme.nmap.org", []string{"scanme.nmap.org"}, false},
		{"list", "a.example, 10.0.0.1 ,b.example", []string{"a.example", "10.0.0.1", "b.example"}, false},
		{"/30 drops network and broadcast", "10.0.0.0/30", []string{"10.0.0.1", "10.0.0.2"}, false},
		{"unmasked address is masked", "10.0.0.6/30", []string{"10.0.0.5", "10.0.0.6"}, false},
		{"/31 keeps both", "10.0.0.0/31", []string{"10.0.0.0", "10.0.0.1"}, false},
		{"/32 is one host", "10.0.0.7/32", []string{"10.0.0.7"}, false},
		{"mixed", "host,192.168.1.0/30", []string{"host", "192.168.1.1", "192.168.1.2"}, false},
		{"ipv6 /126", "fd00::/126", []string{"fd00::", "fd00::1", "fd00::2", "fd00::3"}, false},

		{"empty", "", nil, true},
		{"empty part", "a,,b", nil, true},
		{"trailing comma", "a,", nil, true},
		{"bad prefix length", "10.0.0.0/33", nil, true},
		{"not an address", "abc/24", nil, true},
		{"too large", "10.0.0.0/16", nil, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.spec)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("Parse(%q) = %v, want an error", tc.spec, got)
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

func TestParseLimit(t *testing.T) {
	got, err := Parse("10.0.0.0/20") // 4096 addresses - network - broadcast
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4094 {
		t.Errorf("got %d hosts, want 4094", len(got))
	}

	if _, err := Parse("10.0.0.0/20,10.1.0.0/20"); err == nil {
		t.Error("two /20 ranges should exceed MaxHosts")
	}
}

func TestResolveDropsDuplicates(t *testing.T) {
	got, err := Resolve([]string{"127.0.0.1", "127.0.0.1", "127.0.0.2"})
	if err != nil {
		t.Fatal(err)
	}
	want := []Target{{"127.0.0.1", "127.0.0.1"}, {"127.0.0.2", "127.0.0.2"}}
	if !slices.Equal(got, want) {
		t.Errorf("Resolve = %v, want %v", got, want)
	}
}
