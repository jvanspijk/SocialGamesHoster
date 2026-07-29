# Legacy parity audit

This audit is the removal gate for the former .NET, PostgreSQL, SignalR, and
server-rendered SvelteKit implementation. It compares user capabilities rather
than preserving old endpoint shapes. Database and API compatibility with the
legacy implementation were deliberately excluded.

## Capability mapping

| Legacy capability or TODO | Current result | Evidence |
|---|---|---|
| Administrator login | Replaced by named game-master accounts, 12-hour tokens, active-account checks, and owner-only account management | `Host/internal/features/auth/` |
| Player login/name entry | Replaced by reusable passwordless profiles, explicit GM approval, one-time recovery capabilities, replay protection, and token rotation | `Host/internal/features/profiles/` |
| Current-player lookup | Replaced by `/profiles/me` and the membership-scoped live player view | `Host/internal/features/profiles/routes.go`, `Host/internal/features/games/handlers.go` |
| Create/read/update/delete players | Split into reusable profile moderation and per-game participant join, alias, kick, eliminate, and reinstate actions | `Host/internal/features/profiles/`, `Host/internal/features/games/roster.go` |
| Role and ability CRUD | Replaced by versioned aggregate ruleset drafts with a guided visual editor | `Host/internal/features/rulesets/`, `Web/src/lib/components/rulesets/VisualDefinitionEditor.svelte` |
| Role knowledge | Replaced by selector-based knowledge rules whose projection exposes only configured identity, role, team, and elimination fields | `Host/internal/features/rulesets/contracts.go`, `Host/internal/features/games/projection.go` |
| Ruleset list/detail | Replaced, including draft creation, autosave, validation, immutable publishing, successor versions, archive, duplicate, and guarded delete | `Host/internal/features/rulesets/routes.go`, `Web/src/routes/admin/rulesets/[id]/+page.svelte` |
| Persist ruleset creation | Completed; drafts are durable SQLite records and unfinished new definitions are retained locally until first save | `Host/internal/features/rulesets/routes.go`, `Web/src/routes/admin/rulesets/[id]/+page.svelte` |
| Portable rulesets | Added versioned `.sghrules` ZIP import/export with manifest, checksums, limits, images, and audio | `Host/internal/features/rulesets/bundle.go` |
| Demonstration seed | A validated, importable Blackjack demonstration bundle is included | `Host/fixtures/` |
| Game create/list/detail | Replaced with draft ledger, GM view, player-safe projection, and exactly-one-live invariant | `Host/internal/features/games/handlers.go`, `Host/migrations/1710000000_initial.go` |
| Duplicate/delete games | Completed with immutable snapshot duplication and typed, state-limited deletion of application-owned dependents | `Host/internal/features/games/handlers.go` |
| Ruleset selection | Ruleset version is selected at draft creation and snapshotted before play | `Host/internal/features/games/handlers.go` |
| Start/stop game | Expanded to draft → lobby → running ↔ paused → review → immutable archive | `Host/internal/features/games/lifecycle.go`, `Host/internal/features/games/transitions.go` |
| Start/current/finish round | Replaced with ordered phases, explicit round-start phases, phase revision events, and manual GM advancement | `Host/internal/features/games/transitions.go` |
| Winner selection | Replaced with per-participant win/loss/draw review before archive | `Host/internal/features/games/roster.go` |
| Player name changes | Game aliases can be changed without mutating reusable profile identity | `Host/internal/features/games/roster.go` |
| Player removal/kicking | Completed with participant status change, room access loss, and scoped realtime updates | `Host/internal/features/games/roster.go`, `Host/internal/platform/realtime/authorization.go` |
| Manual role assignment | Completed with composition validation and immediate private projections | `Host/internal/features/games/roster.go` |
| Random role assignment | Added deterministic, constrained backtracking with locks, uniqueness, dependencies, exclusions, bands, and modifiers | `Host/internal/features/rulesets/solver.go` |
| Global in-memory timer | Replaced by per-live-game durable timer state with start/pause/resume/adjust/stop/completion and restart/sleep reconciliation | `Host/internal/features/timer/` |
| Timer display/control TODOs | Completed for player and GM views with server-time correction and completed-timer semantics | `Web/src/lib/components/TimerDisplay.svelte`, `Web/src/lib/components/AdminGamePanel.svelte` |
| Global chat | Replaced by ruleset-controlled general rooms with cursor history and reader-safe message events | `Host/internal/features/chat/` |
| Team chat | Completed with computed role-team membership and phase policy overrides | `Host/internal/features/chat/policy.go` |
| Player DMs | Completed as ruleset-controlled two-player rooms | `Host/internal/features/chat/routes.go` |
| GM DMs | Added as always-available participant/GM rooms | `Host/internal/features/games/service.go`, `Host/internal/features/chat/` |
| Announcements | Added as an always-available GM broadcast room with recent history | `Host/internal/features/games/chat_projection.go` |
| Sender anonymity/labels | Generalized to profile, alias, seat, role, or team labels; source identity is omitted from unauthorized payloads | `Host/internal/features/chat/policy.go` |
| Chat moderation | Added GM read access disclosure, room locks, soft deletion, and immutable historical access windows | `Host/internal/features/chat/` |
| One realtime wrapper per feature | Replaced by one PocketBase SSE client and capability-authorized topics | `Web/src/lib/api/client.ts`, `Host/internal/platform/realtime/` |
| Chat events requiring a follow-up fetch | Replaced by complete reader-safe message payloads in the committed event | `Host/internal/features/chat/routes.go` |
| Role privacy toggle | Preserved as a one-tap, accessible conceal/reveal card without chained animation | `Web/src/lib/components/RoleCard.svelte` |
| Notifications | Replaced by live announcements, visual connection state, unread room counts, and opt-in targeted audio cues | `Web/src/routes/play/+page.svelte`, `Web/src/lib/components/AdminGamePanel.svelte` |
| Achievement awards | Added durable ruleset-defined award/revoke, snapshot text, spoiler-safe live visibility, and global achievement points | `Host/internal/features/achievements/`, `Host/internal/features/profiles/routes.go` |
| Archives and history | Added immutable archives, private role/game history, party-visible aggregate statistics, and historical chat bounds | `Host/internal/features/profiles/routes.go`, `Host/internal/features/chat/routes.go` |
| Admin debug/log viewer | Replaced by owner-only opt-in diagnostics and a redacted support ZIP; core error logging remains always on | `Host/internal/features/diagnostics/` |
| Plain INI admin credentials and hard-coded JWT | Removed; PocketBase auth records, password hashes, token keys, and custom auth routes are used | `Host/internal/features/auth/`, `Host/migrations/1710000000_initial.go` |
| API failures returning HTTP 200 | Replaced by a standard error envelope and explicit 4xx/5xx statuses | `Host/internal/platform/httpx/`, `Host/internal/platform/result/` |
| PostgreSQL/.NET/Node runtime setup | Replaced by embedded SQLite, one console-free Go executable, static embedded Svelte assets, and an Inno Setup installer | `Host/cmd/socialgameshoster/`, `Host/embedded/`, `packaging/windows/installer.iss` |
| Separate install/run batch scripts | Replaced by Windows installer/tray behavior and developer-only PowerShell scripts | `packaging/windows/`, `scripts/` |

## Deliberate non-parity

- Endpoint URLs, generated SDK files, SignalR hub names, Entity Framework models,
  and PostgreSQL migrations are not retained.
- Automatic ability execution and arbitrary ruleset scripting are not added; the
  GM resolves game rules manually.
- The old seeded running game and test players are not retained. Only reusable
  example ruleset bundles are seeded.
- Old visual animation implementations are not retained. Motion is short,
  CSS-based, and disabled by reduced-motion preferences.

## Removal conclusion

Every legacy user-facing capability is either mapped above or deliberately
excluded. The legacy projects and runtime scripts are therefore safe to delete
after the current automated suites and packaging build pass.
