# repo-release-notifier

Small Go service that lets people subscribe (by email) to GitHub release
notifications for a given repo. Users confirm via an emailed link; a
background scanner polls GitHub and sends a mail whenever the latest tag
changes.

Monolith, single binary, PostgreSQL for storage, optional Redis for caching
GitHub responses.

## Running

Everything runs from `docker-compose up --build` — that brings up Postgres,
Redis and the app. Locally, `go run ./cmd/server` works once you have a `.env`
populated (see `.env.example`).

Useful commands while developing:

```
go test ./... -race
golangci-lint run ./...
```

`DATABASE_URL` (or the split `DB_*` vars) plus the SMTP credentials are the
only required config. `REDIS_URL`, `GITHUB_TOKEN` and `API_KEY` are optional
but recommended in anything resembling production.

## Implementation notes

A few things that aren't obvious from reading the code:

- **Soft delete + uniqueness.** `subscriptions` uses GORM soft delete so
  unsubscribing keeps the row. A plain `UNIQUE(email, repo)` would then
  block the user from ever re-subscribing, so migration adds a partial
  unique index scoped to `deleted_at IS NULL` instead.
- **Silent first scan.** When a subscription has no `last_seen_tag` yet,
  the first scan records the current tag without emailing. Otherwise every
  fresh subscriber would immediately receive a notification for whatever
  the latest tag already was.
- **Zero-width space in email URLs.** The confirmation email prints the
  URL twice — once as a button, once as plaintext for copy-paste. Mail
  clients aggressively auto-linkify bare URLs, which defeats the fallback.
  `breakAutoLink` inserts a U+200B between `https` and `://` so the text
  stays text; browsers strip the ZWSP on paste.
- **Redis cache.** `internal/github/cached_client.go` wraps the GitHub
  client with a 10-minute TTL on successful `ValidateRepo` and
  `GetLatestRelease` results (including 404). Rate-limit / network
  failures are passed through uncached so the next call retries. If
  `REDIS_URL` is empty or Redis is down at startup, the service logs and
  runs without the cache.
- **HTML subscribe page.** `GET /` serves a minimal form from
  `internal/templates/pages/subscribe.html`, embedded via `//go:embed`.
  All emitted HTML (emails and pages) lives under `internal/templates/`
  and goes through `RenderEmail` / `Page`.
- **One-click unsubscribe headers.** Outgoing mails include
  `List-Unsubscribe` / `List-Unsubscribe-Post` so Gmail, Apple Mail, etc.
  render a native unsubscribe button.

## Tradeoffs I skipped on purpose

**gRPC.** Listed as a bonus, skipped deliberately. GitHub's public API is
REST/GraphQL, there's no internal service fan-out, and the only callers
are a browser and email links — neither speaks gRPC natively. Adding it
would be pure surface area.

**Constant-time token comparison.** Tokens are 32 bytes from `crypto/rand`
(64 hex chars, 256 bits). A timing attack against an indexed `WHERE token
= ?` lookup on that entropy isn't reachable in practice, and forcing a
constant-time check would mean a full table scan per request. Not worth
the real cost for the theoretical attack.

**API key scope.** `POST /api/subscribe` and `GET /api/subscriptions` are
behind `X-API-Key` when `API_KEY` is set. Confirm and unsubscribe stay
open because they're opened from mail clients, which can't set request
headers — the token in the URL is the capability. When `API_KEY` is unset
the middleware no-ops (handy for local dev; production must set it).

**Service interfaces, handler concrete.** `Service` depends on three small
interfaces (`SubscriptionRepository`, `RepoValidator`,
`ConfirmationSender`) so business-logic tests don't need a real DB/SMTP/
GitHub. Handlers depend on the concrete `*Service` — one implementation,
tested via `httptest` against the real wired stack, so an interface there
would just be ceremony.
