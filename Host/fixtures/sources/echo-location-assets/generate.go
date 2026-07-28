//go:build ignore

// Command generate creates the original Echo Location chart and sound assets.
//
// Run from this directory with:
//
//	go run generate.go
package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
)

const (
	width      = 960
	height     = 640
	sampleRate = 44100
)

var (
	ink        = color.RGBA{R: 207, G: 243, B: 238, A: 255}
	dimInk     = color.RGBA{R: 78, G: 138, B: 143, A: 255}
	gridInk    = color.RGBA{R: 25, G: 70, B: 78, A: 255}
	background = color.RGBA{R: 4, G: 22, B: 30, A: 255}
	accent     = color.RGBA{R: 255, G: 195, B: 92, A: 255}
	danger     = color.RGBA{R: 242, G: 105, B: 99, A: 255}
	// A fixed shuffle makes the authored cues vary across one octave while
	// keeping regeneration byte-for-byte reproducible.
	semitoneOffsets = []int{5, 0, 9, 2, 7, 11, 4, 8, 1, 6, 10, 3}
)

type hazard struct {
	Key         string
	Family      string
	Bearing     string
	Shape       string
	Count       int
	Alphabet    string
	Signal      []tone
	Alternative string
}

type tone struct {
	Pitch  string
	Length string
}

type sonarPlan struct {
	Signal        []tone
	BaseFrequency float64
	PauseAfter    int
}

var hazards = []hazard{
	{"minefield-alpha", "minefield", "port", "circle", 1, "alpha", tones("low-short", "high-long"), "Minefield chart: one round marker on the port side."},
	{"minefield-bravo", "minefield", "starboard", "triangle", 2, "beta", tones("high-short", "low-long"), "Minefield chart: a pair of pointed markers on the starboard side."},
	{"minefield-charlie", "minefield", "above", "diamond", 3, "alpha", tones("low-long", "high-short", "low-short"), "Minefield chart: a cluster of three diamond markers above the vessel."},
	{"minefield-delta", "minefield", "below", "cross", 4, "beta", tones("high-long", "low-short", "high-short"), "Minefield chart: a group of four cross markers below the vessel."},

	{"reef-maze-alpha", "reef maze", "port", "triangle", 3, "alpha", tones("low-short", "low-long", "high-short"), "Reef maze chart: three pointed markers on the port side."},
	{"reef-maze-bravo", "reef maze", "starboard", "diamond", 1, "beta", tones("high-short", "high-long"), "Reef maze chart: one diamond marker on the starboard side."},
	{"reef-maze-charlie", "reef maze", "above", "cross", 4, "alpha", tones("low-long", "low-short", "high-long"), "Reef maze chart: four cross markers above the vessel."},
	{"reef-maze-delta", "reef maze", "below", "circle", 2, "beta", tones("high-long", "high-short", "low-long"), "Reef maze chart: a pair of round markers below the vessel."},

	{"thermal-vent-alpha", "thermal vent", "port", "diamond", 2, "alpha", tones("low-short", "high-short", "high-long"), "Thermal vent chart: two diamond markers on the port side."},
	{"thermal-vent-bravo", "thermal vent", "starboard", "cross", 3, "beta", tones("high-short", "low-short", "low-long"), "Thermal vent chart: three cross markers on the starboard side."},
	{"thermal-vent-charlie", "thermal vent", "above", "circle", 1, "alpha", tones("low-long", "high-long"), "Thermal vent chart: one round marker above the vessel."},
	{"thermal-vent-delta", "thermal vent", "below", "triangle", 4, "beta", tones("high-long", "low-long", "low-short"), "Thermal vent chart: four pointed markers below the vessel."},

	{"wreck-field-alpha", "wreck field", "port", "cross", 4, "alpha", tones("low-short", "low-short", "high-long"), "Wreck field chart: four cross markers on the port side."},
	{"wreck-field-bravo", "wreck field", "starboard", "circle", 2, "beta", tones("high-short", "high-short", "low-long"), "Wreck field chart: two round markers on the starboard side."},
	{"wreck-field-charlie", "wreck field", "above", "triangle", 3, "alpha", tones("low-long", "low-long", "high-short"), "Wreck field chart: three pointed markers above the vessel."},
	{"wreck-field-delta", "wreck field", "below", "diamond", 1, "beta", tones("high-long", "high-long"), "Wreck field chart: one diamond marker below the vessel."},

	{"canyon-current-alpha", "canyon current", "port", "circle", 3, "alpha", tones("low-short", "high-long", "low-long"), "Canyon current chart: three round markers on the port side."},
	{"canyon-current-bravo", "canyon current", "starboard", "triangle", 4, "beta", tones("high-short", "low-long", "high-long"), "Canyon current chart: four pointed markers on the starboard side."},
	{"canyon-current-charlie", "canyon current", "above", "diamond", 2, "alpha", tones("low-long", "high-short", "high-short"), "Canyon current chart: two diamond markers above the vessel."},
	{"canyon-current-delta", "canyon current", "below", "cross", 1, "beta", tones("high-long", "low-short", "low-short"), "Canyon current chart: one cross marker below the vessel."},
}

var practice = hazard{
	Key:         "practice",
	Family:      "minefield",
	Bearing:     "port",
	Shape:       "triangle",
	Count:       2,
	Alphabet:    "alpha",
	Signal:      tones("low-short", "high-short"),
	Alternative: "Practice minefield chart: two pointed markers on the port side.",
}

func tones(values ...string) []tone {
	result := make([]tone, 0, len(values))
	for _, value := range values {
		switch value {
		case "low-short":
			result = append(result, tone{"low", "short"})
		case "low-long":
			result = append(result, tone{"low", "long"})
		case "high-short":
			result = append(result, tone{"high", "short"})
		case "high-long":
			result = append(result, tone{"high", "long"})
		default:
			panic("unknown tone " + value)
		}
	}
	return result
}

func main() {
	output := "generated"
	if err := os.MkdirAll(output, 0o755); err != nil {
		panic(err)
	}

	for index, item := range hazards {
		if err := writeChart(filepath.Join(output, "lookout-"+item.Key+".png"), item, index); err != nil {
			panic(err)
		}
		plan := planSonar(item, index)
		if err := writeSonar(filepath.Join(output, "sonar-"+item.Key+".wav"), plan); err != nil {
			panic(err)
		}
	}
	if err := writeChart(filepath.Join(output, "lookout-practice.png"), practice, len(hazards)); err != nil {
		panic(err)
	}
	if err := writeSonar(filepath.Join(output, "sonar-practice.wav"), sonarPlan{
		Signal:        expandSignal(practice.Signal, 5),
		BaseFrequency: 220,
		PauseAfter:    -1,
	}); err != nil {
		panic(err)
	}

	must(writeWAV(filepath.Join(output, "dive-ambience.wav"), makeDiveAmbience()))
	must(writeWAV(filepath.Join(output, "course-clear.wav"), makeSuccess()))
	must(writeWAV(filepath.Join(output, "hull-warning.wav"), makeWarning()))

	fmt.Printf("generated %d voyage pairs, 1 practice pair, and 3 game sounds in %s\n", len(hazards), output)
}

func planSonar(item hazard, index int) sonarPlan {
	count := 5
	pauseAfter := -1
	switch {
	case strings.HasSuffix(item.Key, "-bravo"):
		count = 8
	case strings.HasSuffix(item.Key, "-charlie"):
		count = 10
		pauseAfter = 5
	case strings.HasSuffix(item.Key, "-delta"):
		count = 12
		pauseAfter = 6
	}
	return sonarPlan{
		Signal:        expandSignal(item.Signal, count),
		BaseFrequency: 220 * math.Pow(2, float64(semitoneOffsets[index%len(semitoneOffsets)])/12),
		PauseAfter:    pauseAfter,
	}
}

func expandSignal(seed []tone, count int) []tone {
	result := make([]tone, count)
	for index := 0; index < count-1; index++ {
		result[index] = seed[index%len(seed)]
	}
	result[count-1] = seed[len(seed)-1]
	return result
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func writeChart(path string, item hazard, index int) error {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), &image.Uniform{background}, image.Point{}, draw.Src)

	drawGrid(img)
	drawFrame(img)
	drawFamily(img, item.Family, index)
	drawSubmarine(img)

	x, y := bearingCenter(item.Bearing)
	drawMarkers(img, x, y, item.Shape, item.Count)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	return encoder.Encode(file, img)
}

func drawGrid(img *image.RGBA) {
	for x := 80; x < width; x += 80 {
		line(img, x, 0, x, height, gridInk, 1)
	}
	for y := 80; y < height; y += 80 {
		line(img, 0, y, width, y, gridInk, 1)
	}
	for r := 80; r <= 240; r += 80 {
		circle(img, width/2, height/2, r, gridInk, 1)
	}
	line(img, width/2, 38, width/2, height-38, dimInk, 2)
	line(img, 38, height/2, width-38, height/2, dimInk, 2)
}

func drawFrame(img *image.RGBA) {
	rectOutline(img, 20, 20, width-21, height-21, ink, 3)
	rectOutline(img, 30, 30, width-31, height-31, dimInk, 1)
	for _, x := range []int{120, 240, 360, 600, 720, 840} {
		line(img, x, 20, x, 35, ink, 3)
		line(img, x, height-36, x, height-21, ink, 3)
	}
	for _, y := range []int{120, 240, 400, 520} {
		line(img, 20, y, 35, y, ink, 3)
		line(img, width-36, y, width-21, y, ink, 3)
	}
}

func drawFamily(img *image.RGBA, family string, variant int) {
	switch family {
	case "minefield":
		for i := 0; i < 11; i++ {
			angle := float64(i)*2*math.Pi/11 + float64(variant)*0.12
			radius := 180 + (i%3)*45
			x := width/2 + int(math.Cos(angle)*float64(radius))
			y := height/2 + int(math.Sin(angle)*float64(radius))
			drawMine(img, x, y, 13+(i%2)*3)
		}
	case "reef maze":
		points := [][]image.Point{
			{{50, 120}, {180, 70}, {260, 150}, {210, 250}, {70, 230}},
			{{700, 70}, {900, 100}, {850, 235}, {730, 250}, {670, 160}},
			{{80, 420}, {220, 380}, {285, 500}, {180, 585}, {55, 550}},
			{{700, 410}, {835, 375}, {915, 465}, {850, 580}, {690, 535}},
		}
		for i, polygon := range points {
			drawPolygon(img, polygon, dimInk, 4)
			if i%2 == variant%2 {
				drawPolygon(img, insetPoints(polygon, width/2, height/2, 8), gridInk, 2)
			}
		}
	case "thermal vent":
		for i := 0; i < 5; i++ {
			x := 120 + i*180
			base := 565 - (i%2)*18
			for j := 0; j < 4; j++ {
				offset := int(math.Sin(float64(j+i+variant)) * 18)
				circle(img, x+offset, base-j*42, 15+j*4, dimInk, 3)
			}
			line(img, x-32, base+20, x, base-25, dimInk, 4)
			line(img, x, base-25, x+32, base+20, dimInk, 4)
		}
		for x := 55; x < width-55; x += 22 {
			y := 575 + int(math.Sin(float64(x+variant*19)/28)*11)
			line(img, x, y, x+18, y+4, gridInk, 3)
		}
	case "wreck field":
		for i := 0; i < 7; i++ {
			x := 115 + (i%4)*245
			y := 120 + (i/4)*390 + ((i+variant)%3)*20
			drawWreck(img, x, y, (i+variant)%2 == 0)
		}
	case "canyon current":
		for side := 0; side < 2; side++ {
			lastX, lastY := 80+side*800, 50
			for step := 1; step <= 12; step++ {
				y := 50 + step*45
				wave := int(math.Sin(float64(step+variant)*0.85) * 38)
				x := 105 + wave
				if side == 1 {
					x = width - x
				}
				line(img, lastX, lastY, x, y, dimInk, 7)
				line(img, lastX+(1-side*2)*18, lastY, x+(1-side*2)*18, y, gridInk, 3)
				lastX, lastY = x, y
			}
		}
		for y := 100; y < 590; y += 70 {
			bend := int(math.Sin(float64(y+variant*13)/70) * 45)
			line(img, 345+bend, y, 590+bend, y, gridInk, 3)
			line(img, 590+bend, y, 570+bend, y-10, gridInk, 3)
			line(img, 590+bend, y, 570+bend, y+10, gridInk, 3)
		}
	}
}

func insetPoints(points []image.Point, centerX, centerY, amount int) []image.Point {
	result := make([]image.Point, len(points))
	for i, point := range points {
		dx, dy := centerX-point.X, centerY-point.Y
		length := math.Hypot(float64(dx), float64(dy))
		result[i] = image.Pt(point.X+int(float64(dx)*float64(amount)/length), point.Y+int(float64(dy)*float64(amount)/length))
	}
	return result
}

func drawMine(img *image.RGBA, x, y, radius int) {
	circle(img, x, y, radius, dimInk, 3)
	for i := 0; i < 8; i++ {
		angle := float64(i) * math.Pi / 4
		line(img, x+int(math.Cos(angle)*float64(radius)), y+int(math.Sin(angle)*float64(radius)),
			x+int(math.Cos(angle)*float64(radius+10)), y+int(math.Sin(angle)*float64(radius+10)), dimInk, 3)
	}
}

func drawWreck(img *image.RGBA, x, y int, flipped bool) {
	sign := 1
	if flipped {
		sign = -1
	}
	points := []image.Point{{x - 60, y + 18}, {x + 55, y + 18}, {x + 72, y - 6}, {x - 42, y - 6}}
	drawPolygon(img, points, dimInk, 4)
	line(img, x, y-6, x+sign*15, y-36, dimInk, 4)
	line(img, x+sign*15, y-36, x+sign*42, y-36, dimInk, 4)
	line(img, x-28, y+18, x-42, y+36, gridInk, 3)
	line(img, x+24, y+18, x+42, y+35, gridInk, 3)
}

func drawSubmarine(img *image.RGBA) {
	x, y := width/2, height/2
	ellipse(img, x, y, 56, 23, ink, 3)
	line(img, x-10, y-22, x, y-42, ink, 4)
	line(img, x, y-42, x+18, y-42, ink, 4)
	line(img, x-58, y, x-82, y-16, ink, 4)
	line(img, x-58, y, x-82, y+16, ink, 4)
	line(img, x-15, y+23, x-35, y+36, ink, 4)
	line(img, x+20, y+23, x+40, y+36, ink, 4)
	circle(img, x+25, y, 4, accent, 0)
}

func bearingCenter(bearing string) (int, int) {
	switch bearing {
	case "port":
		return 260, height / 2
	case "starboard":
		return 700, height / 2
	case "above":
		return width / 2, 150
	case "below":
		return width / 2, 500
	default:
		panic("unknown bearing")
	}
}

func drawMarkers(img *image.RGBA, centerX, centerY int, shape string, count int) {
	offsets := []image.Point{{0, 0}, {-42, 0}, {42, 0}, {0, 42}}
	if count == 2 {
		offsets = []image.Point{{-25, 0}, {25, 0}}
	} else if count == 3 {
		offsets = []image.Point{{0, -27}, {-28, 23}, {28, 23}}
	}
	for i := 0; i < count; i++ {
		drawMarker(img, centerX+offsets[i].X, centerY+offsets[i].Y, shape)
	}
	circle(img, centerX, centerY, 82, accent, 2)
}

func drawMarker(img *image.RGBA, x, y int, shape string) {
	switch shape {
	case "circle":
		circle(img, x, y, 16, danger, 5)
	case "triangle":
		drawPolygon(img, []image.Point{{x, y - 19}, {x - 18, y + 16}, {x + 18, y + 16}}, danger, 5)
	case "diamond":
		drawPolygon(img, []image.Point{{x, y - 19}, {x - 19, y}, {x, y + 19}, {x + 19, y}}, danger, 5)
	case "cross":
		line(img, x-15, y-15, x+15, y+15, danger, 6)
		line(img, x-15, y+15, x+15, y-15, danger, 6)
	default:
		panic("unknown marker")
	}
}

func line(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, thickness int) {
	if thickness < 1 {
		thickness = 1
	}
	dx := math.Abs(float64(x1 - x0))
	dy := math.Abs(float64(y1 - y0))
	steps := int(math.Max(dx, dy))
	if steps == 0 {
		setDisc(img, x0, y0, thickness/2, c)
		return
	}
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		x := int(math.Round(float64(x0) + float64(x1-x0)*t))
		y := int(math.Round(float64(y0) + float64(y1-y0)*t))
		setDisc(img, x, y, thickness/2, c)
	}
}

func setDisc(img *image.RGBA, centerX, centerY, radius int, c color.RGBA) {
	for y := centerY - radius; y <= centerY+radius; y++ {
		for x := centerX - radius; x <= centerX+radius; x++ {
			if (x-centerX)*(x-centerX)+(y-centerY)*(y-centerY) <= radius*radius && image.Pt(x, y).In(img.Bounds()) {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

func circle(img *image.RGBA, centerX, centerY, radius int, c color.RGBA, thickness int) {
	if thickness == 0 {
		for y := -radius; y <= radius; y++ {
			span := int(math.Sqrt(float64(radius*radius - y*y)))
			line(img, centerX-span, centerY+y, centerX+span, centerY+y, c, 1)
		}
		return
	}
	for degree := 0; degree < 360; degree++ {
		angle := float64(degree) * math.Pi / 180
		setDisc(img, centerX+int(math.Cos(angle)*float64(radius)), centerY+int(math.Sin(angle)*float64(radius)), thickness/2, c)
	}
}

func ellipse(img *image.RGBA, centerX, centerY, radiusX, radiusY int, c color.RGBA, thickness int) {
	for degree := 0; degree < 360; degree++ {
		angle := float64(degree) * math.Pi / 180
		setDisc(img, centerX+int(math.Cos(angle)*float64(radiusX)), centerY+int(math.Sin(angle)*float64(radiusY)), thickness/2, c)
	}
}

func rectOutline(img *image.RGBA, x0, y0, x1, y1 int, c color.RGBA, thickness int) {
	line(img, x0, y0, x1, y0, c, thickness)
	line(img, x1, y0, x1, y1, c, thickness)
	line(img, x1, y1, x0, y1, c, thickness)
	line(img, x0, y1, x0, y0, c, thickness)
}

func drawPolygon(img *image.RGBA, points []image.Point, c color.RGBA, thickness int) {
	for i := range points {
		next := (i + 1) % len(points)
		line(img, points[i].X, points[i].Y, points[next].X, points[next].Y, c, thickness)
	}
}

func writeSonar(path string, plan sonarPlan) error {
	var samples []float64
	samples = append(samples, silence(0.30)...)
	for index, item := range plan.Signal {
		frequency := plan.BaseFrequency
		if item.Pitch == "high" {
			frequency *= 1.5
		}
		duration := 0.32
		if item.Length == "long" {
			duration = 0.78
		}
		samples = append(samples, sonarPing(frequency, duration, 0.62)...)
		if index+1 == plan.PauseAfter {
			samples = append(samples, silence(1.0)...)
		} else if index < len(plan.Signal)-1 {
			samples = append(samples, silence(0.32)...)
		}
	}
	samples = append(samples, silence(0.65)...)
	return writeWAV(path, samples)
}

func sonarPing(frequency, sustain, volume float64) []float64 {
	tail := 0.42
	total := int((sustain + tail) * sampleRate)
	samples := make([]float64, total)
	for i := range samples {
		t := float64(i) / sampleRate
		attack := math.Min(1, t/0.018)
		release := 1.0
		if t > sustain {
			release = math.Exp(-8 * (t - sustain) / tail)
		}
		body := math.Sin(2*math.Pi*frequency*t) +
			0.18*math.Sin(2*math.Pi*frequency*2*t) +
			0.07*math.Sin(2*math.Pi*frequency*3*t)
		echo := 0.0
		for _, reflection := range []struct {
			delay float64
			gain  float64
		}{{0.11, 0.22}, {0.23, 0.12}, {0.34, 0.06}} {
			if t >= reflection.delay {
				echo += reflection.gain * math.Sin(2*math.Pi*frequency*(t-reflection.delay))
			}
		}
		samples[i] = (body + echo) * attack * release * volume
	}
	return samples
}

func makeDiveAmbience() []float64 {
	duration := 9.0
	total := int(duration * sampleRate)
	samples := make([]float64, total)
	seed := uint32(0x53444831)
	for i := range samples {
		t := float64(i) / sampleRate
		seed = seed*1664525 + 1013904223
		noise := (float64(seed>>8)/float64(1<<24))*2 - 1
		rumble := math.Sin(2*math.Pi*42*t)*0.13 + math.Sin(2*math.Pi*68*t)*0.06
		pulse := 0.55 + 0.45*math.Sin(2*math.Pi*0.12*t)
		samples[i] = (rumble*pulse + noise*0.025) * fade(t, duration, 0.8)
	}
	for _, at := range []float64{1.4, 4.3, 7.2} {
		add(samples, int(at*sampleRate), ping(390, 0.32, 0.20))
		add(samples, int((at+0.18)*sampleRate), ping(390, 0.45, 0.08))
	}
	return samples
}

func makeSuccess() []float64 {
	samples := silence(2.8)
	notes := []float64{330, 440, 554, 660}
	for i, frequency := range notes {
		add(samples, int((0.12+float64(i)*0.34)*sampleRate), ping(frequency, 0.28, 0.50))
	}
	add(samples, int(1.55*sampleRate), ping(880, 0.70, 0.34))
	return samples
}

func makeWarning() []float64 {
	samples := silence(3.2)
	for i := 0; i < 4; i++ {
		start := 0.16 + float64(i)*0.62
		add(samples, int(start*sampleRate), alarmSweep(0.38, 260, 150, 0.52))
	}
	return samples
}

func ping(frequency, duration, volume float64) []float64 {
	total := int(duration * sampleRate)
	samples := make([]float64, total)
	for i := range samples {
		t := float64(i) / sampleRate
		attack := math.Min(1, t/0.012)
		decay := math.Exp(-3.1 * t / duration)
		fundamental := math.Sin(2 * math.Pi * frequency * t)
		harmonic := math.Sin(2*math.Pi*frequency*2*t) * 0.12
		samples[i] = (fundamental + harmonic) * attack * decay * volume
	}
	return samples
}

func alarmSweep(duration, startFrequency, endFrequency, volume float64) []float64 {
	total := int(duration * sampleRate)
	samples := make([]float64, total)
	for i := range samples {
		t := float64(i) / sampleRate
		progress := t / duration
		frequency := startFrequency + (endFrequency-startFrequency)*progress
		envelope := math.Sin(math.Pi * progress)
		samples[i] = (math.Sin(2*math.Pi*frequency*t) + 0.2*math.Sin(2*math.Pi*frequency*2*t)) * envelope * volume
	}
	return samples
}

func silence(duration float64) []float64 {
	return make([]float64, int(duration*sampleRate))
}

func fade(t, duration, edge float64) float64 {
	return math.Min(1, math.Min(t/edge, (duration-t)/edge))
}

func add(target []float64, start int, source []float64) {
	for i, value := range source {
		if start+i >= len(target) {
			return
		}
		target[start+i] += value
	}
}

func writeWAV(path string, samples []float64) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	dataSize := uint32(len(samples) * 2)
	if _, err := file.Write([]byte("RIFF")); err != nil {
		return err
	}
	mustBinary(file, uint32(36)+dataSize)
	file.Write([]byte("WAVEfmt "))
	mustBinary(file, uint32(16))
	mustBinary(file, uint16(1))
	mustBinary(file, uint16(1))
	mustBinary(file, uint32(sampleRate))
	mustBinary(file, uint32(sampleRate*2))
	mustBinary(file, uint16(2))
	mustBinary(file, uint16(16))
	file.Write([]byte("data"))
	mustBinary(file, dataSize)
	for _, sample := range samples {
		sample = math.Max(-1, math.Min(1, sample))
		mustBinary(file, int16(sample*32767))
	}
	return nil
}

func mustBinary(file *os.File, value any) {
	if err := binary.Write(file, binary.LittleEndian, value); err != nil {
		panic(err)
	}
}
