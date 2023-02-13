#!/bin/bash
die () {
	echo "$*" >&2
	exit 1
}

THISDIR="$( cd "$(dirname "$0")" ; pwd -P )"

# Parse script arguments
for i in "$@"
do
case "$i" in
	--docs=*)
	MAN_DIR="${i#*=}"
	shift # past argument=value
	;;
	--output=*)
	OUT_DIR="${i#*=}"
	shift # past argument=value
	;;
	*)
	die "unknown option '$i'"
	;;
esac
done

# Perform pre-execution checks
if [ -z "$MAN_DIR" ]; then
	die "--docs was not set"
fi

if [ -z "$OUT_DIR" ]; then
	die "--output was not set"
fi

set -e

if ! command -v asciidoctor >/dev/null 2>&1
then
	die "cannot generate man pages: asciidoctor not installed. \
See https://docs.asciidoctor.org/asciidoctor/latest/install for \
installation instructions."
fi

# Generate the man pages
mkdir -p "$OUT_DIR"

asciidoctor -b manpage -I "$THISDIR" -r asciidoctor-extensions.rb -D "$OUT_DIR" "$MAN_DIR/*.adoc"
