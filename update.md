# Post-update improvements

The host now uses `github.com/gogpu/systray` v0.2.8. The v0.2.x API keeps the
existing tray behavior but adds mutable `MenuItem` handles. `Menu.Add`,
`AddCheckbox`, `AddSubmenu`, and `AddWithIcon` return a `*MenuItem`; those
handles can be updated in place with `SetLabel`, `SetChecked`, `SetDisabled`,
and `SetIcon`.

## Recommended tray improvements

The current Windows tray menu creates a static `Start / Stop Hosting` item in
`Host/internal/platform/desktop/desktop_windows.go`. Improve it in this order:

1. Keep the returned handle for the hosting item instead of discarding it.
2. Add a small rendering helper that maps the authoritative hosting state to
   the tray presentation:

   - `Start Hosting` when the host is stopped;
   - `Stop Hosting` when the host is running;
   - `SetDisabled(true)` while a start/stop operation is in progress;
   - `SetDisabled(false)` after the operation completes or fails.

3. Render the initial label and disabled state immediately after the tray menu
   is created, using `actions.IsHosting()` as the read-only state source.
4. After a successful start or stop, render the new state before showing the
   success notification. On failure, restore the previous presentation and
   keep the failure notification.
5. Keep the lifecycle decision in the existing host/controller callbacks. The
   tray package should only render state and invoke callbacks; it should not
   duplicate hosting policy or authorization rules.
6. If hosting can also change through the dashboard or another control path,
   provide a composition-root state-change notification or refresh callback so
   the tray is updated without requiring the user to open the menu. Do not
   introduce a second source of truth in the tray layer.

## Optional tray improvements

- Use `AddCheckbox` plus `SetChecked` only if a checked/unchecked hosting
  indicator is clearer than the label. Do not add a checkmark merely because
  the API supports it.
- Use `SetIcon` only if we define a clear visual distinction between running,
  stopped, and unavailable states and verify the result on Windows 10 and 11.
- Use `SetDisabled` on actions that are unavailable while hosting is stopped or
  while a backup/diagnostic operation is active, but keep the operation rules
  in the owning application callbacks.
- Reuse the same mutable-item pattern for any future tray item whose label,
  enabled state, checked state, or icon can change.

## Implementation cautions

- Do not rebuild the whole menu to reflect state; update the existing item
  handle in place.
- Update the menu only after the authoritative operation result is known.
  Optimistic labels can leave the tray lying about the actual hosting state.
- Preserve the existing serialized operation lock so rapid tray clicks cannot
  start overlapping state transitions.
- Verify callback/thread behavior for each supported Windows build. If a
  platform update must be dispatched through the tray library, use its
  supported update mechanism rather than adding an ad-hoc message loop.
- Add focused tests around the pure state-to-presentation mapping where
  possible, and perform a Windows manual check for label, disabled-state,
  notification, and Explorer-restart behavior.

## Check for reuse elsewhere

For every improvement above, check whether the same state-rendering or
mutable-control pattern can also be applied in other places: dashboard
controls, player/admin panels, connection-state indicators, backup actions,
diagnostics actions, notifications, and any future desktop integration. Keep a
single authoritative state source and apply the presentation improvement at
each relevant surface rather than fixing only the tray menu.
