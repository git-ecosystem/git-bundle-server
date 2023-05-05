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
