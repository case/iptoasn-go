package internal

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

// DataType specifies what data to include in the MMDB.
type DataType string

const (
	DataTypeASN     DataType = "asn"
	DataTypeCountry DataType = "country"
)

// IPVersion specifies which IP versions to include.
type IPVersion string

const (
	IPVersionBoth IPVersion = "both"
	IPVersion4    IPVersion = "4"
	IPVersion6    IPVersion = "6"
)

// Config holds the build configuration.
type Config struct {
	OutputFile string
	DataType   DataType
	IPVersion  IPVersion
	SourceFile string // Local TSV/TSV.gz file
	Download   bool   // If true, download from iptoasn.com
}

// Validate checks the configuration for errors.
func (c Config) Validate() error {
	if c.OutputFile == "" {
		return fmt.Errorf("output file is required")
	}
	if c.DataType != DataTypeASN && c.DataType != DataTypeCountry {
		return fmt.Errorf("invalid data type %q: must be 'asn' or 'country'", c.DataType)
	}
	if c.IPVersion != IPVersionBoth && c.IPVersion != IPVersion4 && c.IPVersion != IPVersion6 {
		return fmt.Errorf("invalid ip-version %q: must be '4', '6', or 'both'", c.IPVersion)
	}
	if c.Download && c.SourceFile != "" {
		return fmt.Errorf("cannot use both --source and --download")
	}
	if !c.Download && c.SourceFile == "" {
		return fmt.Errorf("must specify --source or --download")
	}
	return nil
}

// Build downloads data (if needed) and creates the MMDB file.
func Build(cfg Config) error {
	totalStart := time.Now()

	ipVersion := 6 // Default to IPv6 (supports both v4 and v6)
	if cfg.IPVersion == IPVersion4 {
		ipVersion = 4
	}

	writer, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            databaseType(cfg.DataType),
		RecordSize:              24,
		IPVersion:               ipVersion,
		IncludeReservedNetworks: true,
		DisableIPv4Aliasing:     true,
	})
	if err != nil {
		return fmt.Errorf("creating mmdb writer: %w", err)
	}

	var sources []string
	if cfg.Download {
		sources = cfg.SourceURLs()
	} else {
		sources = []string{cfg.SourceFile}
	}

	for _, src := range sources {
		if err := processSource(writer, src, cfg); err != nil {
			return err
		}
	}

	start := time.Now()

	// Write to temp file first for atomic replacement
	dir := filepath.Dir(cfg.OutputFile)
	tmpFile, err := os.CreateTemp(dir, ".iptoasn-*.mmdb.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	written, writeErr := writer.WriteTo(tmpFile)
	closeErr := tmpFile.Close()

	if writeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("writing mmdb: %w", writeErr)
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("closing temp file: %w", closeErr)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, cfg.OutputFile); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming temp file: %w", err)
	}

	slog.Info("Wrote MMDB", "file", cfg.OutputFile, "size_mb", fmt.Sprintf("%.1f", float64(written)/(1024*1024)), "duration", time.Since(start).Round(time.Millisecond))
	slog.Info("Build complete", "total_duration", time.Since(totalStart).Round(time.Millisecond))
	return nil
}

func databaseType(dt DataType) string {
	if dt == DataTypeCountry {
		return "iptoasn-country"
	}
	return "iptoasn"
}

func processSource(writer *mmdbwriter.Tree, src string, cfg Config) error {
	var r io.ReadCloser
	var err error

	if cfg.Download {
		r, err = Download(src)
	} else {
		r, err = OpenFile(src)
	}
	if err != nil {
		return err
	}
	defer r.Close()

	if cfg.DataType == DataTypeCountry {
		return processCountryRecords(writer, r, cfg.Download)
	}
	return processASNRecords(writer, r, cfg.Download)
}

func processASNRecords(writer *mmdbwriter.Tree, r io.Reader, validateMinimum bool) error {
	records, err := ParseASNRecords(r)
	if err != nil {
		return err
	}

	if validateMinimum && len(records) < minASNRecords {
		return fmt.Errorf("too few records: got %d, expected at least %d (data may be truncated or corrupt)", len(records), minASNRecords)
	}

	start := time.Now()
	inserted := 0
	for _, rec := range records {
		prefixes, err := ipRangeToPrefixes(rec.StartIP, rec.EndIP)
		if err != nil {
			return fmt.Errorf("converting range to prefixes: %w", err)
		}
		record := mmdbtype.Map{
			"autonomous_system_number":       mmdbtype.Uint32(rec.ASN),
			"autonomous_system_organization": mmdbtype.String(rec.Organization),
			"country": mmdbtype.Map{
				"iso_code": mmdbtype.String(rec.Country),
			},
		}

		for _, prefix := range prefixes {
			if err := writer.Insert(prefixToIPNet(prefix), record); err != nil {
				return fmt.Errorf("inserting record: %w", err)
			}
			inserted++
		}
	}

	slog.Info("Inserted prefixes", "count", inserted, "duration", time.Since(start).Round(time.Millisecond))
	return nil
}

func processCountryRecords(writer *mmdbwriter.Tree, r io.Reader, validateMinimum bool) error {
	records, err := ParseCountryRecords(r)
	if err != nil {
		return err
	}

	if validateMinimum && len(records) < minCountryRecords {
		return fmt.Errorf("too few records: got %d, expected at least %d (data may be truncated or corrupt)", len(records), minCountryRecords)
	}

	start := time.Now()
	inserted := 0
	for _, rec := range records {
		prefixes, err := ipRangeToPrefixes(rec.StartIP, rec.EndIP)
		if err != nil {
			return fmt.Errorf("converting range to prefixes: %w", err)
		}
		record := mmdbtype.Map{
			"country": mmdbtype.Map{
				"iso_code": mmdbtype.String(rec.Country),
			},
		}

		for _, prefix := range prefixes {
			if err := writer.Insert(prefixToIPNet(prefix), record); err != nil {
				return fmt.Errorf("inserting record: %w", err)
			}
			inserted++
		}
	}

	slog.Info("Inserted prefixes", "count", inserted, "duration", time.Since(start).Round(time.Millisecond))
	return nil
}
