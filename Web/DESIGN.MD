# Design Aesthetic Guide

The interface uses a **medieval parchment** visual language: warm paper surfaces on a dark wood-like backdrop. Core palette centers on parchment creams (`#f7e7c4`, `#fff9e6`), ink browns (`#3e322b`, `#5b4a3c`), and accent crimson (`#a62a2a`, `#8c2a3e`) for emphasis, actions, and danger states.

Single source of truth for colors and typography is `src/lib/styles/tokens.css`; components should use semantic CSS variables from that file rather than hardcoded values.

Typography reinforces theme and hierarchy: **Cinzel** for headings/buttons (uppercase, letter-spaced, ceremonial tone) and **IM Fell English** for body copy and form text (book-like, readable, period feel). Components avoid rounded modern softness; they use sharper corners, bordered frames, and layered shadows to mimic printed cards and stamped controls.

Interaction style is tactile and **snappy**: the app should feel responsive and immediate, never sluggish. Buttons lift on hover, press downward on active, and keep high-contrast focus outlines. Cards and inputs gain crimson-highlighted borders on hover/focus. Layout is centered in a constrained "parchment sheet" container with generous line height and mobile-safe padding adjustments.

Animations and transitions should be **quick and purposeful** (50-150ms for micro-interactions, max 200ms for larger transitions). Avoid slow fades or drawn-out easing that make the interface feel heavy. Prefer crisp timing functions (`ease-out` for entrances, `ease-in` for exits) over languid curves. Motion should guide attention, not delay it.
