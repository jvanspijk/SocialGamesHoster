# Mobile-first design standard

This guide is normative for every route. The release sign-off matrix lives in
`docs/VISUAL_OVERHAUL_PLAN.md`.

## Voice and visual language

The interface is rich medieval parchment: warm paper, dark ink, crimson actions,
sharp frames, seals, and small illuminated details over a dark wood surround.
Decoration concentrates on focal content and important transitions. Routine
forms and dense host tools use quiet surfaces with minimal nesting.

System-owned interface text must be concise, literal, action-oriented, and
necessary. The medieval identity is visual, not verbal.

- Do not use fantasy metaphors, roleplay narration, ceremonial wording, or
  invented fantasy terms for ordinary actions.
- Prefer `Start game`, `Show roles`, `Pause timer`, `Players can post`, and
  `Save ruleset`.
- Status text explains the current state and, when useful, the next action.
- Remove decorative copy that communicates nothing.
- Product terms such as game master, role, phase, ruleset, achievement, and
  game remain valid.
- Ruleset-authored role names, descriptions, abilities, and instructions are
  displayed as authored.

Cinzel is limited to short headings, navigation labels, and important actions.
IM Fell English is the readable body face for forms, instructions, statuses, and
dense content. Components use semantic variables from `tokens.css`; hardcoded
colors are reserved for media or a documented one-off contrast correction.

## Layout and responsive hierarchy

The product is designed from 320 CSS px upward.

- `--phone`: below 48rem; one task per viewport, full-screen sheets, bottom
  navigation, no horizontal rails.
- `--tablet`: 48rem–63.99rem; wider sheets and selective master-detail.
- `--desktop`: 64rem and above; side rail, side panels, and stable multi-column
  workspaces.
- Content may wrap and vertically scroll except for the running player
  current-moment stage, which fits in `100dvh`.
- Use `minmax(0, 1fr)`, `min(100%, …)`, and wrapping controls. Never rely on
  shrinking a desktop canvas or page-level horizontal scrolling.
- Shell padding includes `env(safe-area-inset-*)`. Bottom navigation, sheets, and
  composers add the relevant safe-area inset.
- Focused sheet content and composers remain visible when a software keyboard
  reduces the visual viewport.

## Semantic tokens

`tokens.css` defines:

- paper/background/ink and interactive accent colors;
- success, warning, danger, information, disabled, and focus treatments;
- display/body fonts and a manual large-text scale;
- a four-point spacing rhythm (`--space-1` through `--space-8`);
- density, border, corner, shadow, safe-area, target-size, layer, breakpoint, and
  motion values.

Default text and controls meet WCAG 2.2 AA. `data-contrast="high"` targets 7:1
text contrast, removes texture behind text, and strengthens outlines.
`data-text-size="large"` raises the base size while preserving reflow. These
device-local settings are chosen in Display settings and are not inferred from
OS contrast or text preferences.

## Components and states

Every interactive control has at least a 44×44 CSS px target on phones and a
visible focus indicator.

### Composition and reuse

Build the interface from generic, composable primitives and small
subcomponents. Prefer one well-defined component with configurable props and
slot/snippet regions over multiple feature-named components that repeat the
same markup, behavior, or styles.

- Put recurring control structure in form primitives. For example, one text
  input primitive owns the label, input, help text, error state, required and
  disabled treatment, while callers provide the id, label, type, value, and
  messages.
- Put recurring layout and surface structure in slot-based primitives. For
  example, a card or panel exposes semantic header, body, and action regions;
  callers supply feature-specific content rather than recreate the frame.
- Extract meaningful subcomponents for repeated parts of a larger pattern, such
  as a dialog header, sheet footer, field error, card actions, or empty-state
  action.
- Keep domain components focused on composition and domain behavior. They
  should assemble primitives instead of copying their HTML and CSS.
- `src/lib/components` is reserved for domain- and route-neutral visual and
  interaction primitives. Components that load data, issue commands, subscribe
  to realtime updates, read application state, or use domain API projections
  live under their owning `src/lib/features/<feature>/components` location.
- Use Svelte 5 snippets/render props for new slot-like APIs. Existing
  `<slot>`-based components may remain until they are otherwise changed.
- Keep primitive APIs semantic and deliberately small. Use props for stable
  variants and state; use snippets for caller-owned structure. Do not add a
  generic escape-hatch prop merely to avoid defining a clear variant.
- Keep accessibility behavior inside the primitive that owns the interaction,
  including label association, descriptions, error announcements, keyboard
  behavior, focus management, and ARIA state.
- Centralize shared styling in primitives and semantic tokens. Feature
  components may add layout specific to the feature, but must not duplicate a
  primitive's borders, spacing, typography, or interaction-state CSS.

Before adding a component, search the existing primitives and subcomponents.
Extend an existing primitive when the difference is a reusable variant. Create
a new primitive when the structure or behavior is meaningfully different.
Avoid speculative abstractions used only once, but consolidate the second
instance of materially repeated markup or CSS unless doing so would make the API
less clear.

- Loading: short literal status, optional progress indicator, no fake content.
- Empty: explain the state and offer one relevant primary action.
- Error: state what failed and provide an actionable recovery when possible.
- Disabled: preserve readable contrast and explain unavailable actions nearby.
- Selected: use shape/border and text in addition to color.
- Destructive: danger styling, explicit object/action wording, and confirmation
  for irreversible changes.
- Success/warning: icon plus text; never color alone.
- Reconnecting/offline: compact persistent status with no data-loss promise.

Use shared status bars, bottom navigation, sheets, dialogs, sticky control decks,
attention cards, channel views, audience pickers, section menus, and display
settings. A mobile sheet fills the viewport and becomes a bounded side panel on
desktop. Sheets/dialogs:

1. have an accessible name;
2. trap focus;
3. close on Escape and explicit Back/Close;
4. restore focus to their trigger;
5. keep the focused element unobscured;
6. do not move focus when background attention arrives.

## Motion and layer order

Micro-interactions last 50–150ms; larger sheet transitions are no longer than
200ms. Entrances use `ease-out`, exits `ease-in`. Motion guides attention and
never delays input. `prefers-reduced-motion` reduces all nonessential movement to
effectively zero.

Layer order is tokenized: base content, sticky controls, navigation/status,
backdrop, sheet, dialog, toast. No component invents a larger arbitrary z-index.

## Content density and decoration

Use a single strong frame around focal content. Avoid nested cards, repeated
ornamental dividers, and shadows on every row. Tables reflow into labeled records
on phones. Forms are one column on phones; related short fields may share a row
only when they still fit at 320 px and 200% zoom.

Icons reinforce a text label unless the action is universally understood and has
an accessible name. Badges use a dot or `New`, never numeric unread counts.
