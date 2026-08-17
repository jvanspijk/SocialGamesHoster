# AGENTS.md

Guidance for coding assistants working in this repository.

## Product and scope

- `docs/ARCHITECTURE.md` documents the current runtime architecture and system
  boundaries.
- The application is a trusted-private-LAN party game host for Windows 10/11 x64.
- Runtime: one console-free Go executable with PocketBase, SQLite, migrations,
  static Svelte assets, tray controls, and no external runtime dependencies.

## Repository layout

- `Host/`: Go host, vertical feature slices, migrations, fixtures, platform code,
  and the generated embedded frontend.
- `Web/`: Svelte 5 static SPA and its contract/browser tests.
- `packaging/windows/`: Inno Setup installer.
- `scripts/`: dependency, development, test, and release-build scripts.
- `docs/`: architecture, user, troubleshooting, and release guidance.
- `docs/functional-tests/`: manual, real-world UI checks that cannot be covered
  meaningfully by programmatic or Playwright tests.

## Required checks

Run `./scripts/Test.ps1` after product changes when the change warrants the full
automated gate. For a feature or fix, select the smallest relevant automated
checks instead; do not run the full suite by default. Functional tests in
`docs/functional-tests/` are time-consuming manual checks: run only the
specific scenario that is relevant, and only when real-device, Windows, LAN, or
multi-person validation is useful. Do not run any functional test merely
because a feature or fix changed.

For every change under `Web/`, run `npm run check` from `Web/` before handing off the work.

`./scripts/Test.ps1` runs:

- `go test ./Host/...` and `go vet ./Host/...`;
- Svelte/TypeScript checks;
- frontend contract tests;
- Prettier and ESLint;
- the production static build and production dependency audit;
- a focused browser journey against the compiled Go host.

Run `./scripts/Build.ps1 -Version <version>` for Windows release packaging.

## Browser debugging

This project provides a project-local Playwright MCP server for exploratory UI
debugging and feature checks. It can drive the locally running app, inspect the
accessible UI, capture screenshots, and surface browser warnings. Use it to
reproduce a reported UI issue before changing code; preserve the fix with the
existing Playwright test suite. It is development tooling only: it is not
embedded in the Go host, static frontend, or Windows installer.

For isolated local Playwright functional tests, use the disposable owner account
`playwrightadmin` with password `secret`. Never use these credentials outside a
disposable local test data directory.

## Engineering style

- For changes under `Web/`, follow the normative UI guidance in
  `Web/DESIGN.MD`, including its primitive-component composition rules.
- Treat the Go host as a modular monolith. Keep behavior in the owning vertical
  feature slice and use `Host/cmd/socialgameshoster` only as the composition
  root for wiring and process lifecycle.
- Keep platform packages generic. They may provide transport and infrastructure
  mechanisms, but must not own feature collection policy, lifecycle semantics,
  or domain authorization decisions.
- Keep shared domain policy small, dependency-neutral, and pure. PocketBase
  lookups belong to feature/application policy; do not put `core.Record`, `dbx`,
  HTTP response writing, handlers, or projections in the shared domain kernel.
- Give cross-slice lifecycle and access semantics one domain owner. Do not
  recreate participant-membership or archived-game policy as raw predicates in
  other feature or platform packages.
- Treat feature-to-feature imports as exceptional. Prefer a neutral domain
  policy, narrow application contract, or composition-root callback when
  behavior crosses slices.
- Keep handlers as an imperative shell around deterministic policy: authorize,
  decode, load, validate, transact, project, then publish after commit.
- Keep PocketBase collection APIs locked; expose domain behavior through custom
  routes and reader-safe projections.
- Use explicit result/error contracts and correct HTTP status codes.
- Translate arbitrary feature and domain errors to HTTP responses only through
  the standard `httpx` error adapter. Route packages must not define local
  `error -> result.AppError -> JSON` adapters.
- Keep secrets, private roles, hidden achievements, anonymous identities, chat,
  history, and diagnostics out of unauthorized projections and events.
- Keep PocketBase and runtime-sensitive dependencies exact-pinned in `go.mod`.
  Keep frontend dependencies that affect API behavior exact-pinned in
  `Web/package.json`.
- Keep the Svelte frontend strict, accessible, responsive, and compatible with
  reduced motion.
- Prefer small functions and avoid heavy dependencies. Do not introduce generic
  repositories, dependency-injection containers, command or event buses, CQRS
  frameworks, or persistence-independent mirror models without a demonstrated
  need.

## Test philosophy

- Test externally observable contracts and business invariants, not file layout,
  call order, private helper structure, generated markup, or other implementation
  details.
- Prefer role/name/label-based UI assertions over element indexes or CSS
  selectors.
- Use deterministic fixtures and semantic assertions.
- Keep the automated suite proportionate to a friends-only LAN application.
  Thirty-player capacity, physical-phone behavior, and clean-VM installation are
  manual release checks documented in `docs/RELEASE_VALIDATION.md`.
