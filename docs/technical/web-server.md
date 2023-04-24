# Web Server API Reference

This document contains an API specification for the web server created with
`git-bundle-web-server`. It is primarily meant to be used by Git via the
[bundle-uri feature][bundle-uris].

> **Warning**
>
> First and foremost, the goal of this API is compatibility with Git's bundle
> URI feature. We will attempt to keep it up-to-date with the latest version of
> Git but, due to both the newness of the feature and experimental state of the
> server, we cannot make guarantees of backward compatibility.

[bundle-uris]: https://git-scm.com/docs/bundle-uri

## Get a repository's bundle list

Get the list of bundles configured for a given bundle server route.

<table>
    <tbody>
        <tr>
            <th>Method</th>
            <td><code>GET</code></td>
        </tr>
        <tr>
            <th>Route</th>
            <td><code>/{route}</code></td>
        </tr>
        <tr>
            <th>Example Request</th>
            <td><code>curl http://localhost:8080/OWNER/REPO</code></td>
        </tr>
        <tr>
            <th>Example Response</th>
<td>

```
[bundle]
	version = 1
	mode = all
	heuristic = creationToken

[bundle "1678494078"]
	uri = REPO/base-1678494078.bundle
	creationToken = 1678494078

[bundle "1679527263"]
	uri = REPO/bundle-1679527263.bundle
	creationToken = 1679527263

[bundle "1680561322"]
	uri = REPO/bundle-1680561322.bundle
	creationToken = 1680561322
```

</td>
        </tr>
    </tbody>
</table>

### Path parameters

| Name    | Type   | Required  | Description |
| ------- | ------ | --------- | ----------- |
| `route` | string | Yes       | The route of a repository created with `git-bundle-server init` for which the list of active bundles is requested. Route should be in `OWNER/REPO` format. |

### HTTP response status codes

| Code  | Description |
| ----- | ----------- |
| `200` | OK          |
| `404` | Specified route does not exist or has no bundles configured |

## Download a bundle

Download an individual bundle.

<table>
    <tbody>
        <tr>
            <th>Method</th>
            <td><code>GET</code></td>
        </tr>
        <tr>
            <th>Route</th>
            <td><code>/{route}</code></td>
        </tr>
        <tr>
            <th>Example Request</th>
            <td><code>curl http://localhost:8080/OWNER/REPO/bundle-1679527263.bundle</code></td>
        </tr>
        <tr>
        <th>Example Response</th>
            <td><i>Binary </i><a href="https://git-scm.com/docs/git-bundle"><code>git bundle</code></a><i> bundle content.</i></td>
        </tr>
    </tbody>
</table>

### Path parameters

| Name     | Type   | Required  | Description |
| -------- | ------ | --------- | ----------- |
| `route`  | string | Yes       | The route of a repository containing the desired bundle. Route should be in `OWNER/REPO` format. |
| `bundle` | string | Yes       | The filename of the desired bundle as identified by the `route`'s bundle list. |

### HTTP response status codes

| Code  | Description |
| ----- | ----------- |
| `200` | OK          |
| `404` | The specified bundle does not exist |
