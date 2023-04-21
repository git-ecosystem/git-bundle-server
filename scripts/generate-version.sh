#!/bin/bash
die () {
	echo "$*" >&2
	exit 1
}

write_version () {
	if [ -n "$1" ]; then
		echo "Setting VERSION to $1"
	else
		echo "No version found!"
	fi
	echo "VERSION := $1" >"$2"
}

# Parse script arguments
for i in "$@"
do
case "$i" in
	--version-file=*)
	VERSION_FILE="${i#*=}"
	shift # past argument=value
	;;
	*)
	die "unknown option '$i'"
	;;
esac
done

if [ -z "$VERSION_FILE" ]; then
    die "error: missing version file path"
fi

set -e

# The pattern match used below is *less* restrictive than it should be (i.e.
# "v<number>.<number>.<number>[-<optional string>]"), but we're limited by the
# glob patterns used in 'git describe'.
#
# That said, we're explicitly specifying 'VERSION' for formal releases(based on
# the more precise pattern match in 'release.yml', so it's not super important
# to be strict here.
VERSION_RAW="$(git describe --tags --match "v[0-9]*.[0-9]*.[0-9]*" 2>/dev/null || echo "")"
VERSION="${VERSION_RAW:1}"

if [ -r "$VERSION_FILE" ]
then
	CACHE_VERSION=$(sed -e 's/^VERSION := //' <"$VERSION_FILE")
	if [ "$VERSION" != "$CACHE_VERSION" ]; then
		write_version "$VERSION" "$VERSION_FILE"
	fi
else
	# If the file doesn't exist, write unconditionally
	write_version "$VERSION" "$VERSION_FILE"
fi
