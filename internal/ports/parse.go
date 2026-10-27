package ports

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func parsePort(s string) (int, error) {
	s1 := strings.TrimSpace(s)
	n, err := strconv.Atoi(s1)

	if err != nil {
		return 0, fmt.Errorf("invalid port %q", s)
	}

	if n < 1 || n > 65535 {
		return 0, fmt.Errorf("port %d out of range 1-65535", n)
	}

	return n, nil
}

func Parse(spec string) ([]int, error) {
	splits := strings.Split(spec, ",")

	seen := map[int]bool{}

	for _, part := range splits {

		if strings.ContainsAny(part, "-") {
			pieces := strings.Split(part, "-")
			if len(pieces) != 2 {
				return nil, fmt.Errorf("invalid range %q", part)
			}

			lo, err := parsePort(pieces[0])
			if err != nil {
				return nil, err
			}

			hi, err := parsePort(pieces[1])
			if err != nil {
				return nil, err
			}

			if lo > hi {
				return nil, fmt.Errorf("invalid range %q: start is greater than end", part)
			}

			// Add numbers from low to high in the seen map
			for p := lo; p <= hi; p++ {
				seen[p] = true
			}
		} else {
			p, err := parsePort(part)
			if err != nil {
				return nil, err
			}
			seen[p] = true

		}
	}

	// remove duplicates, sort them
	result := make([]int, 0, len(seen))
	for p := range seen {
		result = append(result, p)
	}
	slices.Sort(result)
	return result, nil
}
