# Troubleshooting

## A phone cannot open the join link

1. Confirm the host and phone are connected to the same Wi-Fi or wired LAN.
2. On Windows, open **Settings → Network & internet → Properties** and confirm
   the network profile is **Private**.
3. Avoid guest Wi-Fi. Many guest networks enable client isolation, which blocks
   devices from reaching each other.
4. In **Host → Installation**, choose the adapter whose private IPv4 address
   matches the current network, save, and restart the host.
5. Disconnect VPN software temporarily if it has become the preferred adapter.
6. Verify the firewall contains one enabled “Social Games Hoster” inbound TCP
   rule with **Private** profile only.

Never solve this by exposing the application to a Public profile or by
forwarding the router port to the internet.

## The port is already occupied

Social Games Hoster writes a safe error to its application log and does not
start a second listener. Choose another port under **Installation**, approve the
Windows firewall update, then restart. To inspect the owner on Windows:

```powershell
Get-NetTCPConnection -LocalPort 8090 -State Listen |
  Select-Object LocalAddress, LocalPort, OwningProcess
```

Replace `8090` with the configured port. Do not terminate an unfamiliar system
process; use another port instead.

## Windows shows SmartScreen

Community builds without a code-signing certificate may show “Windows protected
your PC.” Download only from the project’s release page and compare the
installer’s SHA-256 hash with `SHA256SUMS.txt`. Signed builds are produced
automatically when the release environment supplies the signing certificate.

## A returning player is no longer signed in

A recovery approval intentionally invalidates all older device tokens for that
profile. Request the same display name again and ask a game master to approve
the new recovery. Disabled profiles also lose access immediately.

## Sound does not play

Browsers block autoplay. Select **Enable sound** after opening the player page.
The preference is local to that browser. Visual notifications remain available
when audio is unavailable.

## The timer changed after sleep or restart

The persisted end timestamp is authoritative. When the computer wakes or the
host restarts, expired timers reconcile to completed once; running timers resume
from the remaining wall-clock duration. Browser countdowns are visual only.

## Restore does not complete on Windows

The application creates a rollback backup before replacing data. If automatic
restart is not available, exit from the tray and launch Social Games Hoster
again. If the selected backup was damaged, restore the generated
`pre_restore_sgh_...zip` rollback backup.

## Collect a support bundle

Launch **Social Games Hoster (Diagnostic Mode)** from the Start menu, sign in as
owner, and choose **Download support bundle** under Installation. The ZIP
contains versions, safe logs, resource counters, collection counts, and network
summary. It excludes the database, auth tokens, chat bodies, profile content,
private roles, and ruleset assets.

## Find application data

The default location is:

```text
%LocalAppData%\SocialGamesHoster\data
```

Back up the whole `SocialGamesHoster` directory before manual recovery. Do not
edit the SQLite database while the tray application is running.
