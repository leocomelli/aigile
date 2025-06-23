.PHONY: build test lint clean integration-test

# Build variables
BINARY_NAME=aigile
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell $(GO) env GOBIN))
GOBIN=$(shell $(GO) env GOPATH)/bin
else
GOBIN=$(shell $(GO) env GOBIN)
endif

all: build

build:
	$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) .

test:
	$(GO) test ./... -v -race -cover

# To run all integration tests (GitHub and Google Sheets), use:
# make integration-test GOOGLE_SHEET_ID=your_id GOOGLE_CREDENTIALS_FILE=your_credentials.json
integration-test:
	$(GO) test -v -race -tags=integration ./internal/provider ./internal/reader -run Integration

lint:
	export GOROOT=$(go env GOROOT) && golangci-lint run --config .golangci.yml ./...

clean:
	$(GO) clean
	rm -f $(BINARY_NAME)

tidy:
	$(GO) mod tidy

deps:
	$(GO) mod download

run:
	$(GO) run main.go 