# Manual functional UI tests

These are short, human-run checks for behavior that cannot be meaningfully
verified through unit tests, contract tests, or automated Playwright browser
tests. They complement, rather than duplicate, the automated suite.

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
