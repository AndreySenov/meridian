GO ?= go
GOLANGCI_LINT ?= golangci-lint
GOIMPORTS ?= goimports
RM ?= rm

.PHONY: clean
clean:
	$(RM) -f coverage.out coverage.html

.PHONY: tidy
tidy:
	$(GO) mod tidy

.PHONY: tidy-check
tidy-check:
	$(GO) mod tidy -diff

.PHONY: fmt
fmt:
	$(GO) fmt ./...
	$(GOIMPORTS) -w .

.PHONY: fmt-check
fmt-check:
	@files="$$($(GOIMPORTS) -l .)"; \
	if [ -n "$$files" ]; then \
		echo "Files are not formatted (run 'make fmt'):"; \
		echo "$$files"; \
		exit 1; \
	fi

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run ./...

.PHONY: test
test:
	$(GO) test -race -count=1 ./...

.PHONY: cover
cover:
	$(GO) test -race -count=1 -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html

.PHONY: tools
tools:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install golang.org/x/tools/cmd/goimports@latest

.PHONY: check
check: tidy-check fmt-check vet lint test
