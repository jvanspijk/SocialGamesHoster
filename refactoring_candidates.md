# Refactoring candidates

## Purpose

This plan records the high-signal component-boundary and composition problems
found by reviewing the hand-written Svelte files under
`Web/src/lib/components`, their public props, application imports, route usage,
markup, styles, and the normative composition guidance in `Web/DESIGN.MD`.

The shared UI inventory should contain generic, composable components whose
names and APIs do not reveal the product domain or the route on which they are
used. Feature views may still be implemented as Svelte files, but they must live
with their owning feature or route and assemble the shared UI primitives. A file
does not become a reusable UI component merely because two routes import it.

The priorities are:

1. separate reusable UI primitives from feature views and application
   controllers;
2. remove API, realtime, global-state, and domain-model dependencies from the
   shared UI layer;
3. consolidate repeated accessibility, interaction, form, feedback, and
   surface behavior in deliberately small primitives; and
4. decompose the largest feature views without creating generic escape-hatch
   components or speculative abstractions.

Unless an issue explicitly says otherwise, preserve all routes, API payloads,
realtime topics, authorization behavior, public copy, accessible names, focus
behavior, responsive behavior, visual appearance, and reduced-motion behavior.
Refactoring should proceed in small reviewable changes. Use the smallest
relevant frontend checks for each issue and run `./scripts/Test.ps1` after the
component-boundary work has been integrated across the SPA.

`AdminChatPage.svelte` and `PlayerChatPage.svelte` are page adapters and are not
the subject of the reusable-component rule. They may remain distinct because
they normalize different data and own different navigation and authorization
behavior. Their shared visual content is covered by the issues below.

## Recommended sequence

- Implement issue 1 first so every later extraction has a clear destination and
  the shared UI inventory has an enforceable boundary.
- Implement issues 8-10 and 13-15 next. They establish the overlay, header,
  feedback, form, button, status, and state-boundary primitives needed by the
  larger decompositions.
- Implement issue 11 after the shared dialog and selectable-list foundations
  exist.
- Implement issues 2-7 and 12 as separate feature-focused changes. Do not mix
  all large component decompositions into one review.
- Implement issue 7 before deleting either timer component so countdown
  semantics and application commands remain continuously covered.
- Do not mark an issue complete when only the new primitive exists. Completion
  means all stated production call sites have migrated, obsolete duplicated
  markup and styles have been removed, and the shared component inventory has
  been reviewed again for domain or route leakage.

## Issue 1: Reserve the shared component inventory for generic UI

**Problem and evidence**

`Web/src/lib/components` currently mixes generic primitives such as `Button`,
`Field`, `Panel`, `Dialog`, and `Sheet` with complete feature views and
application controllers such as:

- `ChatApp.svelte`;
- `PendingProfileRequests.svelte`;
- `RoleReveal.svelte`;
- `GameSummaryCard.svelte`;
- `AttentionCard.svelte`; and
- the editors under `components/rulesets`.

Several of these files import API clients, realtime clients, global application
state, and domain projection types. Their names also expose product concepts in
what otherwise appears to be the reusable UI catalogue. This makes it difficult
to tell whether a component is safe to reuse and encourages feature behavior to
accumulate in the shared directory.

**Proposed change**

Define and enforce two distinct homes:

- a shared UI location containing only generic visual and interaction
  primitives; and
- feature-local locations for domain compositions, data loading, commands,
  realtime subscriptions, and route-adjacent views.

The exact directory names should follow the existing `$lib` organization, but
the boundary must be evident from imports. A shared UI component must not import
the application API client, realtime client, authentication/game stores, or
domain-specific API projections. It may import semantic tokens, generic Svelte
utilities, icons, and other shared UI primitives.

Do not disguise feature views by giving them vague generic names. Move them to
their owning feature and keep truthful feature names there. Only the shared UI
inventory must be domain- and route-neutral.

Add a short boundary rule to `Web/DESIGN.MD` or the relevant architecture
guidance once the destination is established. Avoid a new barrel file unless it
materially improves enforcement or import clarity.

**Arguments and constraints**

- Reuse is determined by responsibility and API shape, not import count.
- Feature compositions are legitimate Svelte components; they are not shared UI
  primitives.
- Do not move files mechanically before their dependencies and callers are
  understood.
- Do not force domain types through `unknown`, broad generics, or render-everything
  escape hatches merely to satisfy the directory rule.

**Verification**

- Inventory every production `.svelte` file in the shared UI location.
- Confirm none imports `$lib/api/client`, PocketBase/realtime state, feature
  stores, or domain-specific API projection types.
- Confirm shared component filenames do not disclose a feature, actor, route,
  or product workflow.
- Run frontend type checks, lint, formatting, contract tests, and the affected
  browser journeys after all moves are integrated.

## Issue 2: Decompose the chat application into UI primitives and a feature controller

**Problem and evidence**

`Web/src/lib/components/ChatApp.svelte` is approximately one thousand lines and
owns several independent responsibilities:

- room loading, sorting, filtering, and selection;
- realtime room and message subscriptions;
- local read-marker persistence;
- message pagination, sending, removal, and scrolling;
- moderation and posting-policy commands;
- the conversation rail and search UI;
- message grouping, unread and day dividers, empty/loading states; and
- the composer and responsive master-detail layout.

It accepts a `gameId`, calls `/games` and `/rooms` endpoints, reads application
authentication state, and contains product-specific labels such as `Game chat`
and `Ruleset channel`. It is a feature application rather than a reusable UI
component.

**Proposed change**

Move data loading, realtime behavior, markers, moderation, and commands into a
chat feature controller or feature-local view. Assemble its presentation from
small neutral components, extracting only established recurring structures.
Strong candidates include:

- `SplitView` for the responsive rail/detail layout;
- `ItemRail` and `SelectableList` for the searchable conversation list;
- `SearchField`;
- `MessageList` and `MessageItem`;
- `DayDivider` and `UnreadDivider`;
- `Composer`;
- `EmptyState`, `LoadingState`, and `StatusBanner`; and
- the shared form and icon-button primitives from later issues.

Keep domain normalization in the feature layer. For example, the feature should
convert room kinds into caller-owned labels and icons rather than teaching a
generic list component about teams, game masters, rulesets, or posting policy.

Do not create one equally large `ConversationView` with the same responsibilities
under a more generic name. The controller may remain cohesive where behavior
must change together, but presentation regions and repeated interaction
patterns should have narrow inputs and callbacks.

**Arguments and constraints**

- Preserve read-marker ordering, scroll restoration, first-unread placement,
  realtime deduplication, and message pagination.
- The admin and player page adapters remain separate.
- Generic message presentation may understand sender, timestamp, body, own,
  deleted, and available actions; it must not understand application room or
  participant policy.
- Avoid a generic chat SDK, event bus, or client-side repository layer.

**Verification**

- Preserve component/contract coverage for loading rooms, selecting a room,
  loading earlier messages, sending, removing, unread markers, and read-only
  states.
- Test the extracted rail, list, message item, composer, and empty states through
  semantic roles and accessible names.
- Verify the master-detail transition at phone and desktop widths.
- Run the focused chat tests and compiled-host browser chat journey.

## Issue 3: Split the visual definition editor into feature sections built from editor primitives

**Problem and evidence**

`Web/src/lib/components/rulesets/VisualDefinitionEditor.svelte` is approximately
1,395 lines. It switches between eight domain sections, creates IDs and default
domain records, mutates the complete ruleset definition, filters assets, and
contains repeated section headings, editable cards, form grids, remove actions,
select fields, checkbox groups, empty states, and nested repeaters.

`RoomPermissionEditor.svelte` and `SelectorEditor.svelte` are also tied directly
to ruleset domain types and vocabulary. Together these files make the shared UI
inventory reveal the product model and concentrate too much behavior in one
component.

**Proposed change**

Move the definition editor and its domain section components into the owning
feature. Split each independently editable section into a feature-local
component so changes to one domain area do not require editing a single
thousand-line switch.

Build those sections from neutral editor primitives such as:

- `SectionHeader`;
- `CollectionEditor`;
- `EditableCard`;
- `FormGrid`;
- `RepeaterRow`;
- `ChoiceGroup` and `CheckboxGroup`;
- `SelectField` and `CheckboxField`; and
- `EmptyState`.

Keep ID generation, default record construction, domain relationships,
inheritance semantics, and ruleset mutations in the feature layer. A shared
primitive should receive labels, values, states, snippets, and callbacks rather
than `RulesetDefinition`, `RulesetRole`, `RulesetTeam`, or related types.

Extract primitives only after comparing at least two real instances. Prefer a
few clear editor structures over a universal schema-driven form renderer.

**Arguments and constraints**

- Do not replace the current file with a configuration language or dynamic form
  engine.
- Do not hide direct Svelte binding behind untyped mutation callbacks.
- Preserve stable IDs, selection relationships, inherited policy behavior,
  asset selection, and empty-list guidance.
- Feature-local section names may remain domain-specific; shared UI names may
  not.

**Verification**

- Add focused tests for adding, removing, and editing records in every section.
- Cover nested collection editing and inherited yes/no/default states.
- Verify labels, field descriptions, keyboard access, 320px reflow, and large
  text for representative dense sections.
- Search the feature sections for repeated editable-card and form-control CSS
  after migration.

## Issue 4: Separate the pending-request workflow from its presentation

**Problem and evidence**

`PendingProfileRequests.svelte` combines API loading, realtime subscription,
count synchronization, approval/rejection commands, conflict recovery, toast
publication, two confirmation workflows, list-row rendering, status tags,
avatars, loading/error/empty states, and compact/full layouts.

The file imports application clients and `ProfileRequest`, calls admin-specific
endpoints, and owns public workflow copy. It is a feature controller and view,
not a reusable UI component.

**Proposed change**

Move the workflow into its owning feature and keep request loading, realtime
updates, decisions, conflict handling, and application copy there. Assemble the
view from neutral primitives:

- `ApprovalQueue` or a generic actionable list composition;
- `IdentityRow`;
- `StatusBadge`;
- `ActionGroup`;
- `EmptyState`, `LoadingState`, and `ErrorState`; and
- the existing dialog and field primitives after issues 8 and 13.

The feature view should decide whether a request is a recovery, what approving
means, which dialog copy to show, and which API command to invoke. Shared list
and row primitives should receive rendered supporting content and actions.

Do not create a global approval framework. Extract only the visual structures
that recur elsewhere or are independently meaningful.

**Arguments and constraints**

- Preserve realtime cleanup, count-change notifications, stale-decision
  handling, and the distinction between approval and recovery approval.
- Preserve the compact embedding behavior without teaching a generic queue
  component about the route where it appears.
- Confirmation dialogs must retain focus restoration and explicit destructive
  wording.

**Verification**

- Test initial loading, empty, error/retry, approval, recovery confirmation,
  rejection with reason, stale decisions, and realtime refresh.
- Verify action disabling and loading states during decisions.
- Run the affected approvals and game-layout frontend tests.

## Issue 5: Turn the role reveal into a feature view assembled from neutral disclosure primitives

**Problem and evidence**

`RoleReveal.svelte` consumes `PlayerGameView`, reads phase and participant
policy, calls ability activation endpoints, refreshes global game state, emits
toasts, formats domain knowledge, and owns three full-screen states: unavailable,
concealed, and revealed. It also combines media hero presentation, private
disclosure, ability cards, knowledge details, and a fixed action deck.

This is route-level feature behavior presented as a shared component.

**Proposed change**

Move the feature flow and ability commands to the owning player feature. Compose
the presentation from neutral components such as:

- `PrivacyGate` or `RevealGate`;
- `MediaHero`;
- `DetailSection`;
- `ActionCard`;
- `StatusBadge`; and
- `ActionDock`.

The feature view should supply the title, description, image, status labels,
available actions, and detail content. The disclosure primitive should own only
the interaction of hiding/revealing sensitive on-screen content and its
accessibility behavior; it must not understand roles, abilities, phases, teams,
or participants.

Keep application commands outside the visual cards. Ability actions should be
passed as callbacks with explicit busy/disabled state.

**Arguments and constraints**

- Preserve the privacy warning and hide/reveal behavior.
- Preserve the fixed mobile action placement, safe-area handling, and desktop
  navigation offset.
- Do not generalize domain knowledge formatting into the UI layer.
- Do not split every text block into a component; extract meaningful recurring
  structures.

**Verification**

- Test unavailable, concealed, revealed, no-media, media, no-abilities, active,
  finalized, and unavailable-action states.
- Test reveal, hide, activate, undo, busy, and error behavior.
- Verify private content is not rendered while concealed.
- Check phone, desktop, large-text, high-contrast, and reduced-motion behavior.

## Issue 6: Rebuild attention presentation as a feature composition

**Problem and evidence**

`AttentionCard.svelte` accepts the application `AttentionItem` projection,
understands announcement kinds, formats sender copy, resolves protected image
and audio attachments, displays queue position, and owns acknowledgement
behavior. It currently handles only one event kind and falls back to version
copy for everything else.

Although its name is generic, its public API and rendering policy are not.

**Proposed change**

Keep attention-item interpretation and acknowledgement in the owning feature.
Compose the visible item from neutral pieces such as:

- `QueuePosition`;
- `Notice` or the existing `Panel` with an appropriate reusable variant;
- `MediaAttachment`;
- `ActionGroup`; and
- the shared button primitive.

The feature view should decide which item kinds exist, sender wording, media
alternatives, acknowledgement copy, and unsupported-event behavior. The shared
surface should accept caller-owned content and actions through Svelte 5
snippets.

Before adding a new notice surface, determine whether `Panel` can express the
established frame with a semantic variant. Do not create another card solely to
rename the same border, background, padding, and action layout.

**Verification**

- Test queue position, text-only, image, audio, busy acknowledgement, and
  unsupported-item behavior at the feature level.
- Verify media alternatives and accessible announcement position remain
  available.
- Compare the final surface against `Panel` and remove duplicated surface CSS.

## Issue 7: Share one presentation-only countdown and keep timer commands outside UI

**Problem and evidence**

`TimerControl.svelte` and `TimerDisplay.svelte` independently calculate remaining
time, update the current time every 250ms, format `MM:SS`, map status to visual
state, and render a countdown. Both import `TimerProjection`; one calls timer
command endpoints and the other loads from a game-specific endpoint.

This duplicates timing behavior while coupling both visual components to the
application API and game model.

**Proposed change**

Create one presentation-only `Countdown` component with a small typed UI model:
status/tone, total or remaining time, optional end time, accessible label, and a
compact visual variant only if both existing presentations require it. Give
time calculation and formatting one owner, either inside this primitive or in a
small adjacent pure helper used by it.

Keep fetching, refreshing on visibility change, and start/pause/resume/adjust/
stop commands in feature-local controllers. Compose controls from `Countdown`,
`SelectField`, `Button`, and `ActionGroup`, passing callbacks and busy state.

Do not combine read-only display and command orchestration into a universal
timer component. They share countdown presentation, not application lifecycle.

**Arguments and constraints**

- Preserve rounding, completion behavior, status announcements, tabular digits,
  compact presentation, and 250ms visual updates.
- Define behavior for an end time crossing zero and for visibility restoration.
- Keep application status mapping exhaustive outside or at a narrow adapter
  boundary.

**Verification**

- Unit-test formatting below one minute, above one hour, zero, and completion.
- Use controlled time in component tests for running and paused states.
- Preserve command tests for start, pause, resume, adjust, restart, and clear.
- Search for duplicate `MM:SS` timer formatting after migration.

## Issue 8: Give dialogs and sheets one modal-overlay foundation

**Problem and evidence**

`Dialog.svelte` and `Sheet.svelte` duplicate the difficult parts of modal
behavior:

- storing and restoring trigger focus;
- calling `showModal` and `close`;
- moving focus to the heading;
- rendering a backdrop and elevated surface;
- providing a close control; and
- managing a scrollable body.

Their implementations already differ slightly in close/cancel handling and
focus restoration, increasing the chance that accessibility fixes apply to only
one component.

**Proposed change**

Create one internal modal-overlay foundation that owns native dialog lifecycle,
Escape/cancel behavior, focus entry, focus restoration, backdrop semantics, and
the close transition. Keep `Dialog` and `Sheet` as clear public presentation
variants if their geometry and semantics remain meaningfully different.

Use Svelte 5 snippets for header/body/actions where caller-owned structure is
needed. Keep the public APIs deliberately small. The foundation should expose a
semantic presentation variant, not arbitrary classes, dimensions, or z-index
escape hatches.

Factor the shared close control through the icon-button primitive from issue 14.

**Arguments and constraints**

- Preserve the dialog's bounded centered geometry and the sheet's full-phone/
  side-panel behavior.
- Preserve native focus trapping and do not introduce a heavy overlay library.
- Ensure every instance has an accessible name and restores focus exactly once.
- Layer values and safe-area padding remain tokenized.

**Verification**

- Extend dialog and sheet harness tests for opening, initial focus, Escape,
  explicit close, external close, and trigger-focus restoration.
- Verify phone full-screen sheet and desktop side-panel geometry.
- Test nested application state updates do not reopen or double-close an
  overlay.

## Issue 9: Consolidate repeated surface-header structure

**Problem and evidence**

`PageHeading.svelte`, `Panel.svelte`, `Dialog.svelte`, and `Sheet.svelte` each
implement a variation of title, optional description, optional actions, spacing,
and responsive alignment. Feature components also recreate local section-title
and card-heading structures.

The repetition is not identical enough to merge the enclosing components, but
the heading composition and typography must change together.

**Proposed change**

Introduce a small neutral heading composition, for example `ContentHeader` or
`SurfaceHeader`, with semantic title content, optional eyebrow/description, and
an optional actions snippet. Provide only established density/alignment
variants.

Allow callers to retain the correct heading level and enclosing semantic
element. Do not force every instance to render `h1`, and do not add an arbitrary
HTML-tag escape hatch. A title snippet or a constrained level prop is acceptable
if it remains typed and accessible.

Use the shared composition inside page headings, panels, overlay headers, and
repeated editor section headers where their structure genuinely matches. Keep
close-button behavior in the overlay rather than in the generic header.

**Arguments and constraints**

- Do not combine `PageHeading`, `Panel`, `Dialog`, and `Sheet` into one component;
  only consolidate their recurring header structure.
- Preserve responsive action stacking and existing heading hierarchy.
- Avoid a styling escape hatch that merely moves scoped CSS into props.

**Verification**

- Test title, description, eyebrow, actions, and heading semantics.
- Verify page, panel, dialog, sheet, and dense-editor examples at phone and
  desktop widths.
- Search for repeated `.section-title`, `.card-heading`, and header action-layout
  CSS and review remaining specialized instances.

## Issue 10: Establish shared feedback-state primitives

**Problem and evidence**

Loading, empty, error, informational, success, warning, and read-only states are
implemented repeatedly across `PendingProfileRequests.svelte`, `ChatApp.svelte`,
`RoleReveal.svelte`, `ErrorNotice.svelte`, and `ToastViewport.svelte`. These
instances repeat icon/text/action structure, borders, muted copy, live-region
semantics, and recovery-action layout.

`ErrorNotice.svelte` is narrowly coupled to `FormError`, while
`ToastViewport.svelte` combines viewport/timer management with the presentation
of each notification.

**Proposed change**

Create a deliberately small feedback family:

- `Alert` for persistent inline status with semantic tone and optional action;
- `EmptyState` for explanation and an optional primary action;
- `LoadingState` for literal progress/status copy; and
- `Toast` for one transient notification item.

Keep `ToastViewport` responsible for placement, timeout, pause/resume, and store
coordination, but render each item through `Toast`. Adapt `FormError` to `Alert`
outside the primitive, passing message and optional technical-detail content.

Do not create a single universal `State` component with dozens of conditional
props. Empty, loading, alert, and toast states have meaningfully different
semantics and behavior.

**Arguments and constraints**

- Preserve assertive form-error announcements and polite toast announcements.
- Success, warning, danger, and information must use icon/shape/text in addition
  to color.
- Recovery actions remain accessible buttons with explicit labels.
- Loading states must not introduce fake skeleton content unless a separate
  established need appears.

**Verification**

- Test every tone, optional action, optional details, dismissal, timeout,
  persistence, hover pause, and focus pause.
- Test empty and loading semantics with and without actions.
- Search feature styles for repeated `.empty`, `.error-state`, `.status`, and
  equivalent feedback frames after migration.

## Issue 11: Generalize the direct-message chooser into a selection dialog

**Problem and evidence**

`DirectMessageChooser.svelte` is structurally a generic list-selection dialog:
it accepts normalized IDs, primary/supporting labels, and avatar text, then
invokes a selection callback. Its public type, fixed title, class names, and
empty-state copy nevertheless bake in direct messages and players.

The file therefore exposes a feature in the shared component inventory even
though almost all of its structure is reusable.

**Proposed change**

Replace it with a neutral `SelectionDialog` or `ChoiceDialog` composed from the
shared dialog foundation and a selectable-list primitive. Accept caller-owned:

- title and description;
- normalized entries;
- accessible item labels;
- optional supporting content or item snippet;
- empty-state content; and
- selection and close callbacks.

Keep participant filtering, avatar text construction, room creation, direct-
message terminology, and navigation in the admin/player feature adapters.

Do not add a generic item-render escape hatch if a small normalized entry model
covers every current use. If richer rows are genuinely required, use a typed
Svelte snippet rather than arbitrary HTML strings or class props.

**Arguments and constraints**

- Preserve native button semantics and keyboard activation for every entry.
- Preserve stable keyed rendering and dialog focus behavior.
- Do not merge the admin and player page adapters.

**Verification**

- Test title/description labeling, empty state, item selection, supporting
  labels, keyboard activation, closing, and focus restoration.
- Preserve wrapper tests for their distinct filtering, room-opening, and
  navigation behavior.

## Issue 12: Replace the domain summary card with reusable summary structures

**Problem and evidence**

`GameSummaryCard.svelte` consumes `GameSummary` directly and combines several
independent presentation structures:

- a dark focal hero with an initial mark;
- title, metadata, duration, and participant count;
- a responsive participant record list;
- outcome status presentation; and
- achievement chips.

The name and API expose the product domain, while the implementation duplicates
surface, record-list, status, and tag patterns that can be useful elsewhere.

**Proposed change**

Move summary-model interpretation and participant-name fallback into the owning
feature. Assemble the output from neutral structures such as:

- `SummaryHero` or an established focal `Panel` variant;
- `RecordList` and `RecordItem`;
- `StatusBadge`; and
- `TagList` or `Tag`.

Prefer extending `Panel` when the hero differs only by an established dark/focal
variant. Create `SummaryHero` only if its mark-plus-copy structure recurs and is
meaningfully distinct from a panel header.

The feature should supply formatted duration, counts, outcome labels, tag text,
and icons. Shared components must not know about games, rulesets, seats,
participants, outcomes, or achievements.

**Arguments and constraints**

- Preserve compact phone reflow, tabular metadata meaning, and non-color outcome
  text.
- Avoid nested ornamental cards and retain one strong focal frame.
- Do not turn the record list into a table abstraction unless table semantics
  fit every consumer.

**Verification**

- Test zero/many tags, long names, status tones, metadata wrapping, and phone
  record reflow.
- Preserve feature-level summary tests for duration, participant names, seats,
  outcomes, and achievement labels.
- Compare surface and tag CSS with existing primitives and remove duplication.

## Issue 13: Add the missing form primitives and rebuild settings/editors from them

**Problem and evidence**

`Field.svelte` centralizes text and textarea behavior, but selects, checkboxes,
toggle-setting rows, checkbox groups, and inherited yes/no/default controls are
implemented repeatedly with raw elements and local CSS in:

- `DisplaySettings.svelte`;
- `RoomPermissionEditor.svelte`;
- `SelectorEditor.svelte`;
- `TimerControl.svelte`; and
- `VisualDefinitionEditor.svelte`.

This duplicates labels, descriptions, target sizes, borders, disabled states,
and accessible association. Some controls use 40px minimum heights despite the
44px phone-target rule.

**Proposed change**

Add narrow form primitives for the established structures:

- `SelectField` with label, help, error, required, disabled, value, and a typed
  option model or options snippet;
- `CheckboxField` for one boolean choice with supporting description;
- `CheckboxGroup` or `ChoiceGroup` for a labeled set of related choices; and
- `ToggleSetting` only if the title-plus-description setting row is materially
  distinct from `CheckboxField` in more than one use.

For inherited yes/no/default values, use `SelectField` with caller-provided
options unless tri-state semantics recur independently enough to justify a
small wrapper.

Do not expand `Field.svelte` into one component controlled by a large `kind`
switch. Text entry, selection, and checkbox interactions have different markup
and accessibility behavior and may remain separate primitives.

`DisplaySettings.svelte` may remain a feature composition outside the shared UI
inventory; it should assemble the generic setting rows while retaining ownership
of `displayPreferences`.

**Arguments and constraints**

- Each primitive owns label association, descriptions, errors, required and
  disabled treatment, focus styles, and minimum target size.
- Option values remain typed; avoid string-casting application unions throughout
  callers.
- Preserve native controls and avoid a custom select or switch implementation
  without a demonstrated need.

**Verification**

- Test labels, help, errors, required, disabled, keyboard interaction, group
  naming, and bound value updates.
- Verify every interactive control reaches the 44x44 phone target.
- Migrate the stated production call sites and search for repeated form-control
  border, label, help, and disabled CSS.
- Run frontend accessibility and type checks.

## Issue 14: Add one reusable icon-button primitive

**Problem and evidence**

Dialog close controls, sheet close controls, chat navigation and removal
controls, new-message actions, and other compact actions use raw `<button>`
elements with repeated square sizing, centering, transparent backgrounds,
focus/hover behavior, and accessible-label requirements.

`Button.svelte` is intentionally optimized for labeled actions and is not a
clear fit for icon-only controls. As a result, every feature recreates the same
interaction primitive.

**Proposed change**

Add an `IconButton` primitive with a required accessible label, icon snippet,
button type, disabled/loading state where genuinely needed, click callback, and
a small set of semantic variants such as default, ghost, and danger. It should
own target size, focus indicator, hover/pressed/disabled states, and hidden
loading text when applicable.

Use it for overlay close buttons and matching feature controls. Do not use it
for actions whose meaning is not universally understood without visible text;
those should continue using `Button` with an icon and label.

Do not add arbitrary size, color, or class props. Add a compact variant only if
it still satisfies target-size and accessibility requirements and has multiple
real consumers.

**Verification**

- Test the required accessible name, click, disabled, loading, and danger
  states.
- Verify keyboard focus and 44x44 phone targets.
- Search for raw icon-only buttons and review each remaining instance.

## Issue 15: Decouple generic-looking components from application state and transport

**Problem and evidence**

Several components have neutral filenames but hidden application dependencies:

- `ProtectedMedia.svelte` imports the application API client and rewrites the
  hard-coded `/api/app/v1` prefix before loading blobs;
- `ConnectionBadge.svelte` reads the global connection store directly;
- `DisplaySettings.svelte` reads and mutates the global display-preference
  store; and
- `ErrorNotice.svelte` accepts the application `FormError` model rather than a
  presentation contract.

These dependencies make otherwise reusable presentation difficult to test and
cause the shared UI layer to own transport or application state.

**Proposed change**

Separate each application connector from its visual primitive:

- a media presentation component receives a usable URL/blob or an explicit
  loader callback; endpoint normalization remains in the feature/transport
  adapter;
- a generic `StatusBadge` receives label, tone, and optional icon/dot, while a
  small application connector maps connection state into those props;
- display settings remain a feature composition over `CheckboxField` or
  `ToggleSetting`; and
- form-error mapping happens outside the generic `Alert`, passing message and
  optional technical-detail content.

For protected media, keep object-URL creation and revocation together in one
owner if callers still provide blobs. Do not leak authentication tokens into
ordinary media URLs or weaken protected loading semantics.

Do not ban every store import from every Svelte file. The rule applies to the
shared UI inventory; feature connectors and application shell components may
legitimately consume stores.

**Arguments and constraints**

- Preserve protected media authentication, loading, failure states, audio
  preload/autoplay behavior, and object-URL cleanup.
- Preserve connection-state wording and non-color status indication.
- Preserve device-local display preference behavior.
- Preserve trace-ID disclosure behavior for form errors.

**Verification**

- Test media loading, failure, source changes, cleanup, image alt text, and audio
  behavior with an injected loader or prepared source.
- Test status-badge tones independently from the connection store and test the
  application state mapping separately.
- Test display-setting persistence through its feature composition.
- Confirm no shared UI component imports application API clients, feature
  stores, or application error models after migration.

| Issue | Completed | Reviewed |
| :---: | :-------: | :------: |
|   1   |    [x]    |   [x]    |
|   2   |    [ ]    |   [ ]    |
|   3   |    [ ]    |   [ ]    |
|   4   |    [ ]    |   [ ]    |
|   5   |    [ ]    |   [ ]    |
|   6   |    [ ]    |   [ ]    |
|   7   |    [ ]    |   [ ]    |
|   8   |    [x]    |   [x]    |
|   9   |    [x]    |   [x]    |
|  10   |    [ ]    |   [ ]    |
|  11   |    [ ]    |   [ ]    |
|  12   |    [ ]    |   [ ]    |
|  13   |    [ ]    |   [ ]    |
|  14   |    [x]    |   [x]    |
|  15   |    [ ]    |   [ ]    |

When all issues in the table are fully completed and reviewed, bump the patch
version of the app.
