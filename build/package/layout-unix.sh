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
INSTALL_TO="$PAYLOAD/usr/local/git-bundle-server"
mkdir -p "$INSTALL_TO"

# Copy built binaries
echo "Copying binaries..."
cp -R "$BINDIR/." "$INSTALL_TO/bin"

# Copy uninstaller script
if [ -n "$UNINSTALLER" ]; then
	echo "Copying uninstall script..."
	cp "$UNINSTALLER" "$INSTALL_TO"
fi

# Create symlinks
if [ -n "$INCLUDE_SYMLINKS" ]; then
	LINK_TO="$PAYLOAD/usr/local/bin"
	mkdir -p "$LINK_TO"

	echo "Creating symlinks..."
	for program in "$INSTALL_TO"/bin/*
	do
		ln -s -r "$program" "$LINK_TO/$(basename $program)"
	done
fi

echo "Layout complete."
