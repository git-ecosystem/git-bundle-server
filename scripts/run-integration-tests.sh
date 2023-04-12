#!/bin/bash

THISDIR="$( cd "$(dirname "$0")" ; pwd -P )"
TESTDIR="$THISDIR/../test/integration"

# Defaults
ARGS=()

# Parse script arguments
for i in "$@"
do
case "$i" in
	--offline)
	ARGS+=("-p" "offline")
	shift # past argument
	;;
	--ci)
	ARGS+=("-p" "ci")
	shift # past argument
	;;
	*)
	die "unknown option '$i'"
	;;
esac
done

# Exit as soon as any line fails
set -e

cd "$TESTDIR"

npm install
npm run integration-test -- ${ARGS:+${ARGS[*]}}
