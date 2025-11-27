// Command iptoasn downloads IP-to-ASN data from iptoasn.com and builds MMDB files.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/case/iptoasn-go/internal"
)

var (
	outputFile = flag.String("o", "iptoasn.mmdb", "Output MMDB path and file")
	dataType   = flag.String("type", "asn", "Data type: 'asn' or 'country'")
	ipVersion  = flag.String("ip-version", "both", "IP version: '4', '6', or 'both'")
	sourceFile = flag.String("source", "", "Local TSV/TSV.gz file")
	download   = flag.Bool("download", false, "Download latest data from iptoasn.com")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: iptoasn [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Builds an MMDB file from iptoasn.com data.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  iptoasn --download -o iptoasn.mmdb\n")
		fmt.Fprintf(os.Stderr, "  iptoasn --download --type country -o country.mmdb\n")
		fmt.Fprintf(os.Stderr, "  iptoasn --download --ip-version 4 -o iptoasn-v4.mmdb\n")
		fmt.Fprintf(os.Stderr, "  iptoasn --download --ip-version 6 -o iptoasn-v6.mmdb\n")
		fmt.Fprintf(os.Stderr, "  iptoasn --source ip2asn-combined.tsv.gz -o iptoasn.mmdb\n")
	}
	flag.Parse()

	// Show help if no flags provided
	if !*download && *sourceFile == "" {
		flag.Usage()
		os.Exit(0)
	}

	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg := internal.Config{
		OutputFile: *outputFile,
		DataType:   internal.DataType(*dataType),
		IPVersion:  internal.IPVersion(*ipVersion),
		SourceFile: *sourceFile,
		Download:   *download,
	}

	if err := cfg.Validate(); err != nil {
		return err
	}

	return internal.Build(cfg)
}
