# Feature: Ruleset editor media preview on a physical phone

## Required visual and usability checklist

- Inspect the initial layout, then resize the viewport and optionally rotate a phone if applicable; it was and remains visually polished and usable.
- Confirm no unintended visual issues: elements do not obscure, overlap, clip, or cut off text or controls.
- Confirm clear visual hierarchy, discoverable actions, and an efficient task path without unnecessary steps.

## Background

Given a game master has a working ruleset open in the ruleset editor
And a representative image and audio cue are available for the game
And a physical iPhone or Android phone is available on the same private LAN

## Scenario: Use uploaded media in a real player browser

When the game master adds the image and audio cue to the working ruleset
And previews the role card, phase flow, and media
And saves the ruleset as Valid
And creates or opens a game using the saved ruleset
And a player opens the game on the physical phone
Then the image displays at a usable size without breaking the player layout
And the audio cue can play after the player explicitly enables sound
And the player can still read and operate the page while the media is present

## Scenario: Verify the saved media on a second mobile browser

When the same game is opened on a second physical mobile browser
Then the image and cue are available without relying on the game master's browser cache
And the controls remain usable on the second browser
