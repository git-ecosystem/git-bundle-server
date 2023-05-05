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
            <th>Field</th>
            <th>Type</th>
            <th>Description</th>
        </tr>
    </thead>
    <tbody>
        <tr>
            <td><code>mode</code></td>
            <td>string</td>
            <td>
                <p>The auth mode to use. Not case-sensitive.</p>
                Available options:
                <ul>
                    <li><code>fixed</code></li>
                </ul>
            </td>
        </tr>
        <tr>
            <td><code>parameters</code></td>
            <td>object</td>
            <td>
                A structure containing mode-specific key-value configuration
                fields, if applicable.
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
