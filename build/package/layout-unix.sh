#!/bin/bash
die () {
	echo "$*" >&2
	exit 1
}

# Parse script arguments
for i in "$@"
do
case "$i" in
	--bindir=*)
	BINDIR="${i#*=}"
	shift # past argument=value
	;;
	--uninstaller=*)
	UNINSTALLER="${i#*=}"
	shift # past argument=value
	;;
	--include-symlinks)
	INCLUDE_SYMLINKS=1
	shift # past argument
	;;
	--output=*)
	PAYLOAD="${i#*=}"
	shift # past argument=value
	;;
	*)
	die "unknown option '$i'"
	;;
esac
done

# Perform pre-execution checks
if [ -z "$BINDIR" ]; then
	die "--bindir was not set"
fi
if [ -z "$PAYLOAD" ]; then
	die "--output was not set"
fi

# Exit as soon as any line fails
set -e

# Cleanup any old payload directory
if [ -d "$PAYLOAD" ]; then
	echo "Cleaning old output directory '$PAYLOAD'..."
	rm -rf "$PAYLOAD"
fi

# Ensure payload directory exists
APP_ROOT="$PAYLOAD/usr/local/git-bundle-server"
mkdir -p "$APP_ROOT"

# Copy built binaries
echo "Copying binaries..."
cp -R "$BINDIR/." "$APP_ROOT/bin"

# Copy uninstaller script
if [ -n "$UNINSTALLER" ]; then
	echo "Copying uninstall script..."
	cp "$UNINSTALLER" "$APP_ROOT"
fi

# Create symlinks
if [ -n "$INCLUDE_SYMLINKS" ]; then
	LINK_TO="$PAYLOAD/usr/local/bin"
	RELATIVE_LINK_TO_BIN="../git-bundle-server/bin"
	mkdir -p "$LINK_TO"

	echo "Creating binary symlinks..."
	for program in "$APP_ROOT"/bin/*
	do
		p=$(basename "$program")
		rm -f "$LINK_TO/$p"
		ln -s "$RELATIVE_LINK_TO_BIN/$p" "$LINK_TO/$p"
	done
fi

echo "Layout complete."
