# `git-bundle-server`: Manage a self-hosted bundle server

By running this software, you can self-host a bundle server to work with Git's
[bundle URI feature][bundle-uris].

[bundle-uris]: https://github.com/git/git/blob/next/Documentation/technical/bundle-uri.txt

## Cloning and Building

Be sure to clone inside the `src` directory of your `GOROOT`.

Once there, you can build the `git-bundle-server` and `git-bundle-web-server`
executables with

```ShellSession
$ go build -o . ./...
```

## Bundle Management through CLI

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

## Web Server Management

Independent of the management of the individual repositories hosted by the
server, you can manage the web server process itself using these commands:

* `git-bundle-server web-server start`: Start the web server process.

* `git-bundle-server web-server stop`: Stop the web server process.

Finally, if you want to run the web server process directly in your terminal,
for debugging purposes, then you can run `git-bundle-web-server`.
