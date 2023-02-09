#!/bin/bash
set -e

THISDIR="$( cd "$(dirname "$0")" ; pwd -P )"
PATH_TO_SYMLINKS="$THISDIR/../bin"


# Ensure we're running as root
if [ $(id -u) != "0" ]
then
	sudo "$0" "$@"
	exit $?
fi

# Get the current logged-in user from the owner of /dev/console
LOGGED_IN_USER=$(stat -f "%Su" /dev/console)
echo "Stopping the web server daemon for user '$LOGGED_IN_USER'..."
sudo -u $LOGGED_IN_USER \
    "$THISDIR/bin/git-bundle-server" web-server stop --remove

# Remove symlinks
for program in "$THISDIR"/bin/*
do
    symlink="$PATH_TO_SYMLINKS/$(basename $program)"
    if [ -L "$symlink" ]
    then
        echo "Deleting '$symlink'..."
	    rm -f "$symlink"
    else
        echo "No symlink found at path '$symlink'."
    fi
done

# Remove application files
if [ -d "$THISDIR" ]
then
	echo "Deleting application files in '$THISDIR'..."
	rm -rf "$THISDIR"
else
	echo "No application files found."
fi

# Forget package installation/delete receipt
echo "Removing installation receipt..."
pkgutil --forget com.github.gitbundleserver || echo "Could not remove package receipt. Exiting..."
exit 0
