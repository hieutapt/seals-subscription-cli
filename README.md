# seal-cli

A CLI for the [Seal Subscriptions](https://www.sealsubscriptions.com) Merchant API.

## Install via Homebrew

```bash
brew tap hieutapt/tap
brew install seal-cli
```



## Quick Start

```bash
# Authenticate with a named profile
seal-cli profile set --name production --token YOUR_SEAL_TOKEN

# Or use the SEAL_TOKEN environment variable
export SEAL_TOKEN=YOUR_SEAL_TOKEN

# List subscriptions
seal-cli subscription list

# Get a specific subscription
seal-cli subscription get 12345

# Cancel / pause / resume
seal-cli subscription cancel 12345
seal-cli subscription pause  12345
seal-cli subscription resume 12345
```

## Commands

### `subscription`

| Command | Description |
|---|---|
| `list` | List subscriptions (supports `--query`, `--active`, `--paused`, `--cancelled`, `--page`) |
| `get <id>` | Get full details of a subscription |
| `cancel <id>` | Cancel a subscription |
| `pause <id>` | Pause a subscription |
| `reactivate <id>` | Reactivate a cancelled subscription |
| `resume <id>` | Resume a paused subscription |
| `edit <id>` | Edit interval, address, cycles, delivery price, etc. |

### `billing-attempt`

| Command | Description |
|---|---|
| `reschedule` | Reschedule a billing attempt to a new date/time |
| `delete` | Delete an unprocessed billing attempt |
| `skip` | Skip a billing attempt |
| `unskip` | Unskip a previously skipped billing attempt |

### `profile`

| Command | Description |
|---|---|
| `set` | Create or update a named profile |
| `use <name>` | Switch the active profile |
| `list` | List all configured profiles |

## Global Flags

| Flag | Description |
|---|---|
| `--profile <name>` | Use a specific named profile |
| `--json` | Output raw JSON instead of pretty table |
| `--version` | Print version info |

## Authentication Priority

1. `SEAL_TOKEN` environment variable (highest priority)
2. Active profile in `~/.seal-cli.yaml`

## Release a New Version

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will build cross-platform binaries and push the Homebrew formula to your tap automatically.
