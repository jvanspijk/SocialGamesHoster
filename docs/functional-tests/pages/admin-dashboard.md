# Feature: Game-master dashboard in a real LAN session

## Required visual and usability checklist

- Inspect the initial layout, then resize the viewport and optionally rotate a phone if applicable; it was and remains visually polished and usable.
- Confirm no unintended visual issues: elements do not obscure, overlap, clip, or cut off text or controls.
- Confirm clear visual hierarchy, discoverable actions, and an efficient task path without unnecessary steps.

## Background

Given a game master is signed in on the Windows host computer
And at least two physical phones are approved and joined to an open lobby
And a valid saved ruleset with a phase and an optional sound cue is available

## Scenario: Coordinate a live game across real devices

When the game master starts the game
And the game master selects a phase and starts its timer
Then each phone shows its own current game state without a manual refresh
And the timer remains visually in step with the game-master dashboard
And each player sees only their own private role and knowledge

## Scenario: Approve a late player without leaving the live game

Given the game master is viewing the live-game dashboard
When another phone requests entry with a new profile name
Then the live-game entry-request control indicates that attention is needed
When the game master opens entry requests and approves the profile
Then the phone continues to the player game page without a manual refresh
And the game master remains on the same live-game page
And the entry-request attention state clears

## Scenario: Deliver a targeted sound announcement after phone interaction

Given one player has selected Enable sound in their phone browser
And another player has not enabled sound
When the game master sends an announcement with a sound cue targeted to the enabled player
Then the enabled player's phone plays the cue after the user gesture that enabled sound
And the other player sees the announcement but does not hear the cue

## Scenario: Deliver one-off announcement media privately

Given one player is the only selected recipient
When the game master uploads an image and audio file in the announcement composer
And provides an image description and audio alternative
Then the selected player's phone displays and plays the attachments
And another signed-in player's phone cannot open either attachment URL
And the uploaded files do not appear in the ruleset Media section or a later announcement

## Scenario: Restore a backup from the owner page

Given the owner has created a manual backup before making a recognizable game change
When the owner restores that backup and enters the required confirmation text
Then the host restarts as indicated by the UI
And the dashboard shows the recorded restore outcome after it becomes available
And the recognizable change made after the backup is absent
