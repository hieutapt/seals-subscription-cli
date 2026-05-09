# AGENTS.md

## Build & run

```bash
go build ./...          # compile; output binary: ./seal-cli (gitignored)
go build -o seal-cli .  # explicit binary name
go run . --help         # run without building
```

No Makefile, no task runner.

```bash
go test ./...             # run all tests
go test ./internal/api/... -v  # run API tests with verbose output
```

## Local SDLC (git hooks)

A `pre-push` hook runs `go test ./...` before every push. Install it once after cloning:

```bash
cp scripts/pre-push .git/hooks/pre-push
chmod +x .git/hooks/pre-push
```

Skip in an emergency: `git push --no-verify`

## CI

`.github/workflows/release.yml` runs two jobs:

- **test** — triggers on every push to `main` and every pull request; runs `go build ./...` then `go test ./... -v -count=1`.
- **release** — triggers on `v*` tags only; gated behind `test` passing; runs GoReleaser.

## Module path

`github.com/hieutapt/seals-subscription-cli` — must match in every `*.go` import and `go.mod`. Do not use `github.com/hieu/`.

## Package layout

```
main.go                  # entry point; injects version/commit/date via ldflags
cmd/                     # Cobra commands (root, subscription, billing_attempt, profile)
internal/api/            # HTTP client for Seal Subscriptions REST API
internal/config/         # profile config (~/.seal-cli.yaml) + SEAL_TOKEN env var
internal/output/         # table + JSON rendering
```

- `cmd/root.go` wires global flags (`--profile`, `--json`) and calls `SetVersionInfo` from `main.go`.
- All commands resolve auth via `config.ActiveToken(profile)` — env var `SEAL_TOKEN` beats any profile.

## tablewriter API

The project uses `github.com/olekukonko/tablewriter v1.1.4`, which has a **new API**. Use `tablewriter.NewTable(w, opts...)` not `tablewriter.NewWriter`. Pass headers via `tablewriter.WithHeader([]string{...})`. The old `SetHeader`, `SetBorder`, `SetAlignment` methods do not exist on this version.

## Release process

Releases are fully automated via GoReleaser + GitHub Actions:

```bash
git tag v1.2.3
git push origin v1.2.3
```

- Workflow: `.github/workflows/release.yml` — triggers on `v*` tags only.
- GoReleaser config: `.goreleaser.yaml` (version 2).
- Builds: `CGO_ENABLED=0`, darwin/linux/windows × amd64/arm64.
- After release, GoReleaser auto-pushes `Formula/seal-cli.rb` to `github.com/hieutapt/homebrew-tap`.
- Required secret on this repo: `HOMEBREW_TAP_GITHUB_TOKEN` (token with write access to `homebrew-tap`).
- `brews:` key is used (not `homebrew_casks:`); GoReleaser warns it's deprecated but it still works.

## Version injection

`main.go` declares `var version, commit, date` and passes them to `cmd.SetVersionInfo`. GoReleaser injects via ldflags:
```
-X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}
```
Local builds show `dev / none / unknown`.

## API surface covered

Seal Subscriptions Merchant API (`https://app.sealsubscriptions.com/shopify/merchant/api`):
- Subscriptions: list, get, cancel, pause, reactivate, resume, edit
- Subscription items: add-item, remove-item
- Billing attempts: reschedule, delete, skip, unskip
- Profile management (local only, not an API endpoint)

Not implemented: create subscription, discount codes, webhooks, fulfillment orders, magic link, quick checkout URL.

## Install script

`scripts/install.sh` is the canonical curl-installable installer. It is also copied verbatim to `docs/install` for GitHub Pages serving.

**Keep them in sync** — after editing `scripts/install.sh`, run:
```bash
cp scripts/install.sh docs/install
```

The install URL (once GitHub Pages is enabled on the `docs/` folder of `main`):
```bash
curl -fsSL https://hieutapt.github.io/seals-subscription-cli/install | bash
```

## Agent skill

`skills/seal-cli/SKILL.md` is the agent skill file for seal-cli. It documents all commands, flags, aliases, and workflow patterns for AI agents.

To install the skill into an agent environment:
```bash
# Copy to the agent skills directory (adjust path to your setup)
cp -r skills/seal-cli ~/.claude/skills/seal-cli
```

The skill is self-contained — no external dependencies.
