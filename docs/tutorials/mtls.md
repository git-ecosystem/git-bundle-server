# Configuring mTLS authentication for the web server

[Mutual TLS (mTLS)][mtls] is a mechanism for mutual authentication between a
server and client. Configuring mTLS for a bundle server allows a server
maintainer to limit bundle server access to only the users that have been
provided with a valid certificate and establishes confidence that users are
interacting with a valid bundle server.

[mtls]: https://www.cloudflare.com/learning/access-management/what-is-mutual-tls/

## mTLS limitations

mTLS in the bundle server is configured **server-wide**, so it only provides
only a limited layer of protection against unauthorized access. Importantly,
**any** user with a valid client cert/private key pair will be able to access
**any** content on the bundle server. The implications of this include:

- If the bundle server manages repositories with separately controlled access,
  providing a user with a valid client cert/key for the bundle server may
  accidentally grant read access to Git data the user is not authorized to
  access on the remote host.
- If the remote host has branch-level security then the bundles may contain Git
  objects reachable from restricted branches.

## Creating certificates

mTLS connections involve the verification of two X.509 certificates: one from
the server, the other from the client. These certificates may be "self-signed",
or issued by a separate certificate authority; either will work with the bundle
server.

For both the server and client(s), both a public certificate (`.pem` file) and a
private key must be generated. Self-signed pairs can easily be generated with
OpenSSL; for example:

```bash
openssl req -x509 -newkey rsa:4096 -days 365 -keyout cert.key -out cert.pem
```

The above command will prompt the user to create a password for the 4096-bit RSA
private key stored in `cert.key` (for no password, use the `-nodes` option),
then fill in certificate metadata including locality information, company, etc.
The resulting certificate - stored in `cert.pem` - will be valid for 365 days.

> :rotating_light: If the "Common Name" of the server certificate does not match
> the bundle web server hostname (e.g. `localhost`), connections to the web
> server may fail.

If instead generating a certificate signed by a certificate authority (which can
itself be a self-signed certificate), a private key and _certificate signing
request_ must first be generated:

```bash
openssl req -new -newkey rsa:4096 -days 365 -keyout cert.key -out cert.csr
```

The user will be prompted to fill in the same certificate metadata (`-nodes` can
again be used to skip the password requirement on the private key). Once the
request is generated (in `cert.csr`), the request can be signed with the CA (in
the following example, with public certificate `ca.pem` and private key
`ca.key`):

```bash
openssl x509 -req -in cert.csr -CA ca.pem -CAkey ca.key -out cert.pem
```

This generates the CA-signed certificate `cert.pem`.

### :rotating_light: IMPORTANT: PROTECTING YOUR CREDENTIALS :rotating_light:

It is _extremely_ important that the private keys associated with the generated
certificates are safeguarded against unauthorized access (e.g. in a password
manager with minimal access).

If the server-side key is exposed, a malicious site could pose as a valid bundle
server and mislead users into providing it credentials or other private
information. Potentially-exposed server credentials should be replaced as soon
as possible, with the appropriate certificate authority/self-signed cert (_not_
the private key) distributed to users that use the server.

If a client-side key is exposed, an unauthorized user or malicious actor will
gain access to the bundle server and all content contained within it. **The
bundle server does not provide a mechanism for revoking certificates**, so
credentials will need to be rolled depending on how client certificates were
generated:

- If the `--client-ca` used by the bundle web server is a self-signed
  certificate corresponding to a single client, a new certificate/key pair will
  need to be generated and the bundle web server restarted[^1] to use the new
  `--client-ca` file.
- If the `--client-ca` is a concatenation of self-signed client certificates,
  the compromised certificate will need to be removed from the file and the
  bundle web server restarted.
- If the `--client-ca` is a certificate authority (a single certificate used to
  sign other certificates), the certificate authority and _all_ client
  certificates will need to be replaced.

## Configuring the web server

To configure the web server, three files are needed:

- If using self-signed client certificate(s), the client certificate `.pem`
  (which may contain one or multiple client certificates concatenated together)
  _or_ the certificate authority `.pem` used to sign client certificate(s). In
  the example below, this is `ca.pem`.
- The server `.pem` certificate file. In the example below, this is
  `server.pem`.
- The server private key file. In the example below, this is `server.key`.

The bundle server can then be configured with the `web-server` command to run in
the background:

```bash
git-bundle-server web-server start --force --port 443 --cert server.pem --key server.key --client-ca ca.pem
```

Alternatively, the web server can be started directly:

```bash
git-bundle-web-server --port 443 --cert server.pem --key server.key --client-ca ca.pem
```

If the contents of any of the certificate or key files change, the web server
process must be restarted. To reload the background web server daemon, run
`git-bundle-server web-server stop` followed by `git-bundle-server web-server
start`.

## Configuring Git

If cloning or fetching from the bundle server via Git, the client needs to be
configured to both verify the server certificate and send the appropriate client
certificate information. This configuration can be applied using environment
variables or `.gitconfig` values. The required configuration is as follows:

| Config (Environment) | Value |
| --- | --- |
| [`http.sslVerify`][sslVerify] (`GIT_SSL_NO_VERIFY`) | `true` for config, `false` for environment var. |
| [`http.sslCert`][sslCert] (`GIT_SSL_CERT`) | Path to the `client.pem` public cert file. |
| [`http.sslKey`][sslKey] (`GIT_SSL_KEY`) | Path to the `client.key` private key file. |
| [`http.sslCertPasswordProtected`][sslKeyPassword] (`GIT_SSL_CERT_PASSWORD_PROTECTED`) | `true` |
| [`http.sslCAInfo`][sslCAInfo] (`GIT_SSL_CAINFO`) | Path to the certificate authority file, including the server self-signed cert _or_ CA.[^2] |
| [`http.sslCAPath`][sslCAPath] (`GIT_SSL_CAPATH`) | Path to the directory containing certificate authority files, including the server self-signed cert _or_ CA.[^2] |

Configuring the certificate authority information, in particular, can be tricky.
Git does not have separate `http` configurations for clones/fetches vs. bundle
URIs; both will use the same settings. As a result, if cloning via HTTP(S) with
a bundle URI, users will need to _add_ the custom bundle server CA to the system
store. The process for adding to the system certificate authorities are
platform-dependent; for example, Ubuntu uses the
[`update-ca-certificates`][update-ca-certificates] command.

To avoid needing to add the bundle server CA to the trusted CA store, users can
instead choose to clone via SSH. In that case, only the bundle URI will use the
`http` settings, so `http.sslCAInfo` can point directly to the standalone server
CA.

[sslVerify]: https://git-scm.com/docs/git-config#Documentation/git-config.txt-httpsslVerify
[sslCert]: https://git-scm.com/docs/git-config#Documentation/git-config.txt-httpsslCert
[sslKey]: https://git-scm.com/docs/git-config#Documentation/git-config.txt-httpsslKey
[sslKeyPassword]: https://git-scm.com/docs/git-config#Documentation/git-config.txt-httpsslCertPasswordProtected
[sslCAInfo]: https://git-scm.com/docs/git-config#Documentation/git-config.txt-httpsslCAInfo
[sslCAPath]: https://git-scm.com/docs/git-config#Documentation/git-config.txt-httpsslCAPath
[update-ca-certificates]: https://manpages.ubuntu.com/manpages/xenial/man8/update-ca-certificates.8.html

[^1]: If using the `git-bundle-server web-server` command _and_ using a
      different `--client-ca` path than the old certificate, the `--force` option
      must be used with `start` to refresh the daemon configuration.
[^2]: These settings are passed to cURL internally, setting `CURLOPT_CAINFO` and
      `CURLOPT_CAPATH` respectively.