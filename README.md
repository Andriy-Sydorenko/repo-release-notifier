# repo-release-notifier

## Architecture Notes

### JSON Tags in GORM Models

GORM model structs use `json:"-"` on internal fields (`ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt`) to exclude them from JSON serialization — these are database-only fields that should never be exposed through the API.

Fields like `Email` and `Repo` carry explicit `json:"email"` tags to control JSON key naming when serialized.

In practice, the API layer uses dedicated `SubscriptionResponse` DTOs (defined in `schema.go`) for all responses, so these tags are a defensive measure — they prevent accidental data exposure if a model is marshaled directly instead of being converted to a DTO first.

### Service Dependencies: Interfaces, Not Concrete Types

`Service` depends on three small interfaces — `SubscriptionRepository`, `RepoValidator`, `ConfirmationSender` — rather than the concrete `*Repository`, `*GitHubClient`, `*Notifier` structs.

**Why:**

- **Testability.** Unit tests for business logic must not touch a real database, the GitHub API, or an SMTP server. Interfaces let tests inject lightweight in-memory fakes (see `service_test.go`) that record calls and return canned responses. Concrete types would force integration tests or runtime stubbing via build tags.
- **Dependency direction.** Interfaces are declared in the consumer (`service.go`), not the implementer. The service defines *what it needs*; the repository, GitHub client, and notifier simply satisfy those shapes structurally. This keeps the service layer independent of GORM, `net/http`, and `net/smtp` — swapping any implementation (e.g., Redis-cached GitHub client, mock SMTP for local dev) requires no change to service code.
- **Minimal surface.** Each interface lists only the methods the service actually calls. `SubscriptionRepository` doesn't expose scanner-only methods like `FindDistinctConfirmedRepos`; `ConfirmationSender` doesn't expose `SendReleaseNotification`. This prevents accidental coupling and makes the dependency graph readable.

**Why not interfaces everywhere?** Handlers depend on the concrete `*Service` — there's one implementation, no test-time substitution needed, and handlers are tested via `httptest` against the real wired stack. Adding an interface there would be ceremony without benefit (per the "no abstractions without 3+ consumers" rule).

### Zero-Width Space in "Copy this link" URLs

The HTML confirmation email shows the confirmation URL twice: once inside the clickable button and once as plain text for copy-paste. Mail clients (Gmail, Apple Mail, Outlook web) aggressively auto-linkify any bare URL they see, turning the "copy" version back into a clickable link — which defeats its purpose and leads to double-click confusion.

To suppress auto-linkification, a U+200B (zero-width space) is inserted between the URL scheme and `://` before rendering (`breakAutoLink` in `notifier.go`). Clients' URL detectors match `https?://`, and the ZWSP breaks that pattern, so the text renders as plain text. When users copy the URL, browsers strip the ZWSP on paste into the address bar, so the link still works.

Tradeoff: in a handful of niche clients the ZWSP may survive copy-paste. Acceptable for a fallback that most users will never use.
