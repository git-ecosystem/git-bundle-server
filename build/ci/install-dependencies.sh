#!/bin/bash
die () {
	echo "$*" >&2
	exit 1
}

# Exit as soon as any line fails
set -e

# Install latest version of Git (minimum v2.40)
if command -v apt >/dev/null 2>&1; then
	sudo add-apt-repository ppa:git-core/ppa
	sudo apt update
	sudo apt -q -y install git
elif command -v brew >/dev/null 2>&1; then
	brew install git
else
	die 'Cannot install git'
fi

# Set up test Git config
git config --global user.name "GitHub Action"
git config --global user.email "action@github.com"
