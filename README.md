# Git Bundle Server

[bundle-uris]: https://github.com/git/git/blob/next/Documentation/technical/bundle-uri.txt
[codeowners]: CODEOWNERS
[contributing]: CONTRIBUTING.md
[license]: LICENSE
[support]: SUPPORT.md

## Background

By running this software, you can self-host a bundle server to work with Git's
[bundle URI feature][bundle-uris].

This repository is under active development, and loves contributions from the
community :heart:. Check out [CONTRIBUTING][contributing] for details on getting
started.

## Getting Started

### Installing

> :warning: Installation on Windows is currently unsupported :warning:

<!-- Common sources -->
[releases]: https://github.com/github/git-bundle-server/releases

#### Linux

Debian packages (for x86_64 systems) can be downloaded from the
[Releases][releases] page and installed with:

```bash
sudo dpkg -i /path/to/git-bundle-server_VVV-RRR_amd64.deb

# VVV: version
# RRR: package revision
```

#### MacOS

Packages for both Intel and M1 systems can be downloaded from the
[Releases][releases] page (identified by the `amd64` vs `arm64` filename suffix,
respectively). The package can be installed by double-clicking the downloaded
file, or on the command line with:

```bash
sudo installer -pkg /path/to/git-bundle-server_VVV-RRR_AAA.pkg -target /

# VVV: version
# RRR: package revision
# AAA: platform architecture (amd64 or arm64)
```

#### From source

> To avoid environment issues building and executing Go code, we recommend that
> you clone inside the `src` directory of your `GOROOT`.

You can also install the bundle server application from source on any Unix-based
system. To install to the system root, clone the repository and run:

```ShellSession
$ make install
```

Note that you will likely be prompted for a password to allow installing to
root-owned directories (e.g. `/usr/local/bin`).

To install somewhere other than the system root, you can manually specify an
`INSTALL_ROOT` when building the `install` target:

```ShellSession
$ make install INSTALL_ROOT=</your/install/root>
```

### Uninstalling

#### From Debian package

To uninstall `git-bundle-server` if it was installed from a Debian package, run:

```ShellSession
$ sudo dpkg -r git-bundle-server
```

#### Everything else

All other installation methods include an executable script that uninstalls all
bundle server resources. On MacOS & Linux, run:

```ShellSession
$ /usr/local/git-bundle-server/uninstall.sh
```

## Usage

### Repository management

The following command-line interface allows you to manage which repositories are
being managed by the bundle server.

* `git-bundle-server init <url> [<route>]`: Initialize a repository by cloning a
  bare repo from `<url>`. If `<route>` is specified, then it is the bundle
  server route to find the data for this repository. Otherwise, the route is
  inferred from `<url>` by removing the domain name. For example,
  `https://github.com/git-for-windows/git` is assigned the route
  `/git-for-windows/git`. Run `git-bundle-server update` to initialize bundle
  information. Configure the web server to recognize this repository at that
  route. Configure scheduler to run `git-bundle-server update-all` as
  necessary.

* `git-bundle-server update [--daily|--hourly] <route>`: For the
  repository in the current directory (or the one specified by `<route>`), fetch
  the latest content from the remote and create a new set of bundles and update
  the bundle list.  The `--daily` and `--hourly` options allow the scheduler to
  indicate the timing of this instance to indicate if the newest bundle should
  be an "hourly" or "daily" bundle. If `--daily` is specified, then collapse the
  existing hourly bundles into a daily bundle. If there are too many daily
  bundles, then collapse the appropriate number of oldest daily bundles into the
  base bundle.

* `git-bundle-server update-all [<options>]`: For every configured route, run
  `git-bundle-server update <options> <route>`. This is called by the scheduler.

* `git-bundle-server stop <route>`: Stop computing bundles or serving content
  for the repository at the specified `<route>`. The route remains configured in
  case it is reenabled in the future.

* `git-bundle-server start <route>`: Start computing bundles and serving content
  for the repository at the specified `<route>`. This does not update the
  content immediately, but adds it back to the scheduler.

* `git-bundle-server delete <route>`: Remove the configuration for the given
  `<route>` and delete its repository data.

### Web server management

Independent of the management of the individual repositories hosted by the
server, you can manage the web server process itself using these commands:

* `git-bundle-server web-server start`: Start the web server process.

* `git-bundle-server web-server stop`: Stop the web server process.

Finally, if you want to run the web server process directly in your terminal,
for debugging purposes, then you can run `git-bundle-web-server`.

## Local development

### Building

> To avoid environment issues building and executing Go code, we recommend that
> you clone inside the `src` directory of your `GOROOT`.

In the root of your cloned repository, you can build the `git-bundle-server` and
`git-bundle-web-server` executables a few ways.

The first is to use GNU Make; from the root of the repository, simply run:

```ShellSession
$ make
```

If you do not have `make` installed on your system, you may instead run (again
from the repository root):

```ShellSession
$ go build -o bin/ ./...
```

### Testing and Linting

Unless otherwise specified, run commands from the repository root.

#### Unit tests

```
go test -v ./...
```

#### Linter

```
go vet ./...
```

#### End-to-end tests

In order to run these tests, you need to have a recent version of
[Node.js](https://nodejs.org) (current LTS version is a pretty safe bet) and NPM
installed.

For the standard set of tests (i.e., excluding exceptionally slow tests), run:

```
make e2e-test
```

To configure the test execution and filtering, set the `E2E_FLAGS` build
variable. The available options are:

* `--offline`: run all tests except those that require internet access.
* `--all`: run all tests, including slow performance tests.

The above modes are mutually exclusive; if multiple are specified, only the last
will be used. For example, `E2E_FLAGS="--offline --all"` is equivalent to
`E2E_FLAGS="--all"`.

:warning: The performance tests that are excluded by default clone very large
repos from the internet and can take anywhere from ~30 minutes to multiple hours
to run, depending on internet connectivity and other system resources.

## License

This project is licensed under the terms of the MIT open source license. Please
refer to [LICENSE][license] for the full terms.

## Maintainers

See [CODEOWNERS][codeowners] for a list of current project maintainers.

## Support

See [SUPPORT][support] for instructions on how to file bugs, feature requests,
and general questions/requests for help.
