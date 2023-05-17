# Auth configuration examples

This directory contains examples of auth configurations that may be used as a
reference for setting up auth for a bundle server.

> **Warning**
>
> The examples contained within this directory should not be used directly in a
> production context due to publicly-visible (in this repo) credentials.

## Built-in modes

### Fixed credential/single-user auth

The file [`config/fixed.json`][fixed-config] configures [Basic
authentication][basic] with username "admin" and password "bundle_server".

[fixed-config]: ./config/fixed.json
[basic]: ../../docs/technical/auth-config.md#basic-auth-server-wide

## Plugin mode

The example plugin implemented in [`_plugins/simple-plugin.go`][simple-plugin]
can be built (from this directory) with:

```bash
go build -buildmode=plugin -o ./plugins/ ./_plugins/simple-plugin.go
```

which will create `simple-plugin.so` - this is your plugin file.

To use this plugin with `git-bundle-web-server`, the config in
[`config/plugin.json`][plugin-config] needs to be updated with the SHA256
checksum of the plugin. This value can be determined by running (from this
directory):

```bash
shasum -a 256 ./_plugins/simple-plugin.so
```

The configured `simple-plugin.so` auth middleware implements Basic
authentication with a hardcoded username "admin" and a password that is based on
the requested route (if the requested route is `test/repo` or
`test/repo/bundle-123456.bundle`, the password is "test_repo").

> **Note**
>
> The example `plugin.json` contains a relative, rather than absolute, path to
> the plugin file, relative to the root of this repository. This is meant to
> facilitate more portable testing and is  _not_ recommended for typical use;
> please use an absolute path to identify your plugin file.

[simple-plugin]: ./_plugins/simple-plugin.go
[plugin-config]: ./config/plugin.json
