# Echo Location host guide

Echo Location is a cooperative game for exactly three players. Assign one
Lookout, one Sonar Operator, and one Captain. The crew must clear three hazards
before its four oxygen is exhausted.

The app enforces private announcements, role-specific report channels, the
emoji-only Lookout composer, Captain-only maneuver abilities, and the Command
timer lock. The game master manually selects hazards, tracks oxygen, checks the
answer, advances phases, and awards the achievement.

## Before play

1. Assign the three roles and show them.
2. For round one, randomize the five Alpha hazards. For round two, randomize
   the ten Bravo and Charlie hazards. For round three, randomize the five Delta
   hazards. Take one from each group and avoid repeating a family when
   practical. Keep the answer column hidden from players.
3. Start with four oxygen.
4. Run the separate practice hazard. Explain its answer after resolving it:
   Minefield + port + two pointed markers starts at **Ascend**; the final
   high-short note translates to red + circle in Alphabet A, so rotate once to
   **Starboard**.
5. Remind the Lookout to use concise emoji gestures without direction arrows or
   maneuver symbols. Remind the Sonar Operator to report the raw low/high and
   short/long sequence. Chat deliberately permits corrections and additional
   messages.

## Round procedure

1. **Dive Briefing:** announce the current oxygen and start the suggested
   30-second timer. The ambience plays automatically.
2. **Observation:** send an announcement with the row's Lookout image only to
   the Lookout. Send a second announcement with its Sonar audio only to the
   Sonar Operator. Keep the saved accessibility descriptions as the image and
   audio alternatives. Start the 60-second timer.
3. **Report Window:** start the 45-second timer. The Lookout posts in Lookout
   Report and the Sonar Operator posts in Sonar Report. The Captain can read
   both channels but cannot post.
4. **Command:** start the 45-second timer. The Captain activates Ascend, Dive,
   Port, or Starboard and may undo the choice until the timer ends.
5. **Resolution:** compare the locked maneuver with the answer below. Announce
   the full chart, signal transcript, and answer to everyone. Play Course Clear
   after a correct maneuver. After an incorrect maneuver, reduce oxygen by one
   and play Hull Warning.
6. Begin the next round, or enter Review after three correct maneuvers or when
   oxygen reaches zero. Award Safe Passage manually if the crew completes the
   voyage.

## Captain manual summary

Use the full manual on the Captain's role card:

1. Find the family-and-bearing starting maneuver.
2. Apply that family's marker condition, rotating through Ascend, Starboard,
   Dive, Port.
3. Translate the final sonar note with Alphabet A for port/above or Alphabet B
   for starboard/below. Rotate once for every translated property that is red,
   square, or cross.

## Hazard answer key

| Hazard | Lookout image | Sonar audio | Answer |
| --- | --- | --- | --- |
| Minefield Alpha | `lookout-minefield-alpha` | `sonar-minefield-alpha` | Dive |
| Minefield Bravo | `lookout-minefield-bravo` | `sonar-minefield-bravo` | Dive |
| Minefield Charlie | `lookout-minefield-charlie` | `sonar-minefield-charlie` | Port |
| Minefield Delta | `lookout-minefield-delta` | `sonar-minefield-delta` | Starboard |
| Reef Maze Alpha | `lookout-reef-maze-alpha` | `sonar-reef-maze-alpha` | Starboard |
| Reef Maze Bravo | `lookout-reef-maze-bravo` | `sonar-reef-maze-bravo` | Port |
| Reef Maze Charlie | `lookout-reef-maze-charlie` | `sonar-reef-maze-charlie` | Dive |
| Reef Maze Delta | `lookout-reef-maze-delta` | `sonar-reef-maze-delta` | Ascend |
| Thermal Vent Alpha | `lookout-thermal-vent-alpha` | `sonar-thermal-vent-alpha` | Dive |
| Thermal Vent Bravo | `lookout-thermal-vent-bravo` | `sonar-thermal-vent-bravo` | Dive |
| Thermal Vent Charlie | `lookout-thermal-vent-charlie` | `sonar-thermal-vent-charlie` | Starboard |
| Thermal Vent Delta | `lookout-thermal-vent-delta` | `sonar-thermal-vent-delta` | Ascend |
| Wreck Field Alpha | `lookout-wreck-field-alpha` | `sonar-wreck-field-alpha` | Port |
| Wreck Field Bravo | `lookout-wreck-field-bravo` | `sonar-wreck-field-bravo` | Ascend |
| Wreck Field Charlie | `lookout-wreck-field-charlie` | `sonar-wreck-field-charlie` | Starboard |
| Wreck Field Delta | `lookout-wreck-field-delta` | `sonar-wreck-field-delta` | Ascend |
| Canyon Current Alpha | `lookout-canyon-current-alpha` | `sonar-canyon-current-alpha` | Port |
| Canyon Current Bravo | `lookout-canyon-current-bravo` | `sonar-canyon-current-bravo` | Ascend |
| Canyon Current Charlie | `lookout-canyon-current-charlie` | `sonar-canyon-current-charlie` | Starboard |
| Canyon Current Delta | `lookout-canyon-current-delta` | `sonar-canyon-current-delta` | Port |

The answer distribution is balanced: five cards resolve to each maneuver.

## Host safeguards

- Never send a chart or sonar cue to the Captain.
- Use the player audience for private clues, not the crew team.
- Do not include the answer in an accessibility alternative.
- Confirm that the Command timer has finalized before revealing the answer.
- Do not reuse a hazard within one voyage.
- If browser audio is unavailable, use the saved raw-note alternative rather
  than describing its translated properties.
