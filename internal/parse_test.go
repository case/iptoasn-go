package internal

import (
	"os"
	"strings"
	"testing"
)

func TestParseASNRecords(t *testing.T) {
	f, err := os.Open("testdata/ip2asn-sample.tsv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	records, err := ParseASNRecords(f)
	if err != nil {
		t.Fatalf("ParseASNRecords failed: %v", err)
	}

	// Should have 6 records (including ASN 0 entries for reserved networks)
	if len(records) != 6 {
		t.Fatalf("expected 6 records, got %d", len(records))
	}

	// Check first record (Cloudflare)
	if records[0].ASN != 13335 {
		t.Errorf("expected ASN 13335, got %d", records[0].ASN)
	}
	if records[0].Country != "US" {
		t.Errorf("expected country US, got %s", records[0].Country)
	}
	if records[0].Organization != "CLOUDFLARENET" {
		t.Errorf("expected org CLOUDFLARENET, got %s", records[0].Organization)
	}

	// Check that ASN 0 entries are included
	if records[1].ASN != 0 {
		t.Errorf("expected ASN 0 for not-routed entry, got %d", records[1].ASN)
	}

	// Check Google entry
	if records[3].ASN != 15169 {
		t.Errorf("expected ASN 15169 for Google, got %d", records[3].ASN)
	}
}

func TestParseASNRecords_IPv6(t *testing.T) {
	f, err := os.Open("testdata/ip2asn-v6-sample.tsv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	records, err := ParseASNRecords(f)
	if err != nil {
		t.Fatalf("ParseASNRecords failed: %v", err)
	}

	if len(records) != 4 {
		t.Fatalf("expected 4 records, got %d", len(records))
	}

	// Check aliased network entry (2001::/32 - Teredo)
	if records[0].ASN != 0 {
		t.Errorf("expected ASN 0 for aliased network, got %d", records[0].ASN)
	}

	// Check AS112 entry
	if records[1].ASN != 112 {
		t.Errorf("expected ASN 112, got %d", records[1].ASN)
	}

	// Check Cloudflare IPv6
	if records[3].ASN != 13335 {
		t.Errorf("expected ASN 13335, got %d", records[3].ASN)
	}
	if !records[3].StartIP.Is6() {
		t.Error("expected IPv6 address")
	}
}

func TestParseCountryRecords(t *testing.T) {
	f, err := os.Open("testdata/ip2country-sample.tsv")
	if err != nil {
		t.Fatalf("failed to open test file: %v", err)
	}
	defer f.Close()

	records, err := ParseCountryRecords(f)
	if err != nil {
		t.Fatalf("ParseCountryRecords failed: %v", err)
	}

	// Should have 5 records (including None entries)
	if len(records) != 5 {
		t.Fatalf("expected 5 records, got %d", len(records))
	}

	// Check first record
	if records[0].Country != "US" {
		t.Errorf("expected country US, got %s", records[0].Country)
	}

	// Check that None entries are included
	if records[1].Country != "None" {
		t.Errorf("expected country None, got %s", records[1].Country)
	}

	// Check Australia entry
	if records[2].Country != "AU" {
		t.Errorf("expected country AU, got %s", records[2].Country)
	}
}

func TestParseASNRecords_EmptyLines(t *testing.T) {
	// Test that empty lines are skipped
	input := "1.0.0.0\t1.0.0.255\t13335\tUS\tCLOUDFLARENET\n\n8.8.8.0\t8.8.8.255\t15169\tUS\tGOOGLE\n"

	records, err := ParseASNRecords(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseASNRecords failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestParseASNRecords_MalformedLines(t *testing.T) {
	// Test that malformed lines are skipped
	input := "1.0.0.0\t1.0.0.255\t13335\tUS\tCLOUDFLARENET\nmalformed line\n8.8.8.0\t8.8.8.255\t15169\tUS\tGOOGLE\n"

	records, err := ParseASNRecords(strings.NewReader(input))
	if err != nil {
		t.Fatalf("ParseASNRecords failed: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected 2 records (skipping malformed), got %d", len(records))
	}
}
