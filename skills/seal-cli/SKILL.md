---
name: seal-cli
description: CLI for the Seal Subscriptions Merchant API. Use when asked to manage subscriptions, billing attempts, or customer profiles via seal-cli — including listing, getting, cancelling, pausing, resuming, editing subscriptions, rescheduling or skipping billing attempts, and managing API token profiles.
---

# seal-cli

CLI for the Seal Subscriptions Merchant API. Pattern: `seal-cli <noun> <verb> [id] [flags]`.

> **Self-discovery:** run `seal-cli --help` or `seal-cli <command> --help` to inspect flags and aliases without loading this skill. Use this file for contracts, rules, and workflow patterns not visible in `--help`.

## Auth

`SEAL_TOKEN` env var > `--profile <name>` > active profile in `~/.seal-cli.yaml`.

```bash
seal-cli profile set <name> --token <token>   # save
seal-cli profile use <name>                   # switch active
seal-cli profile list                         # names only, tokens redacted
SEAL_TOKEN=<tok> seal-cli sub ls              # one-off
```

## Install

```bash
brew install hieutapt/tap/seal-cli
# or
curl -fsSL https://hieutapt.github.io/seals-subscription-cli/install | bash [-s -- --version 0.2.0]
```

## Global Flags

- `--profile <name>` — override active profile
- `--json` — raw API JSON (full payload, for jq/debugging)
- `--agent` — compact whitelisted JSON/NDJSON for agents (~90% fewer tokens than `--json`); wins over `--json`
- `SEAL_AGENT_MODE=1` — session-wide equivalent of `--agent`
- `--help`, `--version`

## Agent Output Mode (`--agent`)

**Use in all agent/automation contexts.** Use `--json` only for debugging raw payloads.

- `sub get --agent` → single compact JSON, fields:
  `id, status, customer_email, customer_name, currency, total, interval, address?, card?, items[]`
  + `next_attempts[]` (auto_charge) **or** `invoices[]` (recurring_invoice) — see [Payment Types](#payment-types)
- `sub ls --agent` → NDJSON, non-active statuses first, trailing `{"_meta":{"page":N,"count":N,"has_more":bool}}`. Row fields: `id, status, customer_email, currency, total, interval`
- Mutations → `{"ok":true,"id":N,"action":"verb"}`
- Errors → `{"ok":false,"status":404,"error":"...","detail":"/tmp/seal-cli-err-XXXX.json"}` (full response in temp file)

## Key Rules

- IDs are positional or `--id`; positional preferred: `seal-cli sub get 12345`
- `sub edit` sends only flags you pass; unspecified fields untouched; **rejects calls with zero flags**
- All `billing-attempt` commands require **both** `--id` (attempt) and `--subscription-id`
- `503` = rate limit; wait and retry
- ⚠️ **Confirm with user before**: `sub cancel`, `ba delete`. `cancel` is irreversible via this CLI (undo via Shopify admin / Seal dashboard)
- Never log/display token values

## Commands

> For full flags and aliases run `seal-cli <command> --help`.

| Group | Aliases | Subcommands |
|---|---|---|
| `subscription` | `sub`, `s` | `list`(ls), `get`(show/g), `cancel`(cx), `pause`(p), `resume`(r), `reactivate`(ra), `edit`(e), `add-item`(ai), `remove-item`(ri/rm) |
| `billing-attempt` | `ba` | `reschedule`(rs), `delete`(rm/d), `skip`(sk), `unskip`(us) |
| `profile` | `prof`, `p` | `set`, `use`, `list` |

Key non-obvious behaviours:
- `sub edit` — only flags you pass are sent; **at least one required**; `reactivate` is for cancelled subs
- `sub add-item` — `--one-time` removes item after next renewal
- `sub remove-item` — item IDs ≠ subscription IDs; get them via `sub get <id>` first
- `ba *` — requires **both** `--id` (attempt) and `--subscription-id`

## Workflow Patterns

### Investigate a customer

```bash
seal-cli sub ls -q "customer@example.com" --with-items --agent
seal-cli sub get 12345 --agent
```

### Pause all active subs for a customer

```bash
seal-cli sub ls -q "customer@example.com" --agent | jq -r .id | xargs -n1 seal-cli sub p
```

### Reschedule next billing

```bash
# 1. Get billing attempt ID
seal-cli sub get 12345 --with-billing --json | jq '.billing_attempts[0]'
# 2. Reschedule
seal-cli ba rs --id <attempt-id> --subscription-id 12345 \
  --date 2025-12-15 --time 09:00 --timezone "+00:00"
```

### One-time add-on

```bash
seal-cli sub ai 12345 --product-id 4648340258949 --variant-id 32694645424261 \
  --title "Holiday gift wrap" --price 4.99 --one-time
```

### Update shipping address

```bash
seal-cli sub edit 12345 --address1 "123 New St" --city "Austin" --zip "78701" \
  --country "United States" --country-code US --province "Texas" --province-code TX
```

## Common Mistakes

- **`ba *` missing `--subscription-id`** — both `--id` (attempt) and `--subscription-id` always required.
- **`sub edit` with zero flags** — rejected; pass at least one.
- **`sub remove-item` wrong ID** — `--item-id` is the item, not the sub; look up via `sub get <id>`.
- **`--json` without `jq`** — raw blob; pipe through `jq .`.
- **503** — rate limit; wait and retry.
- **Token precedence** — `SEAL_TOKEN` always wins over `--profile`.

## Payment Types

| Type | Has | Empty |
|---|---|---|
| `auto_charge` | `billing_attempts[]` → agent: `next_attempts[]` | `invoices[]` |
| `recurring_invoice` | `invoices[]` (id, date, payment_status) | `billing_attempts[]` |
