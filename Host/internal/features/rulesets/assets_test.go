package rulesets

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestInspectAssetReadsImageDimensions(t *testing.T) {
	canvas := image.NewRGBA(image.Rect(0, 0, 12, 7))
	canvas.Set(0, 0, color.Black)
	var content bytes.Buffer
	if err := png.Encode(&content, canvas); err != nil {
		t.Fatal(err)
	}
	mimeType, metadata, err := inspectAsset("image", content.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if mimeType != "image/png" || metadata["width"] != 12 || metadata["height"] != 7 {
		t.Fatalf("unexpected image projection: %s %#v", mimeType, metadata)
	}
}

func TestInspectAssetRejectsAudioLongerThanMinute(t *testing.T) {
	content := testWAV(61, 8000)
	if _, _, err := inspectAsset("audio", content); err == nil {
		t.Fatal("expected long audio to be rejected")
	}
}

func testWAV(seconds int, sampleRate uint32) []byte {
	dataSize := uint32(seconds) * sampleRate
	content := make([]byte, 44+dataSize)
	copy(content[0:4], "RIFF")
	binary.LittleEndian.PutUint32(content[4:8], uint32(len(content)-8))
	copy(content[8:12], "WAVE")
	copy(content[12:16], "fmt ")
	binary.LittleEndian.PutUint32(content[16:20], 16)
	binary.LittleEndian.PutUint16(content[20:22], 1)
	binary.LittleEndian.PutUint16(content[22:24], 1)
	binary.LittleEndian.PutUint32(content[24:28], sampleRate)
	binary.LittleEndian.PutUint32(content[28:32], sampleRate)
	binary.LittleEndian.PutUint16(content[32:34], 1)
	binary.LittleEndian.PutUint16(content[34:36], 8)
	copy(content[36:40], "data")
	binary.LittleEndian.PutUint32(content[40:44], dataSize)
	return content
}
