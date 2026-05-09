# Skill: seal-cli

# seal-cli Usage Guide

Help users interact with the Seal Subscriptions Merchant API from the command line using `seal-cli`.

## Agent Guidance

Best practices and operational guidance for AI coding agents using seal-cli.

### Key Principles

- **Auth comes first** — every command needs a token. `SEAL_TOKEN` env var beats any profile. Run `seal-cli profile set <name> --token <token>` once; then all commands work.
- **IDs are positional** — most commands accept the ID as a positional arg OR via `--id`. Prefer positional: `seal-cli sub get 12345` is faster to type than `seal-cli sub get --id 12345`.
- **Use `--json` for machine-readable output** — pipe through `jq` for filtering. Human-readable table output is not parseable.
- **Edit sends only changed flags** — `seal-cli subscription edit` only sends fields you explicitly pass. It will not overwrite unspecified fields.
- **Billing attempts need both IDs** — `--id` (the billing attempt ID) and `--subscription-id` are both required for all billing-attempt commands.

### Design Conventions

- **`<noun> <verb>` pattern** — `seal-cli subscription list`, `seal-cli billing-attempt skip`
- **Short aliases** — `subscription` → `sub`/`s`; `billing-attempt` → `ba`; `profile` → `prof`/`p`; subcommands: `list` → `ls`/`l`, `get` → `show`/`g`, `cancel` → `cx`, `pause` → `p`, `resume` → `r`, `reactivate` → `ra`, `edit` → `e`/`update`, `add-item` → `ai`, `remove-item` → `ri`/`rm`, `reschedule` → `rs`, `delete` → `rm`/`d`, `skip` → `sk`, `unskip` → `us`
- **Global flags** — `--profile <name>` selects a named profile; `--json` switches all output to raw JSON; `--agent` switches to AI-optimized compact output
- **503 = rate limit** — the client surfaces this as a readable error; wait and retry

### Agent Output Mode (`--agent`)

Use `--agent` (or set `SEAL_AGENT_MODE=1`) for all tool-use calls. This mode:
- **`sub get --agent`** → single compact JSON line, whitelisted fields only (~90% fewer tokens than `--json`)
- **`sub ls --agent`** → NDJSON (one JSON object per line), non-active statuses first, then `{"_meta":{"page":N,"count":N,"has_more":bool}}`
- **Mutations (`cancel`, `pause`, etc.) `--agent`** → `{"ok":true,"id":N,"action":"verb"}`
- **Errors `--agent`** → `{"ok":false,"status":404,"error":"...","detail":"/tmp/seal-cli-err-XXXX.json"}` — full API response in the temp file

**Agent field contract for `sub get --agent`:**
```
id, status, customer_email, customer_name, currency, total, interval,
address (optional), card (optional), items[], next_attempts[]
```

**Agent field contract for `sub ls --agent` rows:**
```
id, status, customer_email, currency, total, interval
```

**Prefer `--agent` over `--json`** in all agent/automated contexts. Use `--json` only when you need the full raw API payload for debugging.

```bash
# Set once for the whole session
export SEAL_AGENT_MODE=1

# Per-call
seal-cli sub get 12345 --agent
seal-cli sub ls --agent | head -10
seal-cli sub cancel 12345 --agent
```

### Context Window Tips

- **Use `--agent` mode** — purpose-built for AI agents, ~90% fewer tokens than `--json`
- Set `SEAL_AGENT_MODE=1` once per session instead of adding `--agent` to every call
- Use `--json | jq` only when you need full raw API payload for debugging
- Use `--with-items` and `--with-billing` on list/get to fetch related data in one call
- Use `-q` to filter `subscription list` by email, first name, or last name before fetching details
- On errors with `--agent`, the full API response is in the temp file at the path in `detail`

### Safety Rules

- Always confirm with the user before running destructive commands: `cancel`, `delete`
- `cancel` is irreversible via this CLI (use Shopify admin or Seal dashboard to undo)
- Never log or display the token value — use `seal-cli profile list` to inspect profile names only

---

## Prerequisites

### Installation

```bash
# Homebrew (recommended)
brew install hieutapt/tap/seal-cli

# One-line installer
curl -fsSL https://hieutapt.github.io/seals-subscription-cli/install | bash

# Specific version
curl -fsSL https://hieutapt.github.io/seals-subscription-cli/install | bash -s -- --version 0.2.0
```

### Authentication

```bash
# Save a token to a named profile (stored in ~/.seal-cli.yaml)
seal-cli profile set default --token <your-api-token>

# Use a one-off token without saving it
SEAL_TOKEN=<your-api-token> seal-cli subscription list

# Switch the active profile
seal-cli profile use production

# List all saved profiles
seal-cli profile list
```

`SEAL_TOKEN` environment variable always wins over any profile.

---

## Command Reference

### subscription (aliases: sub, s)

Manage Seal subscriptions.

#### list (aliases: ls, l)

```bash
seal-cli subscription list [flags]

Flags:
  -q, --query string    Filter by email, first name, or last name
  -p, --page int        Page number (default 1)
      --active          Show active subscriptions only
      --paused          Show paused subscriptions only
      --cancelled       Show cancelled subscriptions only
      --with-items      Include subscription items in response
      --with-billing    Include billing attempts in response
```

Examples:
```bash
seal-cli sub ls
seal-cli sub ls -q "jane@example.com" --json
seal-cli sub ls --active --with-items
seal-cli sub ls --page 2
```

#### get (aliases: show, g)

```bash
seal-cli subscription get [id] [flags]

Flags:
  --id int    Subscription ID
```

Examples:
```bash
seal-cli sub get 12345
seal-cli sub get --id 12345 --json | jq '.status'
```

#### cancel (alias: cx)

```bash
seal-cli subscription cancel [id] [flags]
```

⚠️ Destructive — confirm before running.

```bash
seal-cli sub cancel 12345
seal-cli sub cx 12345 --json
```

#### pause (alias: p)

```bash
seal-cli subscription pause [id]
seal-cli sub p 12345
```

#### resume (alias: r)

```bash
seal-cli subscription resume [id]
seal-cli sub r 12345
```

#### reactivate (alias: ra)

Reactivate a previously cancelled subscription.

```bash
seal-cli subscription reactivate [id]
seal-cli sub ra 12345
```

#### edit (aliases: e, update)

Only flags you pass are sent — unspecified fields are untouched.

```bash
seal-cli subscription edit [id] [flags]

Flags:
  --delivery-interval string   e.g. "2 week"
  --billing-interval string    e.g. "4 week"
  --min-cycles int
  --max-cycles int
  --delivery-price float
  --delivery-title string
  --first-name string
  --last-name string
  --address1 string
  --address2 string
  --city string
  --zip string
  --country string
  --country-code string        e.g. US
  --province string
  --province-code string       e.g. NY
  --phone string
  --company string
```

Examples:
```bash
seal-cli sub edit 12345 --delivery-interval "2 week"
seal-cli sub edit 12345 --min-cycles 3 --max-cycles 12
seal-cli sub edit 12345 --first-name "Jane" --city "Austin" --zip "78701"
```

#### add-item (alias: ai)

```bash
seal-cli subscription add-item [subscription-id] [flags]

Required flags:
  --product-id string    Shopify product ID
  --variant-id string    Shopify variant ID
  --title string         Product title
  --price float          Item price

Optional flags:
  --sku string
  --quantity int         (default 1)
  --taxable
  --requires-shipping    (default true)
  --one-time             Remove after next renewal (one-time add-on)
```

Examples:
```bash
seal-cli sub ai 12345 \
  --product-id 4648340258949 --variant-id 32694645424261 \
  --title "Bag of coffee 1kg" --price 24.00 --quantity 1

# One-time add-on
seal-cli sub ai 12345 \
  --product-id 4648340258949 --variant-id 32694645424261 \
  --title "Gift wrap" --price 5.00 --one-time
```

#### remove-item (aliases: ri, rm)

Use `seal-cli sub get <id>` first to find item IDs.

```bash
seal-cli subscription remove-item [subscription-id] --item-id <id> [--item-id <id> ...]
```

Examples:
```bash
seal-cli sub ri 12345 --item-id 1014923
seal-cli sub rm 12345 --item-id 1014923 --item-id 1014921
```

---

### billing-attempt (alias: ba)

Manage billing attempts. All commands require both `--id` (billing attempt) and `--subscription-id`.

#### reschedule (alias: rs)

```bash
seal-cli billing-attempt reschedule \
  --id <attempt-id> \
  --subscription-id <sub-id> \
  --date YYYY-MM-DD \
  --time HH:MM \
  --timezone "+00:00" \
  [--reset-schedule]
```

Example:
```bash
seal-cli ba rs \
  --id 123 --subscription-id 456 \
  --date 2025-12-01 --time 14:30 --timezone "-05:00"
```

#### delete (aliases: rm, d)

⚠️ Destructive.

```bash
seal-cli ba delete --id 123 --subscription-id 456
seal-cli ba rm --id 123 --subscription-id 456
```

#### skip (alias: sk)

```bash
seal-cli billing-attempt skip --id 123 --subscription-id 456
seal-cli ba sk --id 123 --subscription-id 456
```

#### unskip (alias: us)

```bash
seal-cli billing-attempt unskip --id 123 --subscription-id 456
seal-cli ba us --id 123 --subscription-id 456
```

---

### profile (aliases: prof, p)

Manage named API token profiles stored in `~/.seal-cli.yaml`.

```bash
# Save a token under a profile name
seal-cli profile set <name> --token <token>

# Switch the active default profile
seal-cli profile use <name>

# List all profiles (names only, tokens redacted)
seal-cli profile list
```

---

## Global Options

All commands support:

- `--profile <name>` — use a specific named profile (overrides active profile)
- `--json` — output raw JSON instead of a formatted table (full API payload, for jq/debugging)
- `--agent` — AI-agent-optimized output: compact JSON / NDJSON, whitelisted fields only
- `--help` — show help
- `--version` — show version, commit, and build date

`SEAL_AGENT_MODE=1` env var activates `--agent` mode for all commands in the session. `--agent` wins over `--json` if both are set.

---

## Workflow Patterns

### Investigate a customer's subscription

```bash
# Find by email
seal-cli sub ls -q "customer@example.com" --with-items --json | jq '.[0]'

# Get full details
seal-cli sub get 12345 --json | jq '{status, next_billing_date, items}'
```

### Pause all active subscriptions for a customer

```bash
# Find their sub IDs first
seal-cli sub ls -q "customer@example.com" --json | jq '.[].id'

# Then pause each
seal-cli sub p 12345
seal-cli sub p 67890
```

### Reschedule next billing to a specific date

```bash
# 1. Get the subscription to find the billing attempt ID
seal-cli sub get 12345 --with-billing --json | jq '.billing_attempts[0]'

# 2. Reschedule
seal-cli ba rs \
  --id <attempt-id> --subscription-id 12345 \
  --date 2025-12-15 --time 09:00 --timezone "+00:00"
```

### Add a one-time item to a subscription

```bash
seal-cli sub ai 12345 \
  --product-id 4648340258949 \
  --variant-id 32694645424261 \
  --title "Holiday gift wrap" \
  --price 4.99 \
  --one-time
```

### Update shipping address

```bash
seal-cli sub edit 12345 \
  --address1 "123 New St" \
  --city "Austin" \
  --zip "78701" \
  --country "United States" \
  --country-code "US" \
  --province "Texas" \
  --province-code "TX"
```

---

## Common Mistakes

- **Missing both IDs for billing-attempt commands** — `--id` is the billing attempt ID, `--subscription-id` is the subscription. Both are required, always.
- **Passing no flags to `edit`** — the command rejects requests with no changed fields; at least one flag is required.
- **Using `--json` without `jq`** — output is a raw JSON blob. Pipe through `jq .` for readability.
- **Forgetting `--item-id` on remove-item** — item IDs are not the same as subscription IDs. Use `seal-cli sub get <id>` to look them up first.
- **503 rate limit** — means the Seal Subscriptions API is throttling you. Wait a moment and retry.
- **Token precedence** — `SEAL_TOKEN` env var always wins over `--profile`. If you set both, the env var is used.
