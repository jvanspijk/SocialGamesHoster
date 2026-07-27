# Feature: Player game page on physical mobile browsers

## Required visual and usability checklist

- Inspect the initial layout, then resize the viewport and optionally rotate a phone if applicable; it was and remains visually polished and usable.
- Confirm no unintended visual issues: elements do not obscure, overlap, clip, or cut off text or controls.
- Confirm clear visual hierarchy, discoverable actions, and an efficient task path without unnecessary steps.

## Background

Given a running game has at least two participants
And one participant is signed in on an iPhone Safari browser
And another participant is signed in on an Android Chrome browser

## Scenario: Use the player page in portrait, landscape, and with the keyboard open

When each player views the game page in portrait orientation
And rotates the phone to landscape orientation
And opens the on-screen keyboard to write a room message
Then the current role, timer, and active room remain usable
And no essential control is obscured by the browser chrome, safe area, or keyboard
And the page can be operated by touch without accidental adjacent actions

## Scenario: Receive a real-time room message while the page is backgrounded

When player A backgrounds their browser briefly
And player B sends a message to a room shared with player A
And player A returns to the browser
Then player A can identify the unread message and open the correct room
And the message is present once, with the sender presentation configured for that room

## Scenario: Verify reduced-motion preference on a physical browser

Given the device has Reduce Motion or an equivalent browser preference enabled
When the player opens the game page and changes between its tabs
Then essential state changes remain understandable
And distracting or prolonged motion is reduced
