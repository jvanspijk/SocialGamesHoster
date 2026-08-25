# Social Games Hoster user guide

## First launch

Social Games Hoster appears in the Windows notification area and opens a browser
automatically. On a new installation:

1. Create the owner username, display name, and a password of at least ten
   characters.
2. Review the trusted-LAN notice. The host is visible to devices on the selected
   Windows Private network.
3. The Blackjack demonstration ruleset is installed automatically.
4. Open **Host → Installation** to confirm the selected adapter and port.

Use named game-master accounts for every host. Actions are attributed in the
game audit; accounts should not be shared.

## Prepare a game

1. Open **Host → Games** and create a game from a valid saved ruleset.
2. Select **Open lobby**. Only one game can be live.
3. Show the phone QR from the Installation page or tray. The link contains only
   the private IP address and port, never an account secret.
4. A new or recovering player chooses their profile name. Approve the request
   from **Entry requests** in the live-game header, or under **Approvals**.
5. Assign seats, aliases, and roles manually or use constrained randomization.
   The roster must satisfy the selected composition band before the game starts.

Returning approved devices keep their profile. Recovering that profile on a new
device requires approval and invalidates its old device token.

## Run the game

The Live Table provides lifecycle, roster, role, phase, timer, announcement,
chat, outcome, and achievement controls.

- Players see only their own role and allowed knowledge.
- The persisted timer continues correctly across sleep or restart.
- Announcements and game-master direct messages are always available.
- General, team, and player rooms follow the game's saved ruleset and current
  phase. Private/team rooms disclose that game masters can read them.
- Anonymous display affects player projections only; game masters retain the
  attributed sender.
- Sound is opt-in on each player browser. Every cue also has a visual event.

Move a completed game to Review, enter an outcome (or explicit unset choice) for
each active participant, then Archive it. Archive freezes the historical
snapshot and updates profile statistics and private histories.

Achievements may award zero or more global achievement points. A ruleset author
can hide an achievement until the game reaches Review, which prevents both the
award and the changed point total from spoiling an ongoing game. Game masters
can still see and revoke hidden awards from the Live Table.

## Rulesets

The guided creator covers basics, teams, roles and abilities, player setup,
game flow, information rules, chat, rewards, and image/audio media. Work is
stored only when you choose **Save**; local recovery can restore unsaved browser
changes after an interruption.

A saved ruleset is marked **Valid** or **Invalid**. Valid rulesets can be chosen
for new games. Invalid rulesets remain editable but are unavailable for new
games until their blocking issues are fixed and saved. Use **Preview** to check
role cards, phase flow, player setup, chat availability, and media in context.
Use **Edit**, **Save**, and **Delete ruleset** to manage the library.

`.sghrules` exports include checksums, provenance, the definition, and declared
media. Imports create a separate saved ruleset and never overwrite an existing
one.

## Profiles

Players can edit their display name, short biography, accent, and a validated
JPEG/PNG/WebP avatar. Public party summaries contain aggregate statistics and
achievement snapshots. Role-by-game history remains visible only to the profile
owner and game masters where explicitly authorized.

## Backups and restore

Open **Host → Installation** or use the tray to create a backup. The dashboard
lists automatic and manual backups. Restore:

1. Select the intended backup.
2. Read the replacement warning.
3. Type the exact `RESTORE <backup-name>` phrase.
4. Wait for the host to restart.

A rollback backup of the current ledger is created before restore begins. Do
not terminate the computer while backup or restore is active. After restart,
the Backups card records whether the last restore completed or failed.

## Tray commands

- **Open Dashboard**: host controls.
- **Open Player Join Page**: local player landing page.
- **Copy Join Link**: places the LAN URL on the clipboard.
- **Show QR Code**: opens the scannable join card.
- **Start / Stop Hosting**: closes or reopens LAN request handling without
  exiting the application.
- **Start / Show Diagnostics**: opens diagnostics and explains when the
  Diagnostic Mode shortcut is required.
- **Create Backup**: creates an owner-independent local safety copy.
- **Exit**: gracefully stops HTTP/SSE handling and closes the database.

## Uninstall

Uninstall removes the executable, shortcuts, and private firewall rule. Game
data is preserved by default. Select the explicit permanent-data checkbox and
confirm it only when you truly want to remove every game, profile, ruleset,
message, and backup.
