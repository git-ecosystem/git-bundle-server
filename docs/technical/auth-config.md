# Configuring access control

User access to web server endpoints on the bundle server is configured via the
`--auth-config` option to `git-bundle-web-server` and/or `git-bundle-server
web-server`. The auth config is a JSON file that identifies the type of access
control requested and information needed to configure it.

## Schema

The JSON file contains the following fields:

 <table>
    <thead>
        <tr>
            <th/>
            <th>Field</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <th rowspan="2">Common fields</th>
            <td><code>mode</code></td>
            <td>string</td>
            <td>
                <p>The auth mode to use. Not case-sensitive.</p>
                Available options:
                <ul>
                    <li><code>fixed</code></li>
                    <li><code>plugin</code></li>
                </ul>
            </td>
        </tr>
        <tr>
            <td><code>parameters</code> (optional; depends on mode)</td>
            <td>object</td>
            <td>
                A structure containing mode-specific key-value configuration
                fields, if applicable.
            </td>
        </tr>
        <tr>
            <th rowspan="3"><code>plugin</code>-only</th>
            <td><code>path</code></td>
            <td>string</td>
            <td>
                The absolute path to the auth plugin <code>.so</code> file.
            </td>
        </tr>
        <tr>
            <td><code>initializer</code></td>
            <td>string</td>
            <td>
                The name of the symbol within the plugin binary that can invoked
                to create the <code>AuthMiddleware</code>. The initializer:
                <ul>
                    <li>
                        Must have the signature
                        <code>func(json.RawMessage) (AuthMiddleware, error)</code>.
                    </li>
                    <li>
                        Must be exported in its package (i.e.,
                        <code>UpperCamelCase</code> name).
                    </li>
                </ul>
                See <a href="#plugin-mode">Plugin mode</a> for more details.
            </td>
        </tr>
        <tr>
            <td><code>sha256</code></td>
            <td>string</td>
            <td>
                The SHA256 checksum of the plugin <code>.so</code> file,
                rendered as a hex string. If the checksum does not match the
                calculated checksum of the plugin file, the web server will
                refuse to start.
            </td>
        </tr>
    </tbody>
</table>

## Built-in modes

### Fixed/single-user auth (server-wide)

**Mode: `fixed`**

This mode implements [Basic authentication][basic-rfc], authenticating each
request against a fixed username/password pair that is global to the web server.

[basic-rfc]: https://datatracker.ietf.org/doc/html/rfc7617

#### Parameters

The `parameters` object _must_ be specified for this mode, and both of the
fields below are required.

<table>
    <thead>
        <tr>
            <th>Field</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td><code>username</code></td>
            <td>string</td>
            <td>
                The username string for authentication. The username <i>must
                not</i> contain a colon (":").
            </td>
        </tr>
        <tr>
            <td><code>passwordHash</code></td>
            <td>string</td>
            <td>
                <p>
                    The SHA256 hash of the password string. There are no
                    restrictions on characters used for the password.
                </p>
                <p>
                    The hash of a string can be generated on the command line
                    with the command
                    <code>echo -n '&lt;your string&gt;' | shasum -a 256</code>.
                </p>
            </td>
        </tr>
    </tbody>
</table>

#### Examples

Valid (username `admin`, password `test`):

```json
{
    "mode": "fixed",
    "parameters": {
        "username": "admin",
        "passwordHash": "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
    }
}
```

Valid (empty username & password):

```json
{
    "mode": "fixed",
    "parameters": {
        "usernameHash": "",
        "passwordHash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
    }
}
```

Invalid:

```json
{
    "mode": "fixed",
    "parameters": {
        "username": "admin",
        "passwordHash": "test123"
    }
}
```

Invalid:

```json
{
    "mode": "fixed",
    "parameters": {
        "username": "admin:MY_PASSWORD",
    }
}
```

## Plugin mode

**Mode: `plugin`**

Plugin mode allows users to develop their custom auth middleware to serve a more
specific platform or need than the built-in modes (e.g., host-based federated
access). The bundle server makes use of Go's [`plugin`][plugin] package to load
the plugin and create an instance of the specified middleware.

### The plugin

The plugin is a `.so` shared library built using `go build`'s
`-buildmode=plugin` option. The custom auth middleware must implement the
`AuthMiddleware` interface defined in the exported `auth` package of this
repository. Additionally, the plugin must contain an initializer function that
creates and returns the custom `AuthMiddleware` interface. The function
signature of this initializer is:

```go
func (json.RawMessage) (AuthMiddleware, error)
```

- The `json.RawMessage` input is the raw bytes of the `parameters` object (empty
  if `parameters` is not in the auth config JSON).
- The `AuthMiddleware` is an instance of the plugin's custom `AuthMiddleware`
  implementation. If this is `nil` and `error` is not `nil`, the web server will
  fail to start.
- If the `AuthMiddleware` cannot be initialized, the `error` captures the
  context of the failure. If `error` is not `nil`, the web server will fail to
  start.

> **Note**
>
> While this project is in a pre-release/alpha state, the `AuthMiddleware`
> and initializer interfaces may change, breaking older plugins.

After the `AuthMiddleware` is loaded, its `Authorize()` function will be called
for each valid route request. The `AuthResult` returned must be created with one
of `Allow()` or `Deny()`; an accepted request will continue on to the logic for
serving bundle server content, a rejected one will return immediately with the
specified code and headers.

Note that these requests may be processed in parallel, therefore **it is up to
the developer of the plugin to ensure their middleware's `Authorize()` function
is thread-safe**! Failure to do so could create race conditions and lead to
unexpected behavior.

### The config

When using `plugin` mode in the auth config, there are a few additional fields
that must be specified that are not required for built-in modes: `path`,
`initializer`, `sha256`.

There are multiple ways to determine the SHA256 checksum of a file, but an
easy way to do so on the command line is:

```bash
shasum -a 256 path/to/your/plugin.so
```

> **Warning**
>
> In the current plugin-loading implementation, the SHA256 checksum of the
> specified plugin is calculated and compared before loading its symbols. This
> opens up the possibility of a [time-of-check/time-of-use][toctou] attack
> wherein a malicious actor replaces a valid plugin file with their own plugin
> _after_ the checksum verification of the "good" file but before the plugin is
> loaded into memory.
>
> To mitigate this risk, ensure 'write' permissions are disabled on your plugin
> file. And, as always, practice caution when running third party code that
> interacts with credentials and other sensitive information.

[plugin]: https://pkg.go.dev/plugin
[toctou]: https://en.wikipedia.org/wiki/Time-of-check_to_time-of-use

### Examples

An example plugin and corresponding config can be found in the
[`examples/auth`][examples-dir] directory of this repository.

[examples-dir]: ../../examples/auth
