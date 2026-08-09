# Architecture

## Runtime

The installed product is one console-free Windows x64 executable. Go embeds the
content-hashed output of the static Svelte SPA. PocketBase provides the HTTP
router, SQLite-backed collections, file storage, authentication, logging,
backups, and one SSE connection per browser. There are no external runtime
dependencies.

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

## Architectural model

The host is a modular monolith. It is deployed as one process and one database,
but its Go code is divided into vertical feature slices with explicit dependency
direction. This preserves local reasoning and domain ownership without adding
network boundaries, service orchestration, or distributed consistency.

`Host/cmd/socialgameshoster` is the composition root. It constructs the
PocketBase application, registers platform mechanisms and feature routes,
connects callbacks, starts background jobs, and owns process lifecycle. Wiring
belongs there; business policy does not.

The backend follows a functional-core, imperative-shell style:

- Pure functions own status semantics, lifecycle transitions, validation, and
  other deterministic invariants wherever practical.
- Handlers and services form the imperative shell: authenticate, decode, load
  records, invoke policy, perform a short transaction, project a reader-safe
  result, and publish only after commit.
- Commands and queries remain visibly separate. Commands mutate, revise, audit,
  commit, and publish. Queries use bounded reads and explicit projections.
- These are code-organization rules, not framework requirements. The
  application does not use a command bus, mediator, domain-event bus, or CQRS
  framework.

## Dependency direction

Dependencies point from the composition root into feature/application code and
from feature/application code into small domain policies and generic platform
mechanisms:

```text
composition root
    |-- feature slices ---+-- pure shared domain policy
    |                     |-- generic platform mechanisms
    |                     `-- PocketBase persistence
    `-- platform setup ------- PocketBase
```

- Feature slices own feature routes, commands, queries, projections,
  authorization decisions, and public error contracts.
- Generic `platform` packages may provide HTTP response mechanics, middleware,
  authentication plumbing, realtime publication, event envelopes, tracing, and
  desktop integration. They must not own feature collection names, game or
  participant lifecycle semantics, room-membership policy, or other domain
  authorization decisions.
- A shared domain-policy package is justified only for stable invariants that
  genuinely cross feature boundaries. It stays small and uses pure Go values;
  it does not depend on PocketBase records, `dbx`, HTTP helpers, handlers, or
  projections.
- PocketBase-dependent lookups and composed filters belong to the relevant
  feature/application policy. When several slices need the same lookup, use a
  narrow shared application-policy helper adjacent to the pure domain policy,
  rather than putting persistence into the domain kernel or creating a generic
  repository.
- Feature-to-feature imports are exceptional because they couple slice
  lifecycles and can create cycles. Prefer a dependency-neutral domain policy,
  a narrow application-level contract, or composition-root wiring when behavior
  must cross slices. Do not introduce an event bus solely to avoid an ordinary
  function call.

PocketBase is an intentional application dependency, not an implementation that
must be hidden behind generic repositories. Use typed domain values where they
protect correctness, while allowing PocketBase records at the imperative
application boundary. Do not add repository interfaces, dependency-injection
containers, persistence-independent mirror entities, or generated clients
without a concrete second implementation or other demonstrated need.

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
- `platform`: auth plumbing, error envelopes, middleware, generic realtime
  mechanisms, tracing, and Windows desktop integration.

Cross-slice policy is not a new general-purpose feature. Participant membership
semantics and archived-game invariants are examples that may justify a small
shared domain owner; route behavior, projections, room creation, timer commands,
and feature-specific authorization remain in their slices.

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

Every subscription is validated by application authorization policy registered
against the generic PocketBase realtime hook:

- `game:<id>:public`: active participants and game masters;
- `game:<id>:game-masters`: active game masters only;
- `participant:<id>:private`: that profile or a game master;
- `room:<id>`: current/historical room membership or a game master;
- `profile:<id>`: that profile or a game master;
- `profile-request:<id>:<capability-hash>`: exact request capability;
- `profile-requests:game-masters`: active game masters only.

Events are reader-safe projections in an ID/kind/time/revision envelope. Clients
deduplicate IDs, refresh a snapshot after gaps, and treat the database snapshot
as authoritative. The platform realtime package owns broker interaction,
publication, envelopes, and callback execution. Feature/application policy owns
topic meaning and reader eligibility.

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
