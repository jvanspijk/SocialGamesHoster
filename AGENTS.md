# AGENTS.md

Guidance for coding assistants working in this repository.

## Product and scope

- `REBUILD_PLAN.md` is the product and architecture source of truth.
- The application is a trusted-private-LAN party game host for Windows 10/11 x64.
- Runtime: one console-free Go executable with PocketBase, SQLite, migrations,
  static Svelte assets, tray controls, and no external runtime dependencies.
- Do not reintroduce the legacy .NET, PostgreSQL, SignalR, Node-server, Docker, or
  generated-client architecture.

## Repository layout

- `Host/`: Go host, vertical feature slices, migrations, fixtures, platform code,
  and the generated embedded frontend.
- `Web/`: Svelte 5 static SPA and its contract/browser tests.
- `packaging/windows/`: Inno Setup installer.
- `scripts/`: dependency, development, test, and release-build scripts.
- `docs/`: architecture, user, troubleshooting, parity, and release guidance.
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

- Keep backend changes in the relevant vertical feature slice.
- Keep PocketBase collection APIs locked; expose domain behavior through custom
  routes and reader-safe projections.
- Use explicit result/error contracts and correct HTTP status codes.
- Keep secrets, private roles, hidden achievements, anonymous identities, chat,
  history, and diagnostics out of unauthorized projections and events.
- Pin PocketBase and the browser SDK to the versions in `REBUILD_PLAN.md`.
- Keep the Svelte frontend strict, accessible, responsive, and compatible with
  reduced motion.
- Prefer small functions and avoid heavy dependencies.

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
