#!/bin/bash
die () {
	echo "$*" >&2
	exit 1
}

# Directories
THISDIR="$( cd "$(dirname "$0")" ; pwd -P )"

# Local paths
UNINSTALLER="$THISDIR/uninstall.sh"

# Parse script arguments
for i in "$@"
do
case "$i" in
	--payload=*)
	DEBROOT="${i#*=}"
	shift # past argument=value
	;;
	--scripts=*)
	SCRIPT_DIR="${i#*=}"
	shift # past argument=value
	;;
	--arch=*)
	ARCH="${i#*=}"
	shift # past argument=value
	;;
	--version=*)
	VERSION="${i#*=}"
	shift # past argument=value
	;;
	--output=*)
	DEBOUT="${i#*=}"
	shift # past argument=value
	;;
	*)
	die "unknown option '$i'"
	;;
esac
done

# Perform pre-execution checks
if [ -z "$DEBROOT" ]; then
	die "--payload was not set"
elif [ ! -d "$DEBROOT" ]; then
	die "Could not find '$DEBROOT'. Did you run layout-unix.sh first?"
fi
if [ -z "$ARCH" ]; then
	die "--arch was not set"
fi
if [ -z "$VERSION" ]; then
	die "--version was not set"
fi
if [ -z "$DEBOUT" ]; then
	die "--output was not set"
fi

# Exit as soon as any line fails
set -e

# Cleanup old package
if [ -e "$DEBOUT" ]; then
	echo "Deleting old package '$DEBOUT'..."
	rm -f "$DEBOUT"
fi

CONTROLDIR="$DEBROOT/DEBIAN"

# Ensure the parent directory for the .deb exists
mkdir -p "$(dirname "$DEBOUT")"

# Build .deb
mkdir -m 755 -p "$CONTROLDIR"

# Create the debian control file
cat >"$CONTROLDIR/control" <<EOF
Package: git-bundle-server
Version: $VERSION
Section: vcs
Priority: optional
Architecture: $ARCH
Depends:
Maintainer: Git Bundle Server <git-fundamentals@github.com>
Description: A self-hostable Git bundle server.
EOF

# Copy the maintainer scripts, if they exist
if [ -d "$SCRIPT_DIR" ]; then
	cp -R  "$SCRIPT_DIR/." "$CONTROLDIR"
fi

dpkg-deb -Zxz --build "$DEBROOT" "$DEBOUT"
