# Default target
build:

# Project metadata (note: to package, VERSION *must* be set by the caller)
NAME := git-bundle-server
VERSION :=
PACKAGE_REVISION := 1

# Helpful paths
BINDIR := $(CURDIR)/bin
DISTDIR := $(CURDIR)/_dist

# Platform information
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Packaging information
SUPPORTED_PACKAGE_GOARCHES := amd64 arm64
PACKAGE_ARCH := $(GOARCH)

# Build targets
.PHONY: build
build:
	$(RM) -r $(BINDIR)
	@mkdir -p $(BINDIR)
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build -o $(BINDIR) ./...

# Packaging targets
.PHONY: check-arch
check-arch:
	$(if $(filter $(GOARCH),$(SUPPORTED_PACKAGE_GOARCHES)), , \
		$(error cannot create package for GOARCH "$(GOARCH)"; \
			supported architectures are: $(SUPPORTED_PACKAGE_GOARCHES)))

.PHONY: check-version
check-version:
	$(if $(VERSION), , $(error version is undefined))

ifeq ($(GOOS),linux)
# Linux binary .deb file
# Steps:
#   1. Layout files in _dist/deb/root/ as they'll be installed (unlike MacOS
#      .pkg packages, symlinks created in the payload are preserved, so we
#      create them here to avoid doing so in a post-install step).
#   2. Create the binary deb package in _dist/deb/.

# Platform-specific variables
DEBDIR := $(DISTDIR)/deb
DEB_FILENAME := $(DISTDIR)/$(NAME)_$(VERSION)-$(PACKAGE_REVISION)_$(PACKAGE_ARCH).deb

# Targets
$(DEBDIR)/root: check-arch build
	@echo
	@echo "======== Formatting package contents ========"
	@build/package/layout-unix.sh --bindir="$(BINDIR)" \
				      --include-symlinks \
				      --output="$(DEBDIR)/root"

$(DEB_FILENAME): check-version $(DEBDIR)/root
	@echo
	@echo "======== Creating binary Debian package ========"
	@build/package/deb/pack.sh --payload="$(DEBDIR)/root" \
				   --scripts="$(CURDIR)/build/package/deb/scripts" \
				   --arch="$(PACKAGE_ARCH)" \
				   --version="$(VERSION)" \
				   --output="$(DEB_FILENAME)"

.PHONY: package
package: $(DEB_FILENAME)

else ifeq ($(GOOS),darwin)
# MacOS .pkg file
# Steps:
#   1. Layout files in _dist/pkg/payload/ as they'll be installed (including
#      uninstall.sh script).
#   2. Create the product archive in _dist/.

# Platform-specific variables
PKGDIR := $(DISTDIR)/pkg
PKG_FILENAME := $(DISTDIR)/$(NAME)_$(VERSION)-$(PACKAGE_REVISION)_$(PACKAGE_ARCH).pkg

# Targets
$(PKGDIR)/payload: check-arch build
	@echo
	@echo "======== Formatting package contents ========"
	@build/package/layout-unix.sh --bindir="$(BINDIR)" \
				      --uninstaller="$(CURDIR)/scripts/uninstall.sh" \
				      --output="$(PKGDIR)/payload"

$(PKG_FILENAME): check-version $(PKGDIR)/payload
	@echo
	@echo "======== Creating product archive package ========"
	@build/package/pkg/pack.sh --version="$(VERSION)" \
				   --payload="$(PKGDIR)/payload" \
				   --output="$(PKG_FILENAME)"

.PHONY: package
package: $(PKG_FILENAME)

else
# Packaging not supported for platform, exit with error.
.PHONY: package
package:
	$(error cannot create package for GOOS "$(GOOS)")

endif

# Cleanup targets
.PHONY: clean
clean:
	go clean ./...
	$(RM) -r $(BINDIR)
	$(RM) -r $(DISTDIR)
