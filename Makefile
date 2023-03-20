# Default target
build:

# Project metadata (note: to package, VERSION *must* be set by the caller)
NAME := git-bundle-server
VERSION :=
PACKAGE_REVISION := 1

# Installation information
INSTALL_ROOT := /

# Helpful paths
BINDIR := $(CURDIR)/bin
DISTDIR := $(CURDIR)/_dist
DOCDIR := $(CURDIR)/_docs
TESTDIR := $(CURDIR)/_test

# Platform information
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# Packaging information
SUPPORTED_PACKAGE_GOARCHES := amd64 arm64
PACKAGE_ARCH := $(GOARCH)

# Guard against environment variables
APPLE_APP_IDENTITY =
APPLE_INST_IDENTITY =
APPLE_KEYCHAIN_PROFILE =
E2E_FLAGS=

# Build targets
.PHONY: build
build:
	$(RM) -r $(BINDIR)
	@mkdir -p $(BINDIR)
	GOOS="$(GOOS)" GOARCH="$(GOARCH)" go build -o $(BINDIR) ./...

.PHONY: doc
doc:
	@scripts/make-docs.sh --docs="$(CURDIR)/docs/man" \
			      --output="$(DOCDIR)"

# Testing targets
.PHONY: e2e-test
e2e-test: build
	@echo
	@echo "======== Running end-to-end tests ========"
	$(RM) -r $(TESTDIR)
	@scripts/run-e2e-tests.sh $(E2E_FLAGS)

# Installation targets
.PHONY: install
install: build doc
	@echo
	@echo "======== Installing to $(INSTALL_ROOT) ========"
	@scripts/install.sh --bindir="$(BINDIR)" \
			    --docdir="$(DOCDIR)" \
			    --uninstaller="$(CURDIR)/scripts/uninstall.sh" \
			    --allow-root \
			    --include-symlinks \
			    --install-root="$(INSTALL_ROOT)"

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
$(DEBDIR)/root: check-arch build doc
	@echo
	@echo "======== Formatting package contents ========"
	@scripts/install.sh --bindir="$(BINDIR)" \
			    --docdir="$(DOCDIR)" \
			    --include-symlinks \
			    --install-root="$(DEBDIR)/root"

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
#   2. (Optional) Codesign the package contents in place.
#   3. Create the product archive in _dist/.

# Platform-specific variables
PKGDIR := $(DISTDIR)/pkg
PKG_FILENAME := $(DISTDIR)/$(NAME)_$(VERSION)-$(PACKAGE_REVISION)_$(PACKAGE_ARCH).pkg

# Targets
$(PKGDIR)/payload: check-arch build doc
	@echo
	@echo "======== Formatting package contents ========"
	@scripts/install.sh --bindir="$(BINDIR)" \
			    --docdir="$(DOCDIR)" \
			    --uninstaller="$(CURDIR)/scripts/uninstall.sh" \
			    --install-root="$(PKGDIR)/payload"

ifdef APPLE_APP_IDENTITY
.PHONY: codesign
codesign: $(PKGDIR)/payload
	@echo
	@echo "======== Codesigning package contents ========"
	@build/package/pkg/codesign.sh --payload="$(PKGDIR)/payload" \
				       --identity="$(APPLE_APP_IDENTITY)" \
				       --entitlements="$(CURDIR)/build/package/pkg/entitlements.xml"

$(PKG_FILENAME): codesign
endif

$(PKG_FILENAME): check-version $(PKGDIR)/payload
	@echo
	@echo "======== Creating product archive package ========"
	@build/package/pkg/pack.sh --version="$(VERSION)" \
				   --payload="$(PKGDIR)/payload" \
				   --identity="$(APPLE_INST_IDENTITY)" \
				   --output="$(PKG_FILENAME)"

# Notarization can only happen if the package is fully signed
ifdef APPLE_APP_IDENTITY
ifdef APPLE_INST_IDENTITY
ifdef APPLE_KEYCHAIN_PROFILE
.PHONY: notarize
notarize: $(PKG_FILENAME)
	@echo
	@echo "======== Notarizing package ========"
	@build/package/pkg/notarize.sh --package="$(PKG_FILENAME)" \
				       --keychain-profile="$(APPLE_KEYCHAIN_PROFILE)"

package: notarize
endif
endif
endif

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
	$(RM) -r $(DOCDIR)
	$(RM) -r $(TESTDIR)
