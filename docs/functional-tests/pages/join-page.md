# Feature: Join page on a physical phone

## Required visual and usability checklist

- Inspect the initial layout, then resize the viewport and optionally rotate a phone if applicable; it was and remains visually polished and usable.
- Confirm no unintended visual issues: elements do not obscure, overlap, clip, or cut off text or controls.
- Confirm clear visual hierarchy, discoverable actions, and an efficient task path without unnecessary steps.

## Background

Given the host is running on a private Wi-Fi network
And an owner account exists
And a phone that is not already approved is connected to the same Wi-Fi network

## Scenario: Scan the displayed QR code and request entry

When the host opens the QR code from the dashboard or tray
And the phone scans the QR code with its camera
And the phone opens the resulting join page
And the user requests entry with a new profile name
Then the join page opens without a certificate, mixed-content, or unreachable-host error
And the game master sees the pending request without manually refreshing
And approving the request moves the phone to the player game page

## Scenario: Recover a profile on a replacement phone

Given an approved player profile exists
When a second physical phone requests entry using that profile name
And a game master approves the request
Then the second phone is signed in as that profile
And the first phone can no longer use its previous session to access the player page

## Scenario: Use the join page after a Wi-Fi reconnect

When the phone disconnects from and reconnects to the private Wi-Fi network
And the user returns to the join page
Then the page reconnects or displays an actionable connection state
And the player can continue without creating a duplicate profile request
