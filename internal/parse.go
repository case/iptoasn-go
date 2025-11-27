package internal

import (
	"bufio"
	"fmt"
	"io"
	"net/netip"
	"strconv"
	"strings"
)

// ASNRecord represents a parsed line from ip2asn files.
type ASNRecord struct {
	StartIP      netip.Addr
	EndIP        netip.Addr
	ASN          uint32
	Country      string
	Organization string
}

// CountryRecord represents a parsed line from ip2country files.
type CountryRecord struct {
	StartIP netip.Addr
	EndIP   netip.Addr
	Country string
}

// ParseASNRecords parses ip2asn TSV data from a reader.
// Format: start_ip	end_ip	asn	country	description
// Records with ASN 0 ("Not routed") are included so users can distinguish
// between "IP not in database" and "IP known but not routed".
func ParseASNRecords(r io.Reader) ([]ASNRecord, error) {
	var records []ASNRecord
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}

		startIP, err := netip.ParseAddr(fields[0])
		if err != nil {
			continue
		}

		endIP, err := netip.ParseAddr(fields[1])
		if err != nil {
			continue
		}

		asn, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			continue
		}

		records = append(records, ASNRecord{
			StartIP:      startIP,
			EndIP:        endIP,
			ASN:          uint32(asn),
			Country:      fields[3],
			Organization: fields[4],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	fmt.Printf("parsed %d ASN records\n", len(records))
	return records, nil
}

// ParseCountryRecords parses ip2country TSV data from a reader.
// Format: start_ip	end_ip	country
// Records with country "None" are included so users can distinguish
// between "IP not in database" and "IP known but no country assigned".
func ParseCountryRecords(r io.Reader) ([]CountryRecord, error) {
	var records []CountryRecord
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}

		startIP, err := netip.ParseAddr(fields[0])
		if err != nil {
			continue
		}

		endIP, err := netip.ParseAddr(fields[1])
		if err != nil {
			continue
		}

		records = append(records, CountryRecord{
			StartIP: startIP,
			EndIP:   endIP,
			Country: fields[2],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}

	fmt.Printf("parsed %d country records\n", len(records))
	return records, nil
}
