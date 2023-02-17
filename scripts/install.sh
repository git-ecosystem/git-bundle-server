#!/bin/bash
die () {
	echo "$*" >&2
	exit 1
}

# Setup root escalation operation
SUDO=""
retry_root () {
	if [ -n "$SUDO" ]
	then
		# run as user, then with 'sudo'
		$@ 2>/dev/null || "$SUDO" $@
	else
		# passthrough
		$@
	fi
}

# Parse script arguments
for i in "$@"
do
case "$i" in
	--bindir=*)
	BINDIR="${i#*=}"
	shift # past argument=value
	;;
	--docdir=*)
	DOCDIR="${i#*=}"
	shift # past argument=value
	;;
	--uninstaller=*)
	UNINSTALLER="${i#*=}"
	shift # past argument=value
	;;
	--allow-root)
	SUDO=$(if command -v sudo >/dev/null 2>&1; then echo sudo; fi)
	shift # past argument
	;;
	--include-symlinks)
	INCLUDE_SYMLINKS=1
	shift # past argument
	;;
	--install-root=*)
	INSTALL_ROOT="${i#*=}"
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
if [ -z "$DOCDIR" ]; then
	die "--docdir was not set"
fi
if [ -z "$INSTALL_ROOT" ]; then
	die "--install-root was not set"
fi

if [ "$INSTALL_ROOT" == "/" ]; then
	# Reset $INSTALL_ROOT to empty string to avoid double leading slash
	INSTALL_ROOT=""
fi

# Exit as soon as any line fails
set -e

# Ensure payload directory exists
APP_ROOT="$INSTALL_ROOT/usr/local/git-bundle-server"
retry_root mkdir -p "$APP_ROOT"

# Copy built binaries
echo "Copying binaries..."
retry_root cp -R "$BINDIR/." "$APP_ROOT/bin"

echo "Copying manpages..."
for N in $(find "$DOCDIR" -type f | sed -e 's/.*\.//' | sort -u)
do
	retry_root mkdir -p "$APP_ROOT/share/man/man$N"
	retry_root cp -R "$DOCDIR/"*."$N" "$APP_ROOT/share/man/man$N"

done

# Copy uninstaller script
if [ -n "$UNINSTALLER" ]; then
	echo "Copying uninstall script..."
	retry_root cp "$UNINSTALLER" "$APP_ROOT"
fi

# Create symlinks
if [ -n "$INCLUDE_SYMLINKS" ]; then
	LINK_TO="$INSTALL_ROOT/usr/local/bin"
	RELATIVE_LINK_TO_BIN="../git-bundle-server/bin"
	retry_root mkdir -p "$LINK_TO"

	echo "Creating binary symlinks..."
	for program in "$APP_ROOT"/bin/*
	do
		p=$(basename "$program")
		retry_root rm -f "$LINK_TO/$p"
		retry_root ln -s "$RELATIVE_LINK_TO_BIN/$p" "$LINK_TO/$p"
	done

	echo "Creating manpage symlinks..."
	for mandir in "$APP_ROOT"/share/man/man*/
	do
		mdir=$(basename "$mandir")
		LINK_TO="$INSTALL_ROOT/usr/local/share/man/$mdir"
		RELATIVE_LINK_TO_MAN="../../../git-bundle-server/share/man/$mdir"
		retry_root mkdir -p "$LINK_TO"

		for manpage in "$mandir"/*
		do
			mpage=$(basename "$manpage")
			retry_root rm -f "$LINK_TO/$mpage"
			retry_root ln -s "$RELATIVE_LINK_TO_MAN/$mpage" "$LINK_TO/$mpage"
		done
	done
fi
