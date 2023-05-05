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
                The auth mode to use. Not case-sensitive.
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
