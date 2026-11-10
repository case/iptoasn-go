# iptoasn-go

This is a utility for efficiently using the [iptoasn.com GeoIP data](https://iptoasn.com/) in Go programs, by converting the raw `.tsv` data to the [mmdb format](https://maxmind.github.io/MaxMind-DB/) (MaxMind DB File Format). It uses the [`maxmind/mmdbwriter`](https://github.com/maxmind/mmdbwriter) library for the `mmdb` conversion.

## Usage

1. Installation - `go install github.com/case/iptoasn-go/cmd/iptoasn@v1.0.0`
2. Usage - `iptoasn --download -o iptoasn.mmdb` - downloads and converts the default `ip2asn-combined.tsv.gz` file
3. Use the resulting `.mmdb` file in your Go programs, via the [`oschwald/geoip2-golang`](https://github.com/oschwald/geoip2-golang) library

## Misc

Minimum Go version supported - `1.26`

IPtoASN supported files:

- `ip2asn-combined.tsv.gz`
- `ip2asn-v4.tsv.gz`
- `ip2asn-v6.tsv.gz`
- `ip2country-v4.tsv.gz`
- `ip2country-v6.tsv.gz`

## Development

- `bin/setup` - runs `mise install` if mise is present, downloads Go modules, and checks for the lint tools pinned in `mise.toml`
- `bin/lint` - runs `golangci-lint` (including `gofmt`), `jactionlint`, and `shellcheck`. Fix formatting with `golangci-lint fmt`
- `bin/test` - runs the tests with the race detector
- `bin/ci` - runs all of the above, as CI does

## Thank you 🙏

- [Frank Denis](https://github.com/jedisct1/), for providing the `iptoasn.com` data service
- [MaxMind](https://github.com/maxmind/), for the `mmdb` file format, and the read & write Go code
- [Julia Evans](https://jvns.ca/blog/2024/10/27/asn-ip-address-memory/), for her explorations in the space
