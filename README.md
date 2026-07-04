# `Audited`

[![main](https://github.com/getaudited/audited/actions/workflows/main.yml/badge.svg)](https://github.com/getaudited/audited/actions/workflows/main.yml)

Audit log management for cloud-native applications.

## Installation

### Docker Hub

The published image is `getauditeddev/audited`. Tags follow semver (`1.2.3`, `1.2`, `1`).

```bash
docker pull getauditeddev/audited:latest
```

Run the service (requires external Clickhouse):

```bash
docker run -p 8080:8080 \
  -e ADT_DATABASE_URL="clickhouse://default:password@localhost:9000/default?secure=false&skip_verify=false" \
  -e ADT_HTTP_PORT=8080 \
  -e ADT_JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----\n...\n-----END PUBLIC KEY-----" \
  -e ADT_JWT_PRIVATE_KEY="-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----" \
  -e ADT_ADMIN_EMAIL="admin@example.com" \
  -e ADT_ADMIN_PASSWORD="changeme" \
  getauditeddev/audited:latest
```

| Variable | Required | Description                                                          |
|---|---|----------------------------------------------------------------------|
| `ADT_DATABASE_URL` | Yes | Clickhouse DSN                                                       |
| `ADT_HTTP_PORT` | Yes | HTTP listen port                                                     |
| `ADT_JWT_PUBLIC_KEY` | Yes | ECDSA public key (PEM) for JWT verification                          |
| `ADT_JWT_PRIVATE_KEY` | Yes | ECDSA private key (PEM) for JWT signing                              |
| `ADT_ADMIN_EMAIL` | Yes | Email for the bootstrap admin user                                   |
| `ADT_ADMIN_PASSWORD` | Yes | Password for the bootstrap admin user                                |
| `ADT_ALLOWED_CORS_ORIGIN` | No | Comma-separated list of allowed CORS origins                         |
| `ADT_LOG_LEVEL` | No | Log verbosity (`DEBUG`, `INFO`, `WARN`, `ERROR`). Defaults to `INFO` |

The service auto-applies migrations on startup, so no separate migration step is needed.
