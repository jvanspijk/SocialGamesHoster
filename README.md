# Social Games Hoster

Social Games Hoster is a local-first party game host for Windows 10 and 11. One
computer runs the application; players join from phones and laptops on the same
trusted private network. Internet access, cloud accounts, Docker, PostgreSQL,
.NET, and Node.js are not required at runtime.

The rebuild is a single Go executable containing:

- PocketBase `v0.39.9`, pinned exactly, with SQLite storage and SSE realtime;
- compiled migrations and immutable ruleset/game snapshots;
- a static Svelte 5 application embedded into the executable;
- a Windows tray, single-instance guard, local QR codes, and automatic backups.

## Install and host

1. Download the Windows x64 setup executable from the GitHub Releases page.
2. Run the installer. Windows may show a SmartScreen reputation warning for an
   unsigned community build; verify the adjacent SHA-256 checksum and download
   source before continuing. This differs from an antivirus malware detection.
3. Keep the Windows network category set to **Private**. The installer never
   adds a Public-network firewall rule.
4. Social Games Hoster opens the setup page. Create the first owner account.
5. Open a lobby from the Host dashboard and let players scan the shown QR code.

The tray menu can open the dashboard or join page, copy the join link, show the
QR code, start or stop hosting, create a backup, and exit cleanly. Starting the
application a second time opens the existing dashboard instead of a second
database process.

Read [the user guide](docs/USER_GUIDE.md) and
[troubleshooting guide](docs/TROUBLESHOOTING.md) for operational details.

## Security model

This is a trusted-LAN application, not an internet-facing service. Bind it only
to a network you trust. Custom routes enforce account type, active state,
game/room membership, and private projections. PocketBase collection CRUD and
its administrative dashboard are locked. The installer firewall rule is scoped
to Windows Private networks.

Secret roles, anonymous sender identities, private game history, chat bodies,
authentication material, and diagnostics are never included in public
projections. Detailed diagnostics exist only when launched with
`--diagnostics`, and remain owner-only.

## Data and recovery

Mutable data is stored under:

```text
%LocalAppData%\SocialGamesHoster\data
```

Uninstall preserves this directory unless the separately worded
“permanently delete all data” option is selected and confirmed. The application
creates:

- a safety backup before a version migration;
- one automatic backup on the first active launch each day;
- owner-triggered backups from the dashboard or tray;
- a rollback backup before restore.

Seven automatic daily backups are retained. Restore requires the owner to type
the full backup-specific confirmation phrase.

## Development

Requirements:

- Go 1.25 or newer;
- Node.js 24 and npm (build-time only);
- Inno Setup 6 (installer builds only).

```powershell
./scripts/Install-DevDependencies.ps1
./scripts/Test.ps1
./scripts/Dev.ps1
./scripts/Build.ps1 -Version 0.2.3
```

`Dev.ps1` starts the Go host on port 8090 and the Svelte development UI on port
9091 with a local API proxy. `Build.ps1` verifies the projects, embeds the static
web output, produces a console-free Windows x64 executable, builds the Inno
Setup installer, and writes SHA-256 checksums.

Tests assert user-visible and API contracts rather than component internals.
Clean-VM installation, physical-phone QR joining, and a 30-player party
rehearsal are optional field-validation guides. They are useful before wider
distribution but do not block a friends-only release.

For signed releases, provide `SGH_SIGN_CERT_THUMBPRINT` and make
`signtool.exe` available. The build signs and timestamps both the application
and installer. Use the same trusted publisher identity for every release;
self-signed certificates do not improve SmartScreen reputation. Microsoft's
Artifact Signing service is the preferred non-Store option, but requires a
separate account and identity-verification setup. Without a trusted certificate,
builds remain functional but Windows may warn users.

Never ask users to disable Defender. If Defender identifies a clean release as
malware or potentially unwanted software, submit that exact artifact to the
[Microsoft malware-analysis portal](https://www.microsoft.com/en-us/wdsi/filesubmission)
as a software developer and wait for the classification result before publishing
it broadly. A normal "unrecognized app" SmartScreen prompt is reputation-based
and is not corrected through the malware appeal process.

See [architecture](docs/ARCHITECTURE.md) and
[release validation](docs/RELEASE_VALIDATION.md) for implementation and release
gates.

## License and source

Social Games Hoster is licensed under the
[GNU Affero General Public License v3](LICENSE). The running interface links
back to this source repository as required for network use.
