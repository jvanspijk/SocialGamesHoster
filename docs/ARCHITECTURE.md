# Architecture

## Runtime

The installed product is one console-free Windows x64 executable. Go embeds the
content-hashed output of the static Svelte SPA. PocketBase provides the HTTP
router, SQLite-backed collections, file storage, authentication, logging,
backups, and one SSE connection per browser. There is no runtime Node.js,
.NET, PostgreSQL, Docker, nginx, or internet dependency.

```text
phone / browser
    │ HTTP + one SSE connection
    ▼
Go host ── custom /api/app/v1 handlers ── PocketBase
    │                                      │
    ├── embedded static Svelte SPA         ├── SQLite collections
    ├── tray + single-instance lock        ├── protected files
    └── timer scheduler                    └── backups/logs
```

## Vertical slices

- `Host/internal/features/setup`: loopback first owner and join QR.
- `auth`, `profiles`: named game masters and approved passwordless profiles.
- `rulesets`: contract, validation, solver, bundles, and protected assets.
- `games`: lifecycle, roster, assignments, projections, outcomes, and audit.
- `timer`: persisted transitions plus one cancellable completion scheduler.
- `chat`: policy, rooms, sender-safe projections, moderation, announcements.
- `achievements`: immutable award snapshots, spoiler visibility, and points.
- `owner`: host settings, network selection, backups, and restore.
- `diagnostics`: opt-in owner-only health/log/resource/support routes.
- `platform`: auth guards, error envelopes, middleware, realtime authorization,
  and Windows desktop integration.

## Data boundaries

All domain collection API rules are locked. Generic PocketBase CRUD and the
PocketBase dashboard are not exposed. Custom handlers verify both the auth
collection and active state, then authorize the specific record or scope.

The published ruleset definition is copied into each game at creation. Starting
a game requires a composition-valid assignment. A partial unique SQLite index
enforces one live game (`lobby`, `running`, or `paused`) even under concurrent
requests. Archived game snapshots, sender labels, role/outcome history, and
achievement titles remain stable when a later ruleset version changes.
Achievement point values and spoiler visibility are also snapshotted at award
time. A hidden award is omitted from player events and profile totals until its
game reaches Review or Archive.

## Realtime

The broker validates every subscription:

- `game:<id>:public`: active participants and game masters;
- `game:<id>:game-masters`: active game masters only;
- `participant:<id>:private`: that profile or a game master;
- `room:<id>`: current/historical room membership or a game master;
- `profile:<id>`: that profile or a game master;
- `profile-request:<id>:<capability-hash>`: exact request capability;
- `profile-requests:game-masters`: active game masters only.

Events are reader-safe projections in an ID/kind/time/revision envelope. Clients
deduplicate IDs, refresh a snapshot after gaps, and treat the database snapshot
as authoritative.

## Transactions and concurrency

State-changing handlers keep SQLite transactions short: validate, write,
increment the game revision, append audit, commit, then publish. No mutable game
cache exists. Read projections use bounded queries. Timer completion is the only
domain scheduler and holds one cancellable timer for the current running game;
there are no per-client or per-second server ticks.

## Dependency policy

PocketBase and runtime-sensitive dependencies are exact-pinned in `go.mod`.
Frontend runtime dependencies that affect API behavior are exact-pinned in
`package.json`. Upgrades require changelog review, full tests, a clean-database
migration run, and a pre-upgrade backup test.
