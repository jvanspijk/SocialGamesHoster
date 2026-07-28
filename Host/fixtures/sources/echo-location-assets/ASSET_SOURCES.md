# Echo Location asset sources

All gameplay clues and sounds listed here are original, deterministic,
procedurally generated works created for this repository by `generate.go`. No
third-party images, fonts, recordings, samples, or AI-generated media are used
in those procedural assets.

Run `go run generate.go` from this directory to reproduce the lossless PNG and
PCM WAV intermediates. The checked-in delivery assets are WebP and Ogg Vorbis:
images were converted with Sharp's libwebp encoder, and audio was converted
with the local Kdenlive FFmpeg build's libvorbis encoder. Conversion changes
only the container/encoding, not the authored content.

## Separately generated game images

These two illustrations were created with OpenAI's built-in image-generation
tool on 2026-07-28. They used only the original text prompts recorded below and
no third-party input images. They are not outputs of `generate.go`.

OpenAI's Terms of Use state that, as between the user and OpenAI and to the
extent permitted by applicable law, the user owns the output:
https://openai.com/policies/terms-of-use/

| Asset key | Source |
| --- | --- |
| `echo-location-cover` | OpenAI image generation, original text prompt only; resized and center-cropped to 1280×720 WebP. |
| `safe-passage-badge` | OpenAI image generation, original text prompt only; resized to 512×512 WebP. |

### Cover prompt

> Original cinematic illustration of a small submarine descending through dark
> ocean water toward a field of underwater hazards, with three subtle
> observation motifs suggesting visual lookout, sonar, and command teamwork.
> Richly textured illustrated game cover, dark maritime adventure, polished but
> not photorealistic. Landscape composition with a strong submarine silhouette,
> deep navy water, teal bioluminescence, and warm amber windows. No copyrighted
> characters or logos, typography, text, watermark, or UI.

### Achievement prompt

> Original circular brass-and-enamel badge showing a small submarine passing
> safely through a narrow underwater canyon, symbolizing teamwork and safe
> passage. Crisp readable silhouette, nautical instrument aesthetic, centered
> square composition, deep navy enamel, teal accents, and antique brass. No
> words, letters, numbers, existing military insignia, copyrighted logos, or
> watermark.

## Visual clues

The 20 voyage-deck `lookout-{family}-{variant}.webp` files are 960×640 nautical
instrument charts. They use only programmatically drawn lines and geometric
primitives. Every chart exposes four independent, ordinary visual features:
hazard family, relative bearing, marker shape, and marker count/grouping.

`lookout-practice.webp` is a reproducible onboarding chart outside the 20-card
voyage deck. It depicts a minefield with two pointed markers on the port side.

| Asset key | Family | Bearing | Marker | Grouping | Accessible alternative |
| --- | --- | --- | --- | ---: | --- |
| `lookout-minefield-alpha` | Minefield | Port | Circle | 1 | Minefield chart: one round marker on the port side. |
| `lookout-minefield-bravo` | Minefield | Starboard | Triangle | 2 | Minefield chart: a pair of pointed markers on the starboard side. |
| `lookout-minefield-charlie` | Minefield | Above | Diamond | 3 | Minefield chart: a cluster of three diamond markers above the vessel. |
| `lookout-minefield-delta` | Minefield | Below | Cross | 4 | Minefield chart: a group of four cross markers below the vessel. |
| `lookout-reef-maze-alpha` | Reef maze | Port | Triangle | 3 | Reef maze chart: three pointed markers on the port side. |
| `lookout-reef-maze-bravo` | Reef maze | Starboard | Diamond | 1 | Reef maze chart: one diamond marker on the starboard side. |
| `lookout-reef-maze-charlie` | Reef maze | Above | Cross | 4 | Reef maze chart: four cross markers above the vessel. |
| `lookout-reef-maze-delta` | Reef maze | Below | Circle | 2 | Reef maze chart: a pair of round markers below the vessel. |
| `lookout-thermal-vent-alpha` | Thermal vent | Port | Diamond | 2 | Thermal vent chart: two diamond markers on the port side. |
| `lookout-thermal-vent-bravo` | Thermal vent | Starboard | Cross | 3 | Thermal vent chart: three cross markers on the starboard side. |
| `lookout-thermal-vent-charlie` | Thermal vent | Above | Circle | 1 | Thermal vent chart: one round marker above the vessel. |
| `lookout-thermal-vent-delta` | Thermal vent | Below | Triangle | 4 | Thermal vent chart: four pointed markers below the vessel. |
| `lookout-wreck-field-alpha` | Wreck field | Port | Cross | 4 | Wreck field chart: four cross markers on the port side. |
| `lookout-wreck-field-bravo` | Wreck field | Starboard | Circle | 2 | Wreck field chart: two round markers on the starboard side. |
| `lookout-wreck-field-charlie` | Wreck field | Above | Triangle | 3 | Wreck field chart: three pointed markers above the vessel. |
| `lookout-wreck-field-delta` | Wreck field | Below | Diamond | 1 | Wreck field chart: one diamond marker below the vessel. |
| `lookout-canyon-current-alpha` | Canyon current | Port | Circle | 3 | Canyon current chart: three round markers on the port side. |
| `lookout-canyon-current-bravo` | Canyon current | Starboard | Triangle | 4 | Canyon current chart: four pointed markers on the starboard side. |
| `lookout-canyon-current-charlie` | Canyon current | Above | Diamond | 2 | Canyon current chart: two diamond markers above the vessel. |
| `lookout-canyon-current-delta` | Canyon current | Below | Cross | 1 | Canyon current chart: one cross marker below the vessel. |

## Sonar clues

The 20 voyage-deck `sonar-{family}-{variant}.ogg` files are original synthesized
tones delivered as Ogg Vorbis. No recorded samples are used. Each cue uses its
own deterministic low frequency within the octave from 220 Hz through 415.30
Hz; its high frequency is always a just 3:2 musical fifth above that base.
Short notes sustain for 320 ms and long notes for 780 ms, with a reverberant
420 ms tail, 320 ms ordinary spacing, and 650 ms of final silence.

Alpha, Bravo, Charlie, and Delta contain 5, 8, 10, and 12 beeps respectively.
Every Charlie pauses for exactly 1.0 second after beep 5; every Delta pauses for
exactly 1.0 second after beep 6. `sonar-practice.ogg` contains five beeps with no
internal one-second pause: low short, high short, low short, high short, high
short, using 220 Hz low and 330 Hz high.

| Asset key | Low/high frequency | Exact accessible transcript |
| --- | --- | --- |
| `sonar-minefield-alpha` | 293.66 / 440.50 Hz | Low short, high long, low short, high long, high long. |
| `sonar-minefield-bravo` | 220.00 / 330.00 Hz | High short, low long, high short, low long, high short, low long, high short, low long. |
| `sonar-minefield-charlie` | 369.99 / 554.99 Hz | Low long, high short, low short, low long, high short; 1.0-second pause; low short, low long, high short, low short, low short. |
| `sonar-minefield-delta` | 246.94 / 370.41 Hz | High long, low short, high short, high long, low short, high short; 1.0-second pause; high long, low short, high short, high long, low short, high short. |
| `sonar-reef-maze-alpha` | 329.63 / 494.44 Hz | Low short, low long, high short, low short, high short. |
| `sonar-reef-maze-bravo` | 415.30 / 622.96 Hz | High short, high long, high short, high long, high short, high long, high short, high long. |
| `sonar-reef-maze-charlie` | 277.18 / 415.77 Hz | Low long, low short, high long, low long, low short; 1.0-second pause; high long, low long, low short, high long, high long. |
| `sonar-reef-maze-delta` | 349.23 / 523.84 Hz | High long, high short, low long, high long, high short, low long; 1.0-second pause; high long, high short, low long, high long, high short, low long. |
| `sonar-thermal-vent-alpha` | 233.08 / 349.62 Hz | Low short, high short, high long, low short, high long. |
| `sonar-thermal-vent-bravo` | 311.13 / 466.69 Hz | High short, low short, low long, high short, low short, low long, high short, low long. |
| `sonar-thermal-vent-charlie` | 392.00 / 587.99 Hz | Low long, high long, low long, high long, low long; 1.0-second pause; high long, low long, high long, low long, high long. |
| `sonar-thermal-vent-delta` | 261.63 / 392.44 Hz | High long, low long, low short, high long, low long, low short; 1.0-second pause; high long, low long, low short, high long, low long, low short. |
| `sonar-wreck-field-alpha` | 293.66 / 440.50 Hz | Low short, low short, high long, low short, high long. |
| `sonar-wreck-field-bravo` | 220.00 / 330.00 Hz | High short, high short, low long, high short, high short, low long, high short, low long. |
| `sonar-wreck-field-charlie` | 369.99 / 554.99 Hz | Low long, low long, high short, low long, low long; 1.0-second pause; high short, low long, low long, high short, high short. |
| `sonar-wreck-field-delta` | 246.94 / 370.41 Hz | High long, high long, high long, high long, high long, high long; 1.0-second pause; high long, high long, high long, high long, high long, high long. |
| `sonar-canyon-current-alpha` | 329.63 / 494.44 Hz | Low short, high long, low long, low short, low long. |
| `sonar-canyon-current-bravo` | 415.30 / 622.96 Hz | High short, low long, high long, high short, low long, high long, high short, high long. |
| `sonar-canyon-current-charlie` | 277.18 / 415.77 Hz | Low long, high short, high short, low long, high short; 1.0-second pause; high short, low long, high short, high short, high short. |
| `sonar-canyon-current-delta` | 349.23 / 523.84 Hz | High long, low short, low short, high long, low short, low short; 1.0-second pause; high long, low short, low short, high long, low short, low short. |

## Game sounds

| Asset key | Source and accessible alternative |
| --- | --- |
| `dive-ambience` | Original synthesized low mechanical rumble with three subdued sonar pings. |
| `course-clear` | Original synthesized four-note rising confirmation followed by one sustained high ping. |
| `hull-warning` | Original synthesized sequence of four descending alarm sweeps. |
