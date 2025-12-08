package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with download",
			cfg: Config{
				OutputFile: "test.mmdb",
				DataType:   DataTypeASN,
				IPVersion:  IPVersionBoth,
				Download:   true,
			},
			wantErr: false,
		},
		{
			name: "valid config with source",
			cfg: Config{
				OutputFile: "test.mmdb",
				DataType:   DataTypeASN,
				IPVersion:  IPVersionBoth,
				SourceFile: "data.tsv",
			},
			wantErr: false,
		},
		{
			name: "missing output file",
			cfg: Config{
				DataType:  DataTypeASN,
				IPVersion: IPVersionBoth,
				Download:  true,
			},
			wantErr: true,
			errMsg:  "output file is required",
		},
		{
			name: "invalid data type",
			cfg: Config{
				OutputFile: "test.mmdb",
				DataType:   "invalid",
				IPVersion:  IPVersionBoth,
				Download:   true,
			},
			wantErr: true,
			errMsg:  "invalid data type",
		},
		{
			name: "invalid ip version",
			cfg: Config{
				OutputFile: "test.mmdb",
				DataType:   DataTypeASN,
				IPVersion:  "invalid",
				Download:   true,
			},
			wantErr: true,
			errMsg:  "invalid ip-version",
		},
		{
			name: "both download and source",
			cfg: Config{
				OutputFile: "test.mmdb",
				DataType:   DataTypeASN,
				IPVersion:  IPVersionBoth,
				Download:   true,
				SourceFile: "data.tsv",
			},
			wantErr: true,
			errMsg:  "cannot use both",
		},
		{
			name: "neither download nor source",
			cfg: Config{
				OutputFile: "test.mmdb",
				DataType:   DataTypeASN,
				IPVersion:  IPVersionBoth,
			},
			wantErr: true,
			errMsg:  "must specify",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestBuild_WithReservedNetworks(t *testing.T) {
	// Test that we can build an MMDB with reserved networks (e.g., 10.0.0.0/8, 192.168.0.0/16)
	// The testdata/ip2asn-sample.tsv includes 192.168.0.0-192.168.255.255
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test.mmdb")

	cfg := Config{
		OutputFile: outputFile,
		DataType:   DataTypeASN,
		IPVersion:  IPVersion4,
		SourceFile: "testdata/ip2asn-combined-sample.tsv",
	}

	err := Build(cfg)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Verify the file was created
	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestBuild_WithAliasedNetworks(t *testing.T) {
	// Test that we can build an MMDB with aliased networks (e.g., 2001::/32 Teredo)
	// The testdata/ip2asn-v6-sample.tsv includes 2001::/32
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "test-v6.mmdb")

	cfg := Config{
		OutputFile: outputFile,
		DataType:   DataTypeASN,
		IPVersion:  IPVersion6,
		SourceFile: "testdata/ip2asn-v6-sample.tsv",
	}

	err := Build(cfg)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}

func TestBuild_CountryData(t *testing.T) {
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "country.mmdb")

	cfg := Config{
		OutputFile: outputFile,
		DataType:   DataTypeCountry,
		IPVersion:  IPVersion4,
		SourceFile: "testdata/ip2country-sample.tsv",
	}

	err := Build(cfg)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	info, err := os.Stat(outputFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("output file is empty")
	}
}
