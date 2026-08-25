# Release validation

This document records automated release gates and optional field-validation
guides. The automated gates are sufficient for a friends-only LAN release.
Clean-VM, physical-device, and large-party checks can be performed when the
relevant hardware and group are available.

## Automated gates

Run:

```powershell
./scripts/Test.ps1
./scripts/Build.ps1 -Version <version>
```

Required outcomes:

- all Go unit/integration tests and `go vet` pass;
- Svelte type checking, ESLint, Prettier, static build, and high-severity npm
  production-dependency audit pass;
- the focused setup, ruleset persistence, and mobile-layout browser journey passes;
- the ruleset workflow's keyboard, landmark/name/status, visible-focus, 320 px,
  200% zoom, large-text, high-contrast, and reduced-motion browser checks pass;
- PocketBase remains pinned exactly;
- the Windows GUI-subsystem build succeeds with `CGO_ENABLED=0`;
- `SocialGamesHoster.exe`, one setup executable, and `SHA256SUMS.txt` exist;
- no external runtime files are packaged.

## Signing and reputation

For a public Windows release:

- sign and timestamp both `SocialGamesHoster.exe` and the setup executable with
  the same trusted publisher identity;
- verify both artifacts report `Status: Valid` with
  `Get-AuthenticodeSignature`;
- do not modify either artifact after signing;
- publish the SHA-256 checksum over HTTPS beside the download;
- expect that a newly signed app may still show an "unrecognized app"
  SmartScreen reputation prompt while its reputation is new;
- never tell users to disable Defender or SmartScreen.

A self-signed certificate is suitable only for local signature testing and does
not improve SmartScreen reputation. Microsoft Store distribution is the only
reliable route to avoiding SmartScreen download warnings entirely. For direct
downloads, evaluate Microsoft Artifact Signing (formerly Trusted Signing)
instead of buying an EV certificate solely for SmartScreen.

If Defender reports a specific malware or potentially-unwanted-application
detection, stop the release and submit the exact detected artifact through the
[Microsoft malware-analysis portal](https://www.microsoft.com/en-us/wdsi/filesubmission).
Record the detection name, Defender definition version, artifact SHA-256, and
Microsoft's determination in the release notes. This process is for actual
detections, not the ordinary SmartScreen reputation prompt.

## Security matrix

Verify with guest, approved nonparticipant, participant A, participant B, game
master, inactive game master, owner, and disabled profile:

- raw PocketBase collection CRUD and dashboard routes are unavailable;
- cross-game participant, role, room, message, history, asset, and diagnostic
  IDs do not disclose existence or content;
- player subscriptions to game-master topics are rejected;
- room events contain the complete reader-safe message and no hidden real
  anonymous identity;
- public JSON never includes other players’ roles;
- expired/replayed profile capabilities fail and recovery invalidates old tokens;
- CSP permits only bundled assets and same-origin connections;
- logs/support ZIP contain no passwords, bearer tokens, capability secrets, chat
  bodies, profile content, or private roles.

## Optional clean Windows VM matrix

Use supported Windows 10 x64 and Windows 11 x64 VMs without development tools
or application runtime dependencies installed.

- install and optionally launch;
- confirm the inbound rule is TCP, selected port, Private profile only;
- create owner and named game master;
- verify tray commands and second-launch activation;
- test a path with spaces and a non-ASCII Windows username;
- test an occupied port and adapter change;
- update while running and confirm data/pre-migration backup preservation;
- launch normal and Diagnostic Mode shortcuts;
- uninstall once preserving data and reinstall;
- uninstall again with the separately worded data checkbox and confirmation.

## Optional end-to-end party scenario

At iPhone Safari and common Android viewport sizes:

1. request/approve a new profile and recover it on another browser;
2. edit biography/accent/avatar;
3. create, preview, save as Valid, export, and re-import a ruleset with image/audio;
4. create/open a game and join 30 profiles;
5. manually assign, randomize, start, and verify private role/knowledge;
6. exercise phase and every timer transition, including sleep/restart;
7. exercise announcements with saved and one-off media, GM DMs,
   general/team/player rooms, locks, anonymity, moderation, deletion, unread
   behavior, recipient-only attachment reads, and sound opt-in;
8. eliminate/reinstate, enter outcomes, award/revoke achievements;
9. review/archive and verify public stats/private history;
10. create/restore a backup and collect a redacted support bundle.

Scan the displayed QR from at least one physical phone on the same private
network.

The 30-player scenario is an optional real-world field check. A synthetic
32-client SSE harness is not part of the required automated suite: it added
substantial transport-fixture maintenance for little confidence beyond a real
friends-on-one-router rehearsal. Stable authorization, shared-LAN rate limiting,
policy, lifecycle, and persistence contracts remain automated.

## Performance record

Reference minimum: two logical CPU cores, 4 GB RAM, low-end SSD/eMMC, Windows
10/11 x64. Record date, CPU, RAM, storage, Windows build, browser, commit, and
application version.

| Gate                         |   Target |               Recorded result |
| ---------------------------- | -------: | ----------------------------: |
| Host ready/dashboard request |    < 2 s |    deferred field measurement |
| Idle working set             |  < 75 MB |    deferred field measurement |
| Idle CPU after settle        |     < 1% |    deferred field measurement |
| 30-player working set        | < 150 MB |    deferred field measurement |
| Uncached LAN interactive     |    < 1 s |    deferred field measurement |
| Cached navigation            | < 300 ms |    deferred field measurement |
| Commit-to-event p95          | < 200 ms |    deferred field measurement |
| Database lock errors         |        0 |    deferred field measurement |
| Installer size               |  < 50 MB | produced and checked by build |
| Installed executable/assets  |  < 80 MB | produced and checked by build |

These measurements are not blockers for a friends-only release. Record them
with future release notes when a suitable VM, physical phone, or full party is
available, and investigate any missed target before changing the target.
