# Refactoring candidates

## Purpose

This plan records the high-signal DRY candidates found by mechanically scanning
the Go, Svelte, and TypeScript sources for repeated code fingerprints, literals,
queries, and policy branches, followed by targeted inspection of the strongest
matches. It is not a request to eliminate all duplication. Repetition is worth
removing only when it represents shared policy, a shared external contract, or
substantial behavior that must change together.

The priorities are:

1. prevent authorization and lifecycle policy from drifting;
2. establish clear domain/transport boundaries where the current duplication
   exposes an architectural gap;
3. remove routine boilerplate with narrow, typed abstractions; and
4. avoid generic utility layers, repository abstractions, generated clients, or
   dependency-injection machinery that would be disproportionate for this
   application.

Unless an issue explicitly says otherwise, preserve all API payloads, status
codes, realtime topics, authorization behavior, accessible names, visual
behavior, and PocketBase query limits. Refactoring should proceed in small
reviewable changes. Use the smallest relevant checks for each issue and run
`./scripts/Test.ps1` after the architectural work has been integrated across the
codebase.

## Recommended sequence

- Implement issues 1 and 3 together because they establish the same game-policy
  ownership boundary. Preserve capability-specific errors rather than forcing
  every archived-game check through one public error.
- Implement issue 11 before or with issue 1 so actor classification and
  application authorization no longer depend on platform-owned feature policy.
- Implement issue 2 independently and migrate every equivalent production
  adapter, including transaction boundaries in profiles and rulesets.
- Implement issue 12 before issues 7 and 8 because audio cues and announcement
  pagination should move with their owning chat/announcement slice.
- Implement issue 13 before issue 14 so atomic commands have one
  transaction-composable audit writer.
- Issues 4-6 and 9-10 can be separate changes. Issues 5 and 6 should not be
  mixed with unrelated visual changes.
- Replace issue 8's offset cursor with keyset pagination after issue 12 has
  established endpoint ownership.
- Do not mark an issue complete when only one call site has migrated. Completion
  means the stated scope has been searched again and all applicable production
  occurrences use the new pattern.

## Issue 1: Put participant access policy in the domain

**Problem and evidence**

The meaning of a participant who may still access a game is encoded repeatedly
as `status != 'kicked' && status != 'left'`. It appears in realtime subscription
authorization, timer publishing, ruleset-asset access, game projections, game
handlers, and audio-cue authorization. Some code uses an equivalent Go
condition, while ability activation deliberately uses the narrower
`status = 'active'`. The distinction between "currently belongs to the game" and
"currently alive/active in play" is important and should be named rather than
left implicit in filters.

Representative locations:

- `Host/internal/platform/realtime/authorization.go`
- `Host/internal/features/timer/service.go`
- `Host/internal/features/timer/routes.go`
- `Host/internal/features/rulesets/assets.go`
- `Host/internal/features/games/service.go`
- `Host/internal/features/games/handlers.go`
- `Host/internal/features/games/chat_projection.go`
- `Host/internal/features/profiles/routes.go`

This is an architectural issue because the platform realtime package currently
owns knowledge of feature collections, relations, and participant lifecycle
states.

**Proposed change**

Create a narrow, dependency-neutral domain-policy owner, for example
`Host/internal/features/gamepolicy`. The domain package must use pure Go values
and must not depend on PocketBase `core`, `dbx`, `httpx`, route handlers,
projections, or platform realtime. Do not create a second lifecycle vocabulary:
shared game and participant status constants must have one owner, with the
existing games lifecycle using or aliasing that vocabulary. Give pure concepts
explicit names, for example:

- whether a status represents a current game member;
- whether a game status is archived or mutable; and
- the typed statuses and domain errors needed by those invariants.

Keep PocketBase-dependent operations out of that domain kernel. Put
purpose-specific operations such as finding a current participant by game and
profile, checking whether a participant record belongs to a profile, and
providing a stable composable PocketBase filter in a narrow application-policy
package adjacent to the domain policy. That package may depend on `core` and
`dbx`, but it must not become a generic repository or own response writing.

Do not use the word `active` for the broad kicked/left exclusion because an
eliminated participant still satisfies it. Keep ability eligibility separate.
Prefer pure status predicates in the domain package and purpose-specific
lookups in the application-policy package. Do not turn either package into a
generic repository or a collection of arbitrary query builders.

Move realtime subscription policy out of
`Host/internal/platform/realtime/authorization.go` into an application/feature
package that may depend on the pure domain policy and its narrow
PocketBase-backed application-policy adapter. The platform realtime package
should retain generic topic publication mechanics, event envelopes, and
authorization callback support; it should not decide what a game participant or
chat membership means. Keep topic parsing close to the subscription policy
unless it is independently useful to the generic publisher.

Migrate every applicable production query and predicate. Relation-specific
queries such as historical room access may remain specialized when sharing the
filter would require a brittle general-purpose query builder. Their
current-participant semantics must still be covered by the same policy tests.

**Arguments and constraints**

- This is security-sensitive shared knowledge, not merely duplicated query text.
- A generic repository layer would add indirection without improving the
  invariant; prefer explicit domain functions over arbitrary query builders.
- Do not collapse "current member", "active/alive player", "may post", and
  "historical room reader" into one boolean. They are distinct capabilities.
- Query failures in authorization paths must continue to fail closed.

**Architecture constraint**

`AGENTS.md` and `docs/ARCHITECTURE.md` already establish that shared domain
policy is pure and dependency-neutral, PocketBase access belongs to application
policy, and platform mechanisms do not own feature authorization. Implement
this issue within those rules; no additional package-specific repository rule
is required.

**Verification**

- Add table-driven tests for all participant statuses, explicitly covering
  eliminated, kicked, and left participants.
- Preserve realtime authorization tests for game, participant, profile, and room
  topics, including historical room access and fail-closed query behavior.
- Search production Go sources for the old compound status predicate and review
  every remaining occurrence.
- Run focused tests for realtime, games, timer, ruleset assets, profiles, and
  abilities, followed by the full gate after issues 1 and 3 are integrated.

## Issue 2: Establish one domain-error-to-HTTP boundary

**Problem and evidence**

Chat, abilities, games, and ruleset assets each contain a small adapter that
checks whether an error is a `result.AppError` and otherwise wraps it with
`result.Internal`. The implementations already differ: some use a direct type
assertion while ruleset assets uses `errors.As`, so wrapped domain errors can be
treated differently depending on the route.

Representative locations:

- `Host/internal/features/chat/routes.go`
- `Host/internal/features/abilities/routes.go`
- `Host/internal/features/games/service.go`
- `Host/internal/features/rulesets/assets.go`
- `Host/internal/platform/httpx/responses.go`

**Proposed change**

Add one function to `httpx` that accepts `error`, uses `errors.As` to recover a
possibly wrapped `result.AppError`, and otherwise writes
`result.Internal(err)`. Because `AppError.Error` currently has a value receiver,
both `result.AppError` and `*result.AppError` implement `error`; the adapter must
support both representations or the application must first standardize on one.
An `errors.As` target for only the value form is incomplete. Keep the existing
function that accepts a concrete `result.AppError` for direct validation
failures.

Migrate all equivalent feature-local adapters and transaction fallbacks and
remove them. Search the entire host for manual domain-error type assertions,
including profiles and ruleset publication, not only the four known functions.
Do not fold deliberately specialized responses such as diagnostics support
bundle failures into the generic adapter.

**Arguments and constraints**

- This is the transport boundary for public codes, messages, field errors,
  status codes, and trace IDs; behavior should not vary by feature.
- Do not add automatic logging in several layers. Unexpected-error logging and
  trace exposure should have one clearly documented owner.
- Do not make handlers return raw errors directly to PocketBase unless the
  public JSON error contract remains guaranteed.

**AGENTS.md rule to add**

Document that arbitrary feature/domain errors are translated to HTTP responses
only through the standard `httpx` error adapter. Route packages must not create
local `error -> result.AppError -> JSON` adapters.

**Verification**

- Extend `Host/internal/platform/httpx/responses_test.go` with direct and wrapped
  value-form errors, direct and wrapped pointer-form errors, and unexpected
  errors.
- Assert that expected errors omit trace IDs and internal errors retain them.
- Search for the removed adapters and direct `result.AppError` assertions.
- Run the focused `httpx`, chat, abilities, games, and ruleset-assets tests.

## Issue 3: Make archived-game immutability one invariant

**Problem and evidence**

The archived-game status check and the
`game.archived_immutable` conflict are repeated in chat room changes, roster
changes, and achievement revocation. New mutation endpoints can easily omit the
guard or change its public error contract.

Representative locations:

- `Host/internal/features/chat/routes.go`
- `Host/internal/features/games/roster.go`
- `Host/internal/features/achievements/routes.go`

**Proposed change**

Add a pure archived-status predicate or guard accepting the typed game status to
the dependency-neutral domain policy introduced by issue 1. Do not pass
`*core.Record` into the shared domain kernel. Add a narrow application-policy
guard or error constructor that extracts the record status and returns the
existing `game.archived_immutable` application error where that is genuinely
the stable public contract. Do not force capability-specific archived checks
through that error: `chat.read_only`, `achievement.not_allowed`,
`game.role_visibility_not_allowed`, and `game.transition_not_allowed`
communicate different failed capabilities and must remain stable.

Inventory all custom mutation routes for games, participants, chat,
achievements, timers, abilities, and announcements and explicitly classify each
as:

- forbidden for archived games and guarded by the shared invariant or a
  capability-specific lifecycle predicate;
- intentionally allowed after archival, with a test and explanatory comment; or
- unreachable because a stronger lifecycle precondition applies.

Do not use broad HTTP middleware for this rule. The middleware would have to
infer domain targets and mutation semantics from routes and would become
fragile.

**Arguments and constraints**

- The pure domain policy owns archived-status semantics. The application-policy
  guard owns the general archived-immutability error, not request lookup or
  response writing. Features may use the pure predicate while retaining a more
  specific public error.
- Avoid importing the `games` package from chat or achievements because games
  already imports chat. A dependency-neutral policy package prevents cycles.
- Preserve operations that are intentionally read-only or required for
  diagnostics/history.

**Architecture constraint**

`AGENTS.md` and `docs/ARCHITECTURE.md` already require cross-slice lifecycle
semantics to have a dependency-neutral domain owner. Any new mutation touching
game-owned state must apply the relevant lifecycle policy or include a test and
comment explaining why it remains valid after archival.

**Verification**

- Add a table-driven invariant test for every game status.
- Add route-level regression coverage for at least one mutation in each affected
  feature slice.
- Review all registered mutation routes. Explicitly cover intentional
  exceptions such as archived-game deletion and any post-archive announcement
  acknowledgement in tests.
- Search for direct `status == archived` mutation guards and for duplicated
  `game.archived_immutable` construction.

## Issue 4: Centralize frontend unknown-error messages

**Problem and evidence**

The SPA contains roughly fifty repetitions of
`caught instanceof Error ? caught.message : <operation-specific fallback>`.
This is boilerplate and makes future treatment of `AppApiError`, cancellation,
or non-Error throws difficult to change consistently.

**Proposed change**

Add a small typed function such as `errorMessage(caught: unknown, fallback:
string): string` in the frontend API/error module. Migrate equivalent message
extraction throughout `Web/src`.

Keep operation-specific fallback copy at each call site. Do not add a global
toast-on-error API wrapper: some failures are displayed inline, some are
background refresh failures, and some require additional recovery behavior.
Treat an empty `Error.message` as absent. Do not display arbitrary thrown
strings or objects. Cancellation that should be suppressed must be classified
by the caller before converting the error to a message.

**Verification**

- Unit-test `Error`, `AppApiError`, strings/objects, and the fallback case.
- Search `Web/src` for remaining `caught instanceof Error` expressions and
  review justified exceptions.
- Run frontend checks and focused component tests.

## Issue 5: Introduce a composable page-heading primitive

**Problem and evidence**

Many routes repeat page-heading markup, margin resets, eyebrow typography,
responsive action layout, and mobile stacking. This conflicts with the
composition guidance in `Web/DESIGN.MD` and allows small visual differences to
accumulate.

**Proposed change**

Create a semantic `PageHeading.svelte` primitive with a deliberately small API:
eyebrow, title, optional description, and an optional typed Svelte 5 actions
snippet. It should own the semantic header, `h1`, and shared presentation.
Provide only variants that correspond to established recurring layouts; do not
add arbitrary heading-level or styling escape hatches.

Migrate conventional admin and account page headings. Do not force highly
specialized focal headings—such as the live phase display or ruleset editor
toolbar—through the primitive when their semantics or layout differ.

If eyebrow text is also used outside page headings, place its typography in a
shared semantic class/token rather than recreating it in each scoped style
block.

**Verification**

- Add a component test for heading semantics, description, and optional actions.
- Verify representative phone and desktop routes, including 320px layout and
  large text.
- Search scoped Svelte styles for repeated conventional `.page-heading` and
  `.eyebrow` definitions; inspect rather than blindly removing specialized
  uses.

## Issue 6: Extract the direct-message player chooser

**Problem and evidence**

`AdminChatPage.svelte` and `PlayerChatPage.svelte` duplicate the new-message
dialog, player row, avatar initial, seat label, and all related CSS. The actual
room selection behavior and source models differ.

**Proposed change**

Extract a presentation-focused chooser that accepts normalized entries with an
ID, display label, optional supporting label, and avatar text, plus an
`onchoose` callback. Let each wrapper filter and normalize its own participant
model and perform its own asynchronous room-opening operation.

Do not merge the admin and player chat wrappers. Their authorization, navigation
paths, and room creation behavior are legitimate reasons to remain separate.
Continue composing the existing `Dialog` primitive so that focus trapping,
Escape handling, and focus restoration retain one owner. The chooser must render
a meaningful empty state and use native buttons for entries.

**Verification**

- Test dialog labeling, player selection, empty input, and keyboard-accessible
  buttons.
- Keep wrapper tests focused on the different room-opening behaviors.
- Run the relevant chat component and frontend checks.

## Issue 7: Reuse audio-cue asset resolution and payload construction

**Problem and evidence**

`publishAttentionCue` and `publishAudioCue` repeat the ruleset audio-asset query,
missing-asset behavior, and public payload construction in
`Host/internal/features/games/chat_projection.go`.

**Proposed change**

After issue 12 places announcement/audio behavior in its owning slice, extract a
package-local function there that resolves the audio asset for a game and cue
and returns the complete payload while distinguishing success, absence, and
query failure with an ordinary small Go return signature. Do not introduce an
elaborate result hierarchy for two callers. Keep the two publication functions
and their authorization callbacks distinct: attention-receipt authorization and
audience-based authorization protect different data.

Decide explicitly whether missing assets, query failures, and publication
failures should remain best-effort/silent or become observable. Do not
accidentally change behavior as a side effect of the extraction.

**Verification**

- Test successful resolution, absent assets, query failure behavior, and payload
  fields.
- Preserve separate authorization tests for attention recipients, all players,
  one player, one team, and game masters.

## Issue 8: Replace unstable offset cursors with local keyset pagination

**Problem and evidence**

The activity and announcement listing handlers currently repeat the 50-item
limit, 51-record lookahead, trimming, offset cursor calculation, and response
envelope logic in `Host/internal/features/games/activity.go`. More importantly,
both feeds sort newest-first while using offsets. A record inserted between
requests shifts later offsets and can cause records to be duplicated or
skipped. Issue 12 will move announcement listing to its owning slice, so these
endpoints must not be coupled through a games-local pagination abstraction.

**Proposed change**

Replace the opaque offset cursor with a keyset cursor containing the last
record's `created` and `id` values. Query the next page strictly after that
position in the existing `-created,-id` order. Preserve the 50-item page size,
51-record lookahead, response envelope, endpoint-specific invalid-cursor codes,
and bounded PocketBase queries.

After both endpoints have correct cursor semantics and ownership, keep their
pagination local unless a genuinely neutral opaque cursor codec remains
materially identical. Do not create a cross-application pagination framework.
The message-history cursor remains a separate abstraction because its record and
contract semantics differ.

**Verification**

- Test empty, partial, exact-50, and greater-than-50 result sets.
- Test malformed cursors and tuple ordering when multiple records share a
  timestamp.
- Insert a newer record between page requests and prove that the next page
  neither repeats nor skips older records.
- Confirm response payloads and endpoint-specific error codes are unchanged.

## Issue 9: Share game-status presentation vocabulary

**Problem and evidence**

The frontend independently maps game statuses to labels in the game list and
game layout. A new status or wording change can produce inconsistent navigation
and listing text.

**Proposed change**

Define an exhaustive map adjacent to shared frontend game presentation helpers,
using `satisfies Record<Game['status'], string>` so completeness is checked
without unnecessarily widening the inferred values. Use it everywhere the same
concise status label is intended. Callers must accept `Game['status']`; do not
retain a `string` fallback that silently displays unknown backend values.

Do not send presentation labels from the backend. Contextual sentences such as
"The game is paused" are not the same abstraction and should remain local.

**Verification**

- Let TypeScript enforce exhaustiveness.
- Search for duplicate full status maps and review contextual status copy
  separately.
- Run frontend type checks and affected tests.

## Issue 10: Provide a test-only PocketBase application fixture

**Problem and evidence**

Several Go tests repeat temporary data-directory creation, `core.NewBaseApp`,
the test encryption value, `Bootstrap`, bootstrap-state reset, and cleanup.
Lifecycle boilerplate makes it easier for future integration tests to omit
cleanup or initialize the app differently.

**Proposed change**

PocketBase v0.39.9 already provides `tests.NewTestApp`,
`tests.NewTestAppWithConfig`, migration execution, bootstrap reset, request-log
suppression, and `TestApp.Cleanup`. Add only a thin project test helper around
that supported fixture. Its narrow API should ensure project migrations are
registered, return the app shape needed by fixtures, and register cleanup with
`testing.TB`.

Keep `Host/migrations/migrations_test.go` explicit because a test of migration
registration and execution must control its own lifecycle. Do not reproduce
PocketBase's test lifecycle, and do not introduce production application
factories, repository interfaces, or dependency injection solely for tests.

**Verification**

- Migrate all feature tests that use the standard lifecycle to the thin wrapper;
  leave migration-specific tests explicit and comment why.
- Run affected packages repeatedly to detect leaked global bootstrap state or
  file handles.
- Run `go test ./Host/...`.

## Issue 11: Move actor authorization policy out of platform

**Problem and evidence**

`Host/internal/platform/auth/auth.go` owns the `game_masters` and
`player_profiles` collection names, active-account semantics, owner semantics,
and the public errors for requiring a game master, owner, or player. Feature
packages also import its collection constants for record access and
authorization branches.

These are application identity and authorization decisions, not generic
authentication plumbing. Their placement directly contradicts the rule that
platform packages must not own feature collection policy or domain
authorization.

**Proposed change**

Move actor collection vocabulary, actor classification, and the three route
guards to a narrow application identity/authorization package. It may depend on
PocketBase request/auth records and the standard HTTP error boundary because it
is application policy, not the pure shared domain kernel. Keep collection
constants and active/owner classification in one owner so feature packages do
not recreate raw collection-name and boolean predicates.

Platform may retain only genuinely generic mechanics for obtaining or carrying
an authenticated PocketBase record. If nothing generic remains,
`Host/internal/platform/auth` should be removed.

Migrate all route bindings and collection-constant consumers. Coordinate the
public-error writing with issue 2 so response codes, messages, statuses, and
envelopes remain unchanged.

**Arguments and constraints**

- Do not split the shared guards between the auth and profiles feature packages;
  every feature needs actor classification and that would replace a platform
  dependency with widespread feature-to-feature imports.
- Do not introduce an authentication service interface or dependency-injection
  container. This is a package ownership correction.
- Keep feature-resource authorization, such as game membership or room access,
  out of the actor guard. Actor type is necessary but not sufficient for those
  decisions.

**Verification**

- Add table-driven guard tests for missing auth, wrong collection, inactive
  actors, active game masters, active players, owners, and non-owner game
  masters.
- Preserve the existing unauthorized versus forbidden status distinction and
  exact public error contracts.
- Search production code for imports of `internal/platform/auth` and raw
  `game_masters`/`player_profiles` actor classification outside the new owner.
- Run focused auth, profiles, games, chat, timer, rulesets, achievements, owner,
  and diagnostics tests.

## Issue 12: Restore chat and announcement slice ownership

**Problem and evidence**

The architecture assigns chat policy, rooms, memberships, moderation, and
announcements to the chat slice, but the games slice currently owns substantial
chat state and behavior:

- `Host/internal/features/games/handlers.go` deletes chat rooms, memberships,
  messages, attention items, and receipts and builds chat-related projections;
- `Host/internal/features/games/roster.go` closes chat memberships;
- `Host/internal/features/games/transitions.go` creates role rooms and
  memberships and freezes historical access;
- `Host/internal/features/games/chat_projection.go` contains room projections,
  announcement storage and routes, attention receipts/media, and audio
  publication; and
- `Host/internal/features/games/routes.go` and `activity.go` register and list
  announcements.

Games also imports the chat feature's pure policy while continuing to implement
the persistence around that policy. Ownership is therefore split in both
directions, making cycles and policy drift increasingly likely.

**Proposed change**

Move chat-room and membership persistence, chat reader projections,
announcement/attention behavior, protected announcement media, and
announcement/audio publication behind narrow application functions owned by
the chat slice. Move announcement route registration and listing with that
behavior while preserving every URL and response contract.

Games continues to own lifecycle decisions and aggregate orchestration. When a
game starts, changes roster state, archives, resets, duplicates, or is deleted,
it should call explicit chat application operations using the caller's
`core.App` or transaction. The chat operations must be transaction-composable
and must not start nested transactions. Pass the relevant ruleset snapshot,
participants, and lifecycle facts explicitly rather than allowing chat to reach
back into games.

Keep pure chat policy dependency-neutral where practical. Do not replace direct
calls with a repository abstraction or internal event bus.

**Arguments and constraints**

- Moving files alone is not completion; games must stop knowing chat collection
  cleanup, membership, and projection mechanics.
- Game deletion/reset remains aggregate orchestration, but each affected slice
  owns how its records are changed or removed.
- This issue changes the implementation location assumed by issues 7 and 8.
  Complete it first so those refactors do not entrench the wrong owner.
- Preserve transaction scope, reader-safe projections, historical access,
  realtime topics, authorization callbacks, and query limits.

**Verification**

- Preserve route and projection tests for GM/player room visibility, sender
  labels, moderation, historical access, announcements, media, receipts, and
  audio audiences.
- Add lifecycle integration tests proving start, roster removal, archive,
  reset/duplicate, and delete still produce the correct chat state.
- Search the games package for direct access to `chat_rooms`,
  `chat_memberships`, `chat_messages`, `attention_items`, and
  `attention_receipts`; every remaining occurrence must be explicitly justified
  as aggregate orchestration rather than chat mechanics.
- Run focused games and chat tests, followed by the browser chat journey.

## Issue 13: Give audit persistence one application owner

**Problem and evidence**

`Host/internal/platform/audit/audit.go` is not generic platform infrastructure.
It owns the `game_audit` collection, actor collection classification,
application actor types and labels, target/detail fields, and request ID
projection. `Host/internal/features/games/service.go` contains a second,
substantially equivalent audit writer.

This violates both platform/domain separation and the one-owner rule. The two
writers can already drift in actor classification or persisted fields.

**Proposed change**

Create one narrow application-level audit writer and remove both existing
implementations. It may depend directly on PocketBase and must accept any
`core.App` supplied by the caller so the same function works inside a
transaction. Reuse the actor classification established by issue 11 or accept a
small typed actor snapshot; do not infer feature actor semantics independently
in the audit package.

Feature slices continue to choose their action names, targets, and safe detail
payloads. The shared writer owns only the durable audit schema and persistence.
It is not a logging framework, generic repository, or event bus.

**Arguments and constraints**

- Preserve immutable actor labels and the existing system/game-master/player
  actor types.
- Keep private or secret values out of audit detail. Consolidation must not
  broaden what callers record.
- Do not silently decide that every audit is best-effort. Issue 14 must classify
  which audits are part of a database command and which external operations
  cannot be atomic.

**Verification**

- Add tests for system, game-master, and player actor snapshots; game-scoped and
  host-scoped actions; detail; and request IDs.
- Prove the writer works with both the base app and a transaction app.
- Search for all direct `game_audit` record construction and remove duplicate
  writers.
- Run focused audit consumers across auth, profiles, rulesets, games, chat,
  achievements, and owner.

## Issue 14: Make multi-record commands atomic

**Problem and evidence**

Several lifecycle commands split one logical state change across independent
transactions or saves:

- `abilities.FinalizePhase` starts its own transaction before games or timer
  applies the related lifecycle change;
- game transitions and completion can finalize ability choices, then fail to
  save the new game state;
- timer completion can save timer state before ability finalization fails;
- game start creates chat rooms and memberships before saving the lifecycle
  transition; and
- database mutations commonly commit before their audit write, then explicitly
  discard the audit error.

These orderings permit finalized ability choices without the corresponding
phase/review transition, completed timers without finalized choices, partial
room setup, and durable changes without their required audit record. They
contradict the documented command sequence: validate, write dependent state,
revise, audit, commit, then publish.

**Proposed change**

Inventory multi-record commands and give each one an explicit transaction owner.
Make ability finalization, chat lifecycle operations from issue 12, and the
audit writer from issue 13 transaction-composable: they accept the caller's
`core.App` and do not start a nested transaction. The orchestrating game, timer,
or feature command opens one short PocketBase transaction containing all
required database writes, the revision increment, and the required audit
record. Realtime publication and response projection happen only after commit.

Provide a standalone transactional wrapper only where an operation is genuinely
invoked outside a larger command. Name transaction-required functions clearly
enough that future callers do not accidentally use the base app piecemeal.

Explicitly classify audits for operations involving external filesystem or
process effects, such as backup creation or scheduled restore. Those cannot be
made atomic with SQLite and require a documented best-effort or compensating
policy rather than a discarded error by habit.

**Arguments and constraints**

- Keep transactions short and free of network, filesystem, realtime, and other
  blocking external work.
- Do not publish from inside a transaction.
- Do not solve composition with nested transactions, a distributed transaction
  abstraction, or an event bus.
- Preserve idempotency and existing revision semantics; avoid incrementing the
  game revision twice when ability finalization joins a larger command.

**Verification**

- Add failure-path tests demonstrating that an error after dependent writes
  rolls back ability choices, game/timer state, chat lifecycle records,
  revisions, and required audits together.
- Add success tests asserting one coherent revision and publication only after
  commit.
- Review every ignored audit error and document the few intentionally
  best-effort external-operation cases.
- Search all `RunInTransaction` callers for nested transaction-composing
  operations.
- Run focused games, timer, abilities, chat, and audit tests, followed by
  `go test ./Host/...`.

| Issue | Completed | Reviewed |
| :---: | :-------: | :------: |
|   1   |    [x]    |   [x]    |
|   2   |    [x]    |   [x]    |
|   3   |    [x]    |   [x]    |
|   4   |    [x]    |   [x]    |
|   5   |    [x]    |   [x]    |
|   6   |    [x]    |   [x]    |
|   7   |    [x]    |   [x]    |
|   8   |    [x]    |   [x]    |
|   9   |    [x]    |   [x]    |
|  10   |    [ ]    |   [ ]    |
|  11   |    [x]    |   [x]    |
|  12   |    [x]    |   [x]    |
|  13   |    [ ]    |   [ ]    |
|  14   |    [ ]    |   [ ]    |

When all issues in table are fully completed and reviewed, bump patch version of app.
