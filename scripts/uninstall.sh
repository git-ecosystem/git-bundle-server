#!/bin/bash
set -e

# Setup root escalation operation
SUDO=$(if command -v sudo >/dev/null 2>&1; then echo sudo; fi)
retry_root () {
	$@ 2>/dev/null || "$SUDO" $@
}

THISDIR="$( cd "$(dirname "$0")" ; pwd -P )"
PATH_TO_BIN_SYMLINKS="$THISDIR/../bin"
PATH_TO_MAN_SYMLINKS="$THISDIR/../share/man"

# Ensure we're running as root
if [ $(id -u) == "0" ]
then
	echo
	echo "WARNING: running this script as root will not remove user-scoped resources such"
	echo "as daemon configurations."
	echo
	read -p "Are you sure you want to proceed? (y/N) " response
	case $response in
		[yY]*)
		break # do nothing
		;;
		[nN]*|"")
		exit 0 # exit
		;;
		*)
		echo "Invalid response: $response"
		;;
	esac
fi

"$THISDIR/bin/git-bundle-server" web-server stop --remove

# Remove symlinks
for program in "$THISDIR"/bin/*
do
    symlink="$PATH_TO_BIN_SYMLINKS/$(basename $program)"
    if [ -L "$symlink" ]
    then
        echo "Deleting '$symlink'..."
	retry_root rm -f "$symlink"
    else
        echo "No symlink found at path '$symlink'."
    fi
done

# Remove application files
if [ -d "$THISDIR" ]
then
	echo "Deleting application files in '$THISDIR'..."
	retry_root rm -rf "$THISDIR"
else
	echo "No application files found."
fi

# If installed via MacOS .pkg, remove package receipt
PKG_ID=com.github.gitbundleserver
if command -v pkgutil >/dev/null 2>&1 && pkgutil --pkgs=$PKG_ID >/dev/null 2>&1
then
	# Must run as root
	$SUDO pkgutil --forget $PKG_ID
fi

exit 0
