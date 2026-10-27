package output

import (
	"fmt"
	"io"
)

// art is "GO PORT SCAN" in the ANSI Shadow figlet style. A raw string (backticks)
// keeps every character exactly as written, so no escaping is needed.
const art = `
 ██████╗  ██████╗   ██████╗  ██████╗ ██████╗ ████████╗  ███████╗ ██████╗  █████╗ ███╗   ██╗
██╔════╝ ██╔═══██╗  ██╔══██╗██╔═══██╗██╔══██╗╚══██╔══╝  ██╔════╝██╔════╝ ██╔══██╗████╗  ██║
██║  ███╗██║   ██║  ██████╔╝██║   ██║██████╔╝   ██║     ███████╗██║      ███████║██╔██╗ ██║
██║   ██║██║   ██║  ██╔═══╝ ██║   ██║██╔══██╗   ██║     ╚════██║██║      ██╔══██║██║╚██╗██║
╚██████╔╝╚██████╔╝  ██║     ╚██████╔╝██║  ██║   ██║     ███████║╚██████╗ ██║  ██║██║ ╚████║
 ╚═════╝  ╚═════╝   ╚═╝      ╚═════╝ ╚═╝  ╚═╝   ╚═╝     ╚══════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═══╝
`

// Banner writes the ASCII-art title.
func Banner(w io.Writer) {
	fmt.Fprint(w, art)
}
