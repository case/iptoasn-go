package internal

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const baseURL = "https://iptoasn.com/data"

// Supported source filenames from iptoasn.com.
const (
	FileASN4        = "ip2asn-v4.tsv.gz"
	FileASN6        = "ip2asn-v6.tsv.gz"
	FileASNCombined = "ip2asn-combined.tsv.gz"
	FileCountry4    = "ip2country-v4.tsv.gz"
	FileCountry6    = "ip2country-v6.tsv.gz"
)

// SourceURLs returns the URLs to download based on configuration.
func (c Config) SourceURLs() []string {
	if c.DataType == DataTypeCountry {
		switch c.IPVersion {
		case IPVersion4:
			return []string{baseURL + "/" + FileCountry4}
		case IPVersion6:
			return []string{baseURL + "/" + FileCountry6}
		default:
			return []string{baseURL + "/" + FileCountry4, baseURL + "/" + FileCountry6}
		}
	}

	// ASN data
	switch c.IPVersion {
	case IPVersion4:
		return []string{baseURL + "/" + FileASN4}
	case IPVersion6:
		return []string{baseURL + "/" + FileASN6}
	default:
		return []string{baseURL + "/" + FileASNCombined}
	}
}

// Download fetches a URL and returns a reader for the uncompressed content.
// The caller is responsible for closing the returned ReadCloser.
func Download(url string) (io.ReadCloser, error) {
	fmt.Printf("Downloading %s\n", url)

	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", url, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("downloading %s: status %d", url, resp.StatusCode)
	}

	size := resp.ContentLength
	if size > 0 {
		fmt.Printf("Downloaded %s in %s\n", FormatMB(size), FormatDuration(time.Since(start)))
	} else {
		fmt.Printf("Downloaded in %s\n", FormatDuration(time.Since(start)))
	}

	if strings.HasSuffix(url, ".gz") {
		gr, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("creating gzip reader: %w", err)
		}
		return &gzipReadCloser{gr, resp.Body}, nil
	}

	return resp.Body, nil
}

// OpenFile opens a local file and returns a reader for the uncompressed content.
// The caller is responsible for closing the returned ReadCloser.
func OpenFile(path string) (io.ReadCloser, error) {
	fmt.Printf("Reading %s\n", path)

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}

	if strings.HasSuffix(path, ".gz") {
		gr, err := gzip.NewReader(f)
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("creating gzip reader: %w", err)
		}
		return &gzipReadCloser{gr, f}, nil
	}

	return f, nil
}

// gzipReadCloser wraps a gzip.Reader to close both it and the underlying reader.
type gzipReadCloser struct {
	*gzip.Reader
	underlying io.Closer
}

func (g *gzipReadCloser) Close() error {
	return errors.Join(g.Reader.Close(), g.underlying.Close())
}
