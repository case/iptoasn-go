package internal

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ANSI color codes
const (
	colorCyan  = "\033[36m"
	colorReset = "\033[0m"
)

// FormatNumber returns a number formatted with commas.
func FormatNumber(n int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", n)
}

// FormatDuration returns a colored duration string.
func FormatDuration(d time.Duration) string {
	return fmt.Sprintf("%s%s%s", colorCyan, d.Round(time.Millisecond), colorReset)
}

// FormatFile returns a colored filename string.
func FormatFile(name string) string {
	return fmt.Sprintf("%s%s%s", colorCyan, name, colorReset)
}

// FormatMB returns bytes formatted as MB with one decimal place.
func FormatMB(bytes int64) string {
	mb := float64(bytes) / (1024 * 1024)
	return fmt.Sprintf("%.1f MB", mb)
}

// ipRangeToPrefixes converts an IP range to a list of CIDR prefixes that
// exactly cover the range.
func ipRangeToPrefixes(start, end netip.Addr) []netip.Prefix {
	var prefixes []netip.Prefix

	for start.Compare(end) <= 0 {
		bits := start.BitLen()
		maxBits := bits

		for prefixLen := bits; prefixLen >= 0; prefixLen-- {
			prefix, err := start.Prefix(prefixLen)
			if err != nil {
				break
			}
			if prefix.Addr() != start {
				break
			}
			last := lastAddr(prefix)
			if last.Compare(end) > 0 {
				break
			}
			maxBits = prefixLen
		}

		prefix, _ := start.Prefix(maxBits)
		prefixes = append(prefixes, prefix)

		last := lastAddr(prefix)
		next := last.Next()
		if !next.IsValid() || next.Compare(last) < 0 {
			break
		}
		start = next
	}

	return prefixes
}

// lastAddr returns the last address in a prefix.
func lastAddr(p netip.Prefix) netip.Addr {
	addr := p.Addr()
	bits := p.Bits()
	addrBits := addr.BitLen()

	if bits == addrBits {
		return addr
	}

	b := addr.AsSlice()
	for i := bits; i < addrBits; i++ {
		byteIdx := i / 8
		bitIdx := 7 - (i % 8)
		b[byteIdx] |= 1 << bitIdx
	}

	if addr.Is4() {
		return netip.AddrFrom4([4]byte(b))
	}
	return netip.AddrFrom16([16]byte(b))
}

// prefixToIPNet converts a netip.Prefix to *net.IPNet for mmdbwriter compatibility.
func prefixToIPNet(p netip.Prefix) *net.IPNet {
	addr := p.Addr()
	bits := p.Bits()

	return &net.IPNet{
		IP:   addr.AsSlice(),
		Mask: net.CIDRMask(bits, addr.BitLen()),
	}
}
