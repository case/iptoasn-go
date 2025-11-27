package internal

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const baseURL = "https://iptoasn.com/data"

// joinURL joins a base URL with path elements.
func joinURL(base string, elem ...string) string {
	u, _ := url.JoinPath(base, elem...)
	return u
}

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
			return []string{joinURL(baseURL, FileCountry4)}
		case IPVersion6:
			return []string{joinURL(baseURL, FileCountry6)}
		default:
			return []string{joinURL(baseURL, FileCountry4), joinURL(baseURL, FileCountry6)}
		}
	}

	// ASN data
	switch c.IPVersion {
	case IPVersion4:
		return []string{joinURL(baseURL, FileASN4)}
	case IPVersion6:
		return []string{joinURL(baseURL, FileASN6)}
	default:
		return []string{joinURL(baseURL, FileASNCombined)}
	}
}

// Download fetches a URL and returns a reader for the uncompressed content.
// The caller is responsible for closing the returned ReadCloser.
func Download(urlStr string) (io.ReadCloser, error) {
	slog.Info("Downloading", "url", urlStr)

	start := time.Now()
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, fmt.Errorf("downloading %s: %w", urlStr, err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("downloading %s: status %d", urlStr, resp.StatusCode)
	}

	// Validate Content-Type for gzip downloads
	if strings.HasSuffix(urlStr, ".gz") {
		contentType := resp.Header.Get("Content-Type")
		validTypes := []string{"application/gzip", "application/x-gzip", "application/octet-stream", "application/binary"}
		valid := false
		for _, t := range validTypes {
			if strings.HasPrefix(contentType, t) {
				valid = true
				break
			}
		}
		if !valid {
			resp.Body.Close()
			return nil, fmt.Errorf("downloading %s: unexpected content type %q (expected gzip)", urlStr, contentType)
		}
	}

	size := resp.ContentLength
	if size > 0 {
		slog.Info("Downloaded", "size_mb", fmt.Sprintf("%.1f", float64(size)/(1024*1024)), "duration", time.Since(start).Round(time.Millisecond))
	} else {
		slog.Info("Downloaded", "duration", time.Since(start).Round(time.Millisecond))
	}

	if strings.HasSuffix(urlStr, ".gz") {
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
	slog.Info("Reading", "path", path)

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
