package targets

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

// MaxHosts caps how many hosts one -target value may expand to,
// so a typo like 10.0.0.0/8 can't start a scan of 16 million addresses.
const MaxHosts = 4096

// Target is one host to scan.
type Target struct {
	Name string `json:"name"` // what the user typed (the IP itself for CIDR entries)
	IP   string `json:"ip"`   // the address we connect to
}

// Parse splits a comma-separated list of host names, IPs and CIDR ranges,
// e.g. "scanme.nmap.org,10.0.0.5,192.168.1.0/24", into single entries.
func Parse(spec string) ([]string, error) {
	var out []string
	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty target in %q", spec)
		}

		if strings.Contains(part, "/") {
			ips, err := expandCIDR(part)
			if err != nil {
				return nil, err
			}
			out = append(out, ips...)
		} else {
			out = append(out, part)
		}

		if len(out) > MaxHosts {
			return nil, fmt.Errorf("too many targets: more than %d hosts", MaxHosts)
		}
	}
	return out, nil
}

// expandCIDR lists the host addresses in a range like "192.168.1.0/24".
func expandCIDR(cidr string) ([]string, error) {
	prefix, err := netip.ParsePrefix(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR range %q", cidr)
	}
	prefix = prefix.Masked() // 192.168.1.77/24 → 192.168.1.0/24

	hostBits := prefix.Addr().BitLen() - prefix.Bits()
	if hostBits > 12 { // 2^12 = 4096 addresses
		return nil, fmt.Errorf("CIDR range %q is too large: at most %d addresses", cidr, MaxHosts)
	}

	var ips []string
	for a := prefix.Addr(); prefix.Contains(a); a = a.Next() {
		ips = append(ips, a.String())
	}

	// In IPv4 ranges larger than /31 the first address names the network
	// and the last one is the broadcast address; neither is a real host.
	if prefix.Addr().Is4() && prefix.Bits() < 31 {
		ips = ips[1 : len(ips)-1]
	}
	return ips, nil
}

// Resolve looks up each entry once and drops duplicates, e.g. two names for the same IP.
func Resolve(entries []string) ([]Target, error) {
	seen := make(map[string]bool)
	var out []Target
	for _, e := range entries {
		addrs, err := net.LookupHost(e) // an IP literal comes back as itself, no DNS query
		if err != nil {
			return nil, fmt.Errorf("cannot resolve %s", e)
		}
		ip := addrs[0]
		if seen[ip] {
			continue
		}
		seen[ip] = true
		out = append(out, Target{Name: e, IP: ip})
	}
	return out, nil
}
