.PHONY: install
install:
	go install ./cmd/iptoasn

.PHONY: test
test:
	go test ./...
