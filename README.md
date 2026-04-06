# secrets

A self-hostable secret manager & browser extension.

## Features

- Store **logins**, **secure notes**, and **payment cards**
- Per-user encrypted databases - each account's data is independently encrypted at rest
- **Web UI** for browser-based management
- **CLI** for terminal and scripting workflows
- **Browser extension** for autofill and quick access
- **Go client library** for programmatic access

## Getting Started

### Docker

The quickest way to get up and running is with Docker. You'll need a config file first - create `config.toml`:

```toml
[http]
bind = "0.0.0.0:8080"

[database]
path = "/data"
ttl = "1h"
master_key = "<base64-encoded 32-byte key>"

[jwt]
issuer = "secrets"
audience = "secrets"
ttl = "30m"
signing_key = "<base64-encoded key>"
```

> **Warning:** Do not use all-zero keys in production. Generate secure random keys before deploying.

Then run the server:

```sh
docker run -d \
  -p 8080:8080 \
  -v $(pwd)/config.toml:/etc/secrets/config.toml \
  -v $(pwd)/data:/data \
  ghcr.io/dsb-labs/secrets serve /etc/secrets/config.toml
```

The web UI will be available at `http://localhost:8080`.

### Binary

Download the latest release for your platform from the [releases page](https://github.com/dsb-labs/secrets/releases), then run:

```sh
secrets serve config.toml
```

## Configuration

The server is configured via a TOML file passed as an argument to `secrets serve`.

| Section    | Key          | Description                                      | Default                        |
|------------|--------------|--------------------------------------------------|--------------------------------|
| `http`     | `bind`       | Address and port to listen on                    | `0.0.0.0:8080`                 |
| `database` | `path`       | Directory to store encrypted databases           | Platform config dir            |
| `database` | `ttl`        | How long before an idle account database closes  | `1h`                           |
| `database` | `master_key` | Base64-encoded 32-byte AES encryption key        | -                              |
| `jwt`      | `issuer`     | JWT issuer claim                                 | -                              |
| `jwt`      | `audience`   | JWT audience claim                               | -                              |
| `jwt`      | `ttl`        | How long issued tokens remain valid              | `1h`                           |
| `jwt`      | `signing_key`| Base64-encoded key used to sign JWTs             | -                              |

## Security Architecture

Each user's data is stored in its own independently-encrypted database. No two users share storage, and the server cannot access a user's data without their password.

### Encryption

When an account is created, an encryption key is derived from the user's password and their unique account ID using [Argon2id](https://en.wikipedia.org/wiki/Argon2) (3 iterations, 64 MB memory, 4 threads, 32-byte output). This derived key is used to encrypt the user's personal database with AES-256. The plaintext password is never stored.

The master key in the server config encrypts a separate top-level database that holds account records (email addresses, bcrypt password hashes, display names). User data is never stored there.

### Authentication

Passwords are hashed with bcrypt before being stored. On login, the password is verified against the stored hash, then the Argon2id key is re-derived to unlock the user's database. A JWT is issued for subsequent requests, signed with the configured signing key and valid for the configured TTL.

Idle databases are automatically locked after the configured `database.ttl` - the user must reauthenticate to unlock them again.

### Account Recovery

At account creation (and on password change), a **restore key** is returned. This is the raw Argon2id-derived encryption key - it can be used to decrypt the database directly if the account password is lost. It should be stored securely offline. The server does not store it.

### Summary

| Concern               | Mechanism                                      |
|-----------------------|------------------------------------------------|
| Data at rest          | AES-256, per-user encrypted Badger database    |
| Key derivation        | Argon2id (password + account ID as salt)       |
| Password storage      | bcrypt                                         |
| Session tokens        | JWT (configurable TTL and signing key)         |
| Idle session lockout  | Automatic database lock after configurable TTL |
| Account recovery      | Restore key (Argon2id-derived, user-held)      |

## CLI

The `secrets` binary provides a full CLI for interacting with a running server. All commands accept `--api-url` (default: `http://localhost:8080`) and `--config` flags.

| Command                   | Description                                         |
|---------------------------|-----------------------------------------------------|
| `serve`                   | Start the server                                    |
| `auth login`              | Authenticate and store a session token              |
| `auth logout`             | Clear the stored session token                      |
| `account create`          | Create a new account                                |
| `account info`            | Display current account details                     |
| `account change-password` | Change account password                             |
| `account delete`          | Delete the current account                          |
| `account restore`         | Restore an account using a restore key              |
| `login create`            | Store a new login credential                        |
| `login list`              | List stored logins (filter by `--domain`, `--name`) |
| `login get`               | Retrieve a login by ID                              |
| `login delete`            | Delete a login by ID                                |
| `note create`             | Store a new note                                    |
| `note list`               | List notes (filter by `--query`)                    |
| `note get`                | Retrieve a note by ID                               |
| `note delete`             | Delete a note by ID                                 |
| `tool export`             | Export the full database as JSON                    |

## Browser Extension

The browser extension lets you view and autofill your stored logins directly from your browser.

It is not currently published to any browser store. To install it:

1. Download `extension_<version>.zip` from the [releases page](https://github.com/dsb-labs/secrets/releases) and unzip it.
2. Open your browser's extension management page and enable **Developer mode**.
3. Click **Load unpacked** and select the unzipped directory.
4. Open the extension popup and enter your server URL to get started.

## Go Client Library

A Go client library is available at `github.com/dsb-labs/secrets/pkg/secrets` for integrating with the server programmatically. See the [package documentation](https://pkg.go.dev/github.com/dsb-labs/secrets/pkg/secrets) for the full API reference.

## Building from Source

**Requirements:** Go 1.26+, Node.js, pnpm

```sh
git clone https://github.com/dsb-labs/secrets
cd secrets

# Install Node dependencies
pnpm install --frozen-lockfile

# Build the web UI assets (must be done before the binary, as they are embedded in it)
pnpm build:ui

# Build the server binary
go build -o secrets .

# Build the browser extension
pnpm build:extension
```
