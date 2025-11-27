package internal

import (
	"fmt"
	"io"
	"os"
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
	f, err := os.Create(cfg.OutputFile)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	written, err := writer.WriteTo(f)
	if err != nil {
		return fmt.Errorf("writing mmdb: %w", err)
	}

	fmt.Printf("Wrote %s (%s) in %s\n", FormatFile(cfg.OutputFile), FormatMB(written), FormatDuration(time.Since(start)))
	fmt.Printf("Total time: %s\n", FormatDuration(time.Since(totalStart)))
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
		return processCountryRecords(writer, r)
	}
	return processASNRecords(writer, r)
}

func processASNRecords(writer *mmdbwriter.Tree, r io.Reader) error {
	records, err := ParseASNRecords(r)
	if err != nil {
		return err
	}

	start := time.Now()
	inserted := 0
	for _, rec := range records {
		prefixes := ipRangeToPrefixes(rec.StartIP, rec.EndIP)
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

	fmt.Printf("Inserted %s prefixes in %s\n", FormatNumber(inserted), FormatDuration(time.Since(start)))
	return nil
}

func processCountryRecords(writer *mmdbwriter.Tree, r io.Reader) error {
	records, err := ParseCountryRecords(r)
	if err != nil {
		return err
	}

	start := time.Now()
	inserted := 0
	for _, rec := range records {
		prefixes := ipRangeToPrefixes(rec.StartIP, rec.EndIP)
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

	fmt.Printf("Inserted %s prefixes in %s\n", FormatNumber(inserted), FormatDuration(time.Since(start)))
	return nil
}
