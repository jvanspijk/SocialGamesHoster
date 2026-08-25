package rulesets

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"path/filepath"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"golang.org/x/image/webp"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const (
	maxImageSize = 2 << 20
	maxAudioSize = 5 << 20
)

// MediaUploadLimit and InspectMediaUpload keep every feature that accepts
// ruleset-compatible media on the same byte, signature, dimension, and
// duration policy.
func MediaUploadLimit(kind string) int64 {
	if kind == "image" {
		return maxImageSize
	}
	return maxAudioSize
}

func InspectMediaUpload(kind string, content []byte) (string, error) {
	mimeType, _, err := inspectAsset(kind, content)
	return mimeType, err
}

func registerAssetPreviewRoute(event *core.ServeEvent) {
	event.Router.GET("/api/app/v1/ruleset-assets/{id}", previewAsset)
}
func previewAsset(event *core.RequestEvent) error {
	if event.Auth == nil || !event.Auth.GetBool("active") {
		return httpx.WriteError(event, result.AppError{Code: "auth.required", Message: "Sign in to view this ruleset asset.", Status: http.StatusUnauthorized})
	}
	asset, err := event.App.FindRecordById("ruleset_assets", event.Request.PathValue("id"))
	if err != nil || asset.GetString("storage_state") != "ready" {
		return httpx.WriteError(event, result.AppError{Code: "asset.not_found", Message: "Ruleset asset not found.", Status: http.StatusNotFound})
	}
	switch event.Auth.Collection().Name {
	case actorauth.GameMastersCollection:
	case actorauth.PlayerProfilesCollection:
		if !playerMayReadVersionAsset(event.App, event.Auth.Id, asset.GetString("ruleset_version"), asset.GetString("asset_key")) {
			return httpx.WriteError(event, result.Forbidden("asset.forbidden", "This asset is not part of your current game."))
		}
	default:
		return httpx.WriteError(event, result.Forbidden("asset.forbidden", "This asset is not available to this account."))
	}
	fsys, err := event.App.NewFilesystem()
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer fsys.Close()
	reader, err := fsys.GetReader(asset.BaseFilesPath() + "/" + asset.GetString("file"))
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	defer reader.Close()
	content, err := io.ReadAll(io.LimitReader(reader, maxAudioSize+1))
	if err != nil || len(content) > maxAudioSize {
		return httpx.WriteError(event, result.Internal(errors.New("asset storage content exceeds limit")))
	}
	event.Response.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	event.Response.Header().Set("ETag", `"`+asset.GetString("checksum")+`"`)
	event.Response.Header().Set("Content-Disposition", `inline; filename="`+filepath.Base(asset.GetString("file"))+`"`)
	return event.Blob(http.StatusOK, asset.GetString("mime_type"), content)
}

func playerMayReadVersionAsset(app core.App, profileID, versionID, assetKey string) bool {
	games, err := app.FindRecordsByFilter(
		"games",
		"ruleset_version = {:version} && status != 'draft'",
		"",
		100,
		0,
		dbx.Params{"version": versionID},
	)
	if err != nil {
		return false
	}
	for _, game := range games {
		participant, findErr := gamepolicyapp.CurrentParticipantByGameAndProfile(app, game.Id, profileID)
		if findErr == nil {
			attachments, attachmentErr := app.FindRecordsByFilter(
				"attention_items",
				"game = {:game} && (image_asset_key = {:key} || audio_asset_key = {:key})",
				"",
				100,
				0,
				dbx.Params{"game": game.Id, "key": assetKey},
			)
			if attachmentErr != nil {
				continue
			}
			if len(attachments) > 0 {
				mayReadAttachment := false
				for _, attachment := range attachments {
					receipts, receiptErr := app.FindRecordsByFilter(
						"attention_receipts",
						"attention_item = {:item} && participant = {:participant}",
						"",
						1,
						0,
						dbx.Params{"item": attachment.Id, "participant": participant.Id},
					)
					if receiptErr == nil && len(receipts) == 1 {
						mayReadAttachment = true
						break
					}
				}
				if !mayReadAttachment {
					continue
				}
			}
			if !game.GetBool("roles_visible") && isPrivateRoleAsset(game, assetKey) {
				continue
			}
			return true
		}
	}
	return false
}

func isPrivateRoleAsset(game *core.Record, assetKey string) bool {
	definition, err := DecodeSnapshot(game.Get("ruleset_snapshot"))
	if err != nil {
		return true
	}
	for _, role := range definition.Roles {
		if role.ImageAssetKey == assetKey {
			return true
		}
	}
	for _, team := range definition.Teams {
		if team.ImageAssetKey == assetKey {
			return true
		}
	}
	for _, ability := range definition.Abilities {
		if ability.ImageAssetKey == assetKey {
			return true
		}
	}
	return false
}

func inspectAsset(kind string, content []byte) (string, map[string]any, error) {
	detected := http.DetectContentType(content)
	if kind == "image" {
		var configuration image.Config
		var err error
		if bytes.HasPrefix(content, []byte("RIFF")) && len(content) >= 12 && string(content[8:12]) == "WEBP" {
			configuration, err = webp.DecodeConfig(bytes.NewReader(content))
			detected = "image/webp"
		} else {
			configuration, _, err = image.DecodeConfig(bytes.NewReader(content))
		}
		if err != nil || (detected != "image/jpeg" && detected != "image/png" && detected != "image/webp") {
			return "", nil, errors.New("the image signature is not a supported JPEG, PNG, or WebP file")
		}
		if configuration.Width < 1 || configuration.Height < 1 || configuration.Width > 4096 || configuration.Height > 4096 {
			return "", nil, errors.New("images must be no larger than 4096 by 4096 pixels")
		}
		return detected, map[string]any{"width": configuration.Width, "height": configuration.Height}, nil
	}
	duration, mimeType, err := audioDuration(content, detected)
	if err != nil {
		return "", nil, err
	}
	if duration <= 0 || duration > 60_000 {
		return "", nil, errors.New("audio must be no longer than 60 seconds")
	}
	return mimeType, map[string]any{"durationMs": duration}, nil
}

func audioDuration(content []byte, detected string) (int64, string, error) {
	if len(content) >= 12 && string(content[:4]) == "RIFF" && string(content[8:12]) == "WAVE" {
		duration, err := wavDuration(content)
		return duration, "audio/wav", err
	}
	if bytes.HasPrefix(content, []byte("OggS")) {
		duration, err := oggDuration(content)
		return duration, "audio/ogg", err
	}
	if len(content) >= 12 && string(content[4:8]) == "ftyp" {
		duration, err := mp4Duration(content)
		return duration, "audio/mp4", err
	}
	if detected == "audio/mpeg" || bytes.HasPrefix(content, []byte("ID3")) || looksLikeMP3(content) {
		duration, err := mp3Duration(content)
		return duration, "audio/mpeg", err
	}
	return 0, "", errors.New("the audio signature is not a supported MP3, M4A, Ogg, or WAV file")
}

func wavDuration(content []byte) (int64, error) {
	var byteRate uint32
	var dataSize uint32
	for offset := 12; offset+8 <= len(content); {
		size := int(binary.LittleEndian.Uint32(content[offset+4 : offset+8]))
		end := offset + 8 + size
		if end > len(content) {
			return 0, errors.New("the WAV file is truncated")
		}
		switch string(content[offset : offset+4]) {
		case "fmt ":
			if size >= 12 {
				byteRate = binary.LittleEndian.Uint32(content[offset+16 : offset+20])
			}
		case "data":
			dataSize = uint32(size)
		}
		offset = end + size%2
	}
	if byteRate == 0 || dataSize == 0 {
		return 0, errors.New("the WAV file has no playable audio data")
	}
	return int64(dataSize) * 1000 / int64(byteRate), nil
}

func oggDuration(content []byte) (int64, error) {
	var sampleRate uint32
	var finalGranule uint64
	for offset := 0; offset+27 <= len(content); {
		if string(content[offset:offset+4]) != "OggS" {
			return 0, errors.New("the Ogg file contains an invalid page")
		}
		segments := int(content[offset+26])
		if offset+27+segments > len(content) {
			return 0, errors.New("the Ogg file is truncated")
		}
		bodySize := 0
		for _, size := range content[offset+27 : offset+27+segments] {
			bodySize += int(size)
		}
		bodyStart := offset + 27 + segments
		bodyEnd := bodyStart + bodySize
		if bodyEnd > len(content) {
			return 0, errors.New("the Ogg file is truncated")
		}
		if sampleRate == 0 {
			body := content[bodyStart:bodyEnd]
			if len(body) >= 16 && body[0] == 1 && string(body[1:7]) == "vorbis" {
				sampleRate = binary.LittleEndian.Uint32(body[12:16])
			} else if len(body) >= 8 && string(body[:8]) == "OpusHead" {
				sampleRate = 48000
			}
		}
		granule := binary.LittleEndian.Uint64(content[offset+6 : offset+14])
		if granule != ^uint64(0) && granule > finalGranule {
			finalGranule = granule
		}
		offset = bodyEnd
	}
	if sampleRate == 0 || finalGranule == 0 {
		return 0, errors.New("the Ogg duration could not be determined")
	}
	return int64(finalGranule) * 1000 / int64(sampleRate), nil
}

func mp4Duration(content []byte) (int64, error) {
	for offset := 0; offset+8 <= len(content); {
		size := int(binary.BigEndian.Uint32(content[offset : offset+4]))
		header := 8
		if size == 1 {
			if offset+16 > len(content) {
				break
			}
			size = int(binary.BigEndian.Uint64(content[offset+8 : offset+16]))
			header = 16
		}
		if size < header || offset+size > len(content) {
			return 0, errors.New("the M4A file contains an invalid box")
		}
		boxType := string(content[offset+4 : offset+8])
		if boxType == "moov" {
			if duration, ok := findMVHDDuration(content[offset+header : offset+size]); ok {
				return duration, nil
			}
		}
		offset += size
	}
	return 0, errors.New("the M4A duration could not be determined")
}

func findMVHDDuration(content []byte) (int64, bool) {
	for offset := 0; offset+8 <= len(content); {
		size := int(binary.BigEndian.Uint32(content[offset : offset+4]))
		if size < 8 || offset+size > len(content) {
			return 0, false
		}
		if string(content[offset+4:offset+8]) == "mvhd" {
			data := content[offset+8 : offset+size]
			if len(data) < 20 {
				return 0, false
			}
			if data[0] == 1 {
				if len(data) < 32 {
					return 0, false
				}
				scale := binary.BigEndian.Uint32(data[20:24])
				duration := binary.BigEndian.Uint64(data[24:32])
				return scaledDuration(duration, scale)
			}
			scale := binary.BigEndian.Uint32(data[12:16])
			duration := binary.BigEndian.Uint32(data[16:20])
			return scaledDuration(uint64(duration), scale)
		}
		offset += size
	}
	return 0, false
}

func scaledDuration(duration uint64, scale uint32) (int64, bool) {
	if scale == 0 || duration == 0 {
		return 0, false
	}
	return int64(duration * 1000 / uint64(scale)), true
}

func looksLikeMP3(content []byte) bool {
	for index := 0; index+1 < len(content) && index < 4096; index++ {
		if content[index] == 0xff && content[index+1]&0xe0 == 0xe0 {
			return true
		}
	}
	return false
}

func mp3Duration(content []byte) (int64, error) {
	offset := 0
	if len(content) >= 10 && string(content[:3]) == "ID3" {
		offset = 10 + int(content[6]&0x7f)<<21 + int(content[7]&0x7f)<<14 + int(content[8]&0x7f)<<7 + int(content[9]&0x7f)
	}
	var samples int64
	var sampleRate int
	frames := 0
	for offset+4 <= len(content) {
		header := binary.BigEndian.Uint32(content[offset : offset+4])
		versionBits := (header >> 19) & 3
		layerBits := (header >> 17) & 3
		bitrateIndex := (header >> 12) & 15
		rateIndex := (header >> 10) & 3
		padding := int((header >> 9) & 1)
		if header&0xffe00000 != 0xffe00000 || versionBits == 1 || layerBits != 1 || bitrateIndex == 0 || bitrateIndex == 15 || rateIndex == 3 {
			offset++
			continue
		}
		version := 1
		if versionBits == 2 {
			version = 2
		} else if versionBits == 0 {
			version = 25
		}
		rates := []int{44100, 48000, 32000}
		rate := rates[rateIndex]
		if version == 2 {
			rate /= 2
		} else if version == 25 {
			rate /= 4
		}
		var ratesKbps []int
		if version == 1 {
			ratesKbps = []int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320}
		} else {
			ratesKbps = []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160}
		}
		bitrate := ratesKbps[bitrateIndex] * 1000
		frameSamples := 1152
		coefficient := 144
		if version != 1 {
			frameSamples = 576
			coefficient = 72
		}
		frameSize := coefficient*bitrate/rate + padding
		if frameSize < 4 || offset+frameSize > len(content) {
			break
		}
		sampleRate = rate
		samples += int64(frameSamples)
		frames++
		offset += frameSize
	}
	if frames == 0 || sampleRate == 0 {
		return 0, errors.New("the MP3 duration could not be determined")
	}
	return samples * 1000 / int64(sampleRate), nil
}

func validateDeclaredAsset(kind, mimeType string, content []byte) (string, map[string]any, error) {
	detected, metadata, err := inspectAsset(kind, content)
	if err != nil {
		return "", nil, err
	}
	if mimeType != detected && !(kind == "audio" && mimeType == "audio/wav" && detected == "audio/wav") {
		return "", nil, fmt.Errorf("declared MIME type %q does not match detected type %q", mimeType, detected)
	}
	return detected, metadata, nil
}

func InspectProfileImage(content []byte) (string, error) {
	mimeType, _, err := inspectAsset("image", content)
	return mimeType, err
}

func InspectProfileImageUpload(content []byte) (string, int, int, error) {
	mimeType, metadata, err := inspectAsset("image", content)
	if err != nil {
		return "", 0, 0, err
	}
	width, _ := metadata["width"].(int)
	height, _ := metadata["height"].(int)
	return mimeType, width, height, nil
}
