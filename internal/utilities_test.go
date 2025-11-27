package internal

import (
	"net/netip"
	"strings"
	"testing"
)

func TestIpRangeToPrefixes_SingleIP(t *testing.T) {
	start := netip.MustParseAddr("1.0.0.0")
	end := netip.MustParseAddr("1.0.0.0")

	prefixes, err := ipRangeToPrefixes(start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prefixes) != 1 {
		t.Fatalf("expected 1 prefix, got %d", len(prefixes))
	}
	if prefixes[0].String() != "1.0.0.0/32" {
		t.Errorf("expected 1.0.0.0/32, got %s", prefixes[0])
	}
}

func TestIpRangeToPrefixes_SinglePrefix(t *testing.T) {
	// 1.0.0.0 - 1.0.0.255 is exactly a /24
	start := netip.MustParseAddr("1.0.0.0")
	end := netip.MustParseAddr("1.0.0.255")

	prefixes, err := ipRangeToPrefixes(start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prefixes) != 1 {
		t.Fatalf("expected 1 prefix, got %d: %v", len(prefixes), prefixes)
	}
	if prefixes[0].String() != "1.0.0.0/24" {
		t.Errorf("expected 1.0.0.0/24, got %s", prefixes[0])
	}
}

func TestIpRangeToPrefixes_MultiplePrefixes(t *testing.T) {
	// 1.0.1.0 - 1.0.3.255 requires multiple prefixes:
	// 1.0.1.0/24 + 1.0.2.0/23
	start := netip.MustParseAddr("1.0.1.0")
	end := netip.MustParseAddr("1.0.3.255")

	prefixes, err := ipRangeToPrefixes(start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prefixes) != 2 {
		t.Fatalf("expected 2 prefixes, got %d: %v", len(prefixes), prefixes)
	}
	if prefixes[0].String() != "1.0.1.0/24" {
		t.Errorf("expected 1.0.1.0/24, got %s", prefixes[0])
	}
	if prefixes[1].String() != "1.0.2.0/23" {
		t.Errorf("expected 1.0.2.0/23, got %s", prefixes[1])
	}
}

func TestIpRangeToPrefixes_IPv6(t *testing.T) {
	start := netip.MustParseAddr("2001:4:112::")
	end := netip.MustParseAddr("2001:4:112:ffff:ffff:ffff:ffff:ffff")

	prefixes, err := ipRangeToPrefixes(start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(prefixes) != 1 {
		t.Fatalf("expected 1 prefix, got %d: %v", len(prefixes), prefixes)
	}
	if prefixes[0].String() != "2001:4:112::/48" {
		t.Errorf("expected 2001:4:112::/48, got %s", prefixes[0])
	}
}

func TestIpRangeToPrefixes_InvalidRange(t *testing.T) {
	// Test that start > end returns an error
	start := netip.MustParseAddr("8.8.8.255")
	end := netip.MustParseAddr("8.8.8.0")

	_, err := ipRangeToPrefixes(start, end)
	if err == nil {
		t.Fatal("expected error for invalid range (start > end), got nil")
	}
	if !strings.Contains(err.Error(), "invalid IP range") {
		t.Errorf("expected 'invalid IP range' error, got: %v", err)
	}
}

func TestLastAddr(t *testing.T) {
	tests := []struct {
		prefix string
		want   string
	}{
		{"1.0.0.0/24", "1.0.0.255"},
		{"1.0.0.0/32", "1.0.0.0"},
		{"10.0.0.0/8", "10.255.255.255"},
		{"2001:4:112::/48", "2001:4:112:ffff:ffff:ffff:ffff:ffff"},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			prefix := netip.MustParsePrefix(tt.prefix)
			got := lastAddr(prefix)
			if got.String() != tt.want {
				t.Errorf("lastAddr(%s) = %s, want %s", tt.prefix, got, tt.want)
			}
		})
	}
}

func TestPrefixToIPNet(t *testing.T) {
	tests := []struct {
		prefix string
	}{
		{"1.0.0.0/24"},
		{"10.0.0.0/8"},
		{"192.168.1.0/24"},
		{"2001:4:112::/48"},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			prefix := netip.MustParsePrefix(tt.prefix)
			ipnet := prefixToIPNet(prefix)

			// Convert back to string and compare
			if ipnet.String() != tt.prefix {
				t.Errorf("prefixToIPNet(%s) = %s, want %s", tt.prefix, ipnet.String(), tt.prefix)
			}
		})
	}
}
