# Default target
build:

# Helpful paths
BINDIR := $(CURDIR)/bin

# Platform information
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Build targets
.PHONY: build
build:
	$(RM) -r $(BINDIR)
	@mkdir -p $(BINDIR)
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build -o $(BINDIR) ./...

# Cleanup targets
.PHONY: clean
clean:
	go clean ./...
	$(RM) -r $(BINDIR)
