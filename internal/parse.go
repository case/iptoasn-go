package internal

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net/netip"
	"strconv"
	"strings"
	"time"
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

// Estimated record counts for slice preallocation (based on typical iptoasn.com data sizes)
const (
	estimatedASNRecords     = 700_000
	estimatedCountryRecords = 700_000
)

// Minimum record thresholds - fail if below these counts (likely truncated/corrupt data)
const (
	minASNRecords     = 100_000
	minCountryRecords = 100_000
)

// ParseASNRecords parses ip2asn TSV data from a reader.
// Format: start_ip	end_ip	asn	country	description
// Records with ASN 0 ("Not routed") are included so users can distinguish
// between "IP not in database" and "IP known but not routed".
func ParseASNRecords(r io.Reader) ([]ASNRecord, error) {
	start := time.Now()
	records := make([]ASNRecord, 0, estimatedASNRecords)
	scanner := bufio.NewScanner(r)
	lineNum := 0
	skipped := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			slog.Warn("Skipping malformed line", "line", lineNum, "error", "expected 5 fields", "got", len(fields))
			skipped++
			continue
		}

		startIP, err := netip.ParseAddr(fields[0])
		if err != nil {
			slog.Warn("Skipping malformed line", "line", lineNum, "error", "invalid start IP", "value", fields[0])
			skipped++
			continue
		}

		endIP, err := netip.ParseAddr(fields[1])
		if err != nil {
			slog.Warn("Skipping malformed line", "line", lineNum, "error", "invalid end IP", "value", fields[1])
			skipped++
			continue
		}

		asn, err := strconv.ParseUint(fields[2], 10, 32)
		if err != nil {
			slog.Warn("Skipping malformed line", "line", lineNum, "error", "invalid ASN", "value", fields[2])
			skipped++
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

	if len(records) == 0 {
		if skipped > 0 {
			return nil, fmt.Errorf("no valid records found (%d malformed lines)", skipped)
		}
		return nil, fmt.Errorf("no records found in input")
	}

	slog.Info("Parsed ASN records", "count", len(records), "skipped", skipped, "duration", time.Since(start).Round(time.Millisecond))
	return records, nil
}

// ParseCountryRecords parses ip2country TSV data from a reader.
// Format: start_ip	end_ip	country
// Records with country "None" are included so users can distinguish
// between "IP not in database" and "IP known but no country assigned".
func ParseCountryRecords(r io.Reader) ([]CountryRecord, error) {
	start := time.Now()
	records := make([]CountryRecord, 0, estimatedCountryRecords)
	scanner := bufio.NewScanner(r)
	lineNum := 0
	skipped := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			slog.Warn("Skipping malformed line", "line", lineNum, "error", "expected 3 fields", "got", len(fields))
			skipped++
			continue
		}

		startIP, err := netip.ParseAddr(fields[0])
		if err != nil {
			slog.Warn("Skipping malformed line", "line", lineNum, "error", "invalid start IP", "value", fields[0])
			skipped++
			continue
		}

		endIP, err := netip.ParseAddr(fields[1])
		if err != nil {
			slog.Warn("Skipping malformed line", "line", lineNum, "error", "invalid end IP", "value", fields[1])
			skipped++
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

	if len(records) == 0 {
		if skipped > 0 {
			return nil, fmt.Errorf("no valid records found (%d malformed lines)", skipped)
		}
		return nil, fmt.Errorf("no records found in input")
	}

	slog.Info("Parsed country records", "count", len(records), "skipped", skipped, "duration", time.Since(start).Round(time.Millisecond))
	return records, nil
}
