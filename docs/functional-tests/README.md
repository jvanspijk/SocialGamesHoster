# Manual functional UI tests

These are short, human-run checks for behavior that cannot be meaningfully
verified through unit tests, contract tests, or automated Playwright browser
tests. They complement, rather than duplicate, the automated suite.

## Required visual and usability checklist

Complete this checklist while performing every scenario in this suite:

- Confirm no unintended visual issues: elements do not obscure, overlap, clip, or cut off text or controls.
- Confirm clear visual hierarchy, discoverable actions, and an efficient task path without unnecessary steps.

Complete this checklist once per page while performing a scenario in the suite:
- Inspect the initial layout, then resize the viewport and optionally rotate a phone where applicable; it remains visually polished and usable.

Run a scenario only when its environment is affected or when preparing a
release. Use a private LAN, disposable test profiles, and a non-production
backup. Record the application version, Windows version, browser/device, and
outcome with any failure report.

| Page or environment | Scenarios |
| --- | --- |
| Join page (`/`) | [join-page.md](pages/join-page.md) |
| Game-master dashboard (`/admin`) | [admin-dashboard.md](pages/admin-dashboard.md) |
| Player game page (`/play`) | [player-game-page.md](pages/player-game-page.md) |
| Ruleset editor (`/admin/rulesets/:id`) | [ruleset-editor.md](pages/ruleset-editor.md) |

The existing automated browser suite owns ordinary form validation, navigation,
and deterministic workflow checks. Do not add those cases here.
