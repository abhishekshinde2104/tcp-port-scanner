package output

import (
	"fmt"
	"io"
	"strings"
)

const barWidth = 20

// Progress returns a callback for scanner.Scan that redraws one status line on w:
//
//	██████████░░░░░░░░░░  50%  512/1024 ports
//
// It only redraws when the percentage changes, and it clears the line when the scan is done.
func Progress(w io.Writer) func(done, total int) {
	last := -1 // kept between calls: the closure "remembers" it
	return func(done, total int) {
		pct := done * 100 / total
		if pct == last {
			return
		}
		last = pct

		filled := pct * barWidth / 100
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		// \r moves the cursor back to the start of the line, so the next draw overwrites this one
		fmt.Fprintf(w, "\r%s %3d%%  %d/%d ports", bar, pct, done, total)

		if done == total {
			fmt.Fprintf(w, "\r%s\r", strings.Repeat(" ", 70))
		}
	}
}
