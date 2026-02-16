package gitserver

import (
	"fmt"
	"io"
)

// writePktLine writes a single pkt-line formatted string.
// Format: 4-char hex length prefix + data. Length includes the 4-byte prefix itself.
func writePktLine(w io.Writer, data string) error {
	_, err := fmt.Fprintf(w, "%04x%s", len(data)+4, data)
	return err
}

// writePktFlush writes a flush packet (0000).
func writePktFlush(w io.Writer) error {
	_, err := fmt.Fprint(w, "0000")
	return err
}

// writeServiceAdvertisement writes the pkt-line service header that wraps
// the output of `git <service> --advertise-refs`. This is the first line
// the client sees during info/refs discovery.
func writeServiceAdvertisement(w io.Writer, service string) error {
	return writePktLine(w, "# service="+service+"\n")
}
