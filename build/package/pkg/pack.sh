#!/bin/bash
die () {
	echo "$*" >&2
	exit 1
}

# Directories
THISDIR="$( cd "$(dirname "$0")" ; pwd -P )"

# Product information
IDENTIFIER="com.github.gitbundleserver"
INSTALL_LOCATION="/usr/local/git-bundle-server"

# Defaults
IDENTITY=

# Parse script arguments
for i in "$@"
do
case "$i" in
	--version=*)
	VERSION="${i#*=}"
	shift # past argument=value
	;;
	--payload=*)
	PAYLOAD="${i#*=}"
	shift # past argument=value
	;;
	--identity=*)
	IDENTITY="${i#*=}"
	shift # past argument=value
	;;
	--output=*)
	PKGOUT="${i#*=}"
	shift # past argument=value
	;;
	*)
	die "unknown option '$i'"
	;;
esac
done

# Perform pre-execution checks
if [ -z "$VERSION" ]; then
	die "--version was not set"
fi
if [ -z "$PAYLOAD" ]; then
	die "--payload was not set"
elif [ ! -d "$PAYLOAD" ]; then
	die "Could not find '$PAYLOAD'. Did you run layout-unix.sh first?"
fi
if [ -z "$PKGOUT" ]; then
	die "--output was not set"
fi

# Exit as soon as any line fails
set -e

# Cleanup any old component
if [ -e "$PKGOUT" ]; then
	echo "Deleting old component '$PKGOUT'..."
	rm -f "$PKGOUT"
fi

# Ensure the parent directory for the component exists
mkdir -p "$(dirname "$PKGOUT")"

# Build the component package
PKGTMP="$PKGOUT.component"

# Remove any unwanted .DS_Store files
echo "Removing unnecessary files..."
find "$PAYLOAD" -name '*.DS_Store' -type f -delete

# Set full read, write, execute permissions for owner and just read and execute permissions for group and other
echo "Setting file permissions..."
/bin/chmod -R 755 "$PAYLOAD"

# Remove any extended attributes (ACEs)
echo "Removing extended attributes..."
/usr/bin/xattr -rc "$PAYLOAD"

# Build component package
echo "Building core component package..."
/usr/bin/pkgbuild \
	--root "$PAYLOAD/" \
	--install-location "/" \
	--scripts "$THISDIR/scripts" \
	--identifier "$IDENTIFIER" \
	--version "$VERSION" \
	"$PKGTMP"

echo "Component pack complete."

# Build product installer
echo "Building product package..."
/usr/bin/productbuild \
	--package "$PKGTMP" \
	--identifier "$IDENTIFIER" \
	--version "$VERSION" \
	${IDENTITY:+"--sign"} ${IDENTITY:+"$IDENTITY"} \
	"$PKGOUT"

echo "Product build complete."
