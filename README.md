<p align="center">
    <picture>
      <source media="(prefers-color-scheme: light)" srcset="/.github/assets/audited-light.png" height="128" width="auto">
      <source media="(prefers-color-scheme: dark)" srcset="/.github/assets/audited-dark.png" height="128" width="auto">
      <img alt="Fallback image description" src="/.github/assets/audited-dark.png" height="128" width="auto">
    </picture>
</p>

[![Go Coverage](https://github.com/getaudited/audited/wiki/coverage.svg)](https://raw.githack.com/wiki/getaudited/audited/coverage.html)
[![main](https://github.com/getaudited/audited/actions/workflows/main.yml/badge.svg)](https://github.com/getaudited/audited/actions/workflows/main.yml)
![GitHub tag](https://img.shields.io/github/tag/getaudited/audited.svg)
[![Docker Pulls](https://img.shields.io/docker/pulls/getauditeddev/audited?logo=docker)](https://hub.docker.com/r/getauditeddev/audited/)

Audit log management for cloud-native applications. Written in Go and Clickhouse.

## Documentation

Audited documentation can be found [here](https://docs.getaudited.dev).

## Installation

### Docker Hub

The published image is `getauditeddev/audited`. Tags follow semver (`1.2.3`, `1.2`, `1`).

```bash
docker pull getauditeddev/audited:latest
```

Run the service (requires external Clickhouse):

```bash
docker run -p 8080:8080 \
  -e ADT_CLICKHOUSE_HOSTS="clickhouse:9000" \
  -e ADT_CLICKHOUSE_DBNAME="default" \
  -e ADT_CLICKHOUSE_USERNAME="default" \
  -e ADT_CLICKHOUSE_PASSWORD="password" \
  -e ADT_HTTP_PORT=8080 \
  -e ADT_JWT_SECRET="a-long-random-secret" \
  -e ADT_ADMIN_EMAIL="admin@example.com" \
  -e ADT_ADMIN_PASSWORD="changeme" \
  getauditeddev/audited:latest
```

### Connecting to a managed Clickhouse

Managed Clickhouse services (Clickhouse Cloud and similar) require TLS on the native port. Set `ADT_CLICKHOUSE_TLS_ENABLED=true` and point `ADT_CLICKHOUSE_HOSTS` at the secure native endpoint (usually port `9440`):

```bash
docker run -p 8080:8080 \
  -e ADT_CLICKHOUSE_HOSTS="your-instance.clickhouse.cloud:9440" \
  -e ADT_CLICKHOUSE_TLS_ENABLED=true \
  ... \
  getauditeddev/audited:latest
```

TLS is off by default, which works for a local Clickhouse on port `9000`. If you terminate TLS with a self-signed certificate, `ADT_CLICKHOUSE_TLS_INSECURE_SKIP_VERIFY=true` disables certificate verification — it makes the connection vulnerable to man-in-the-middle attacks, so keep it out of production.

### JWT signing

You must configure JWT signing using **one** of the following options:

- **HMAC (simpler)** — set `ADT_JWT_SECRET` to a long random string. Tokens are signed and verified with this shared secret.
- **ECDSA key pair** — set both `ADT_JWT_PUBLIC_KEY` and `ADT_JWT_PRIVATE_KEY` (PEM). Tokens are signed with the private key and verified with the public key.

If `ADT_JWT_SECRET` is set it takes precedence; otherwise both key pair variables are required. Startup fails if neither option is fully configured.

| Variable | Required | Description                                                          |
|---|---|----------------------------------------------------------------------|
| `ADT_CLICKHOUSE_HOSTS` | Yes | Comma-separated list of Clickhouse hosts (e.g. `clickhouse:9000`)   |
| `ADT_CLICKHOUSE_DBNAME` | Yes | Clickhouse database name                                            |
| `ADT_CLICKHOUSE_USERNAME` | Yes | Clickhouse username                                                 |
| `ADT_CLICKHOUSE_PASSWORD` | Yes | Clickhouse password                                                 |
| `ADT_CLICKHOUSE_TLS_ENABLED` | No | Connect to Clickhouse over TLS (default `false`). Required by most managed/cloud Clickhouse providers |
| `ADT_CLICKHOUSE_TLS_INSECURE_SKIP_VERIFY` | No | Skip verification of the Clickhouse TLS certificate (default `false`). Only for self-signed certificates in development |
| `ADT_ADMIN_EMAIL` | Yes | Email for the bootstrap admin user                                   |
| `ADT_ADMIN_PASSWORD` | Yes | Password for the bootstrap admin user                                |
| `ADT_JWT_SECRET` | Conditional | HMAC secret for signing/verifying JWTs. Alternative to the key pair below; required unless `ADT_JWT_PUBLIC_KEY` and `ADT_JWT_PRIVATE_KEY` are set |
| `ADT_JWT_PUBLIC_KEY` | Conditional | ECDSA public key (PEM) for JWT verification. Required (with the private key) unless `ADT_JWT_SECRET` is set |
| `ADT_JWT_PRIVATE_KEY` | Conditional | ECDSA private key (PEM) for JWT signing. Required (with the public key) unless `ADT_JWT_SECRET` is set |
| `ADT_HTTP_PORT` | No | HTTP listen port (default `8080`)                                    |
| `ADT_ALLOWED_CORS_ORIGIN` | No | Comma-separated list of allowed CORS origins                         |
| `ADT_LOG_LEVEL` | No | Log verbosity (`DEBUG`, `INFO`, `WARN`, `ERROR`). Defaults to `INFO` |

The service auto-applies migrations on startup, so no separate migration step is needed.

## Client library overview

| Language    | Officially supported |
|-------------|----------------------|
| [Go](https://github.com/getaudited/audited-go)      | ✅                    |
| Rust        | ⚠️                   |
| Typescript  | ⚠️                   |
| Java        | ⚠️                   |
| PHP         | ⚠️                   |
| C# (dotnet) | ⚠️                   |
| Python      | ⚠️                   |
| Elixir      | ⚠️                   |

## Observability and Monitoring

_Needs work, stay tuned._

## Status

Audited is still in early-stage development.

## License

Audited is free and open source software, licensed under the [AGPL v3](./LICENSE) While often misunderstood, this license is very permissive and allows the following without any additional requirements from you or your organization:

* Internal use
* Private modifications for internal use without sharing any source code

You can freely use Audited without having to share any source code, including proprietary work product or any modifications to Audited you make.

AGPL was written specifically for organizations that offer Audited as a public service (e.g. database cloud providers) and require those organizations to share any modifications they make to Audited, including new features and bug fixes.

## Contributions

Please read our [Contributions Guidelines](./CONTRIBUTING.md).