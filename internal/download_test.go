package internal

import (
	"testing"
)

func TestSourceURLs_ASN(t *testing.T) {
	tests := []struct {
		name      string
		dataType  DataType
		ipVersion IPVersion
		wantURLs  []string
	}{
		{
			name:      "ASN both versions",
			dataType:  DataTypeASN,
			ipVersion: IPVersionBoth,
			wantURLs:  []string{baseURL + "/" + FileASNCombined},
		},
		{
			name:      "ASN v4 only",
			dataType:  DataTypeASN,
			ipVersion: IPVersion4,
			wantURLs:  []string{baseURL + "/" + FileASN4},
		},
		{
			name:      "ASN v6 only",
			dataType:  DataTypeASN,
			ipVersion: IPVersion6,
			wantURLs:  []string{baseURL + "/" + FileASN6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				DataType:  tt.dataType,
				IPVersion: tt.ipVersion,
			}

			urls := cfg.SourceURLs()

			if len(urls) != len(tt.wantURLs) {
				t.Fatalf("expected %d URLs, got %d", len(tt.wantURLs), len(urls))
			}

			for i, want := range tt.wantURLs {
				if urls[i] != want {
					t.Errorf("URL[%d] = %s, want %s", i, urls[i], want)
				}
			}
		})
	}
}

func TestSourceURLs_Country(t *testing.T) {
	tests := []struct {
		name      string
		ipVersion IPVersion
		wantURLs  []string
	}{
		{
			name:      "Country both versions",
			ipVersion: IPVersionBoth,
			wantURLs:  []string{baseURL + "/" + FileCountry4, baseURL + "/" + FileCountry6},
		},
		{
			name:      "Country v4 only",
			ipVersion: IPVersion4,
			wantURLs:  []string{baseURL + "/" + FileCountry4},
		},
		{
			name:      "Country v6 only",
			ipVersion: IPVersion6,
			wantURLs:  []string{baseURL + "/" + FileCountry6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				DataType:  DataTypeCountry,
				IPVersion: tt.ipVersion,
			}

			urls := cfg.SourceURLs()

			if len(urls) != len(tt.wantURLs) {
				t.Fatalf("expected %d URLs, got %d", len(tt.wantURLs), len(urls))
			}

			for i, want := range tt.wantURLs {
				if urls[i] != want {
					t.Errorf("URL[%d] = %s, want %s", i, urls[i], want)
				}
			}
		})
	}
}

func TestOpenFile(t *testing.T) {
	r, err := OpenFile("testdata/ip2asn-combined-sample.tsv")
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer r.Close()

	// Read a bit to confirm it works
	buf := make([]byte, 100)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n == 0 {
		t.Error("expected to read some bytes")
	}
}

func TestOpenFile_NotFound(t *testing.T) {
	_, err := OpenFile("testdata/nonexistent.tsv")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
