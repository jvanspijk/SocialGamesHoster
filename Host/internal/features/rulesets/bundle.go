package rulesets

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

const (
	BundleFormatVersion = 1
	MaxBundleSize       = 25 << 20
	MaxBundleFiles      = 100
	MaxRulesetJSONSize  = 2 << 20
)

type BundleManifest struct {
	FormatVersion             int                   `json:"formatVersion"`
	SourceApplicationVersion  string                `json:"sourceApplicationVersion"`
	MinimumApplicationVersion string                `json:"minimumApplicationVersion"`
	LogicalSourceRulesetID    string                `json:"logicalSourceRulesetId"`
	SourceVersionNumber       int                   `json:"sourceVersionNumber"`
	Name                      string                `json:"name"`
	Description               string                `json:"description"`
	RulesetChecksum           string                `json:"rulesetChecksum"`
	Assets                    []BundleAssetManifest `json:"assets"`
}

type BundleAssetManifest struct {
	Path     string `json:"path"`
	AssetKey string `json:"assetKey"`
	Kind     string `json:"kind"`
	MIMEType string `json:"mimeType"`
	ByteSize int64  `json:"byteSize"`
	Checksum string `json:"checksum"`
}

type ImportedBundle struct {
	Manifest   BundleManifest
	Definition DefinitionV1
	Assets     map[string][]byte
}

func ReadBundle(data []byte) (ImportedBundle, error) {
	if len(data) == 0 || len(data) > MaxBundleSize {
		return ImportedBundle{}, fmt.Errorf("bundle must be between 1 byte and %d bytes", MaxBundleSize)
	}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return ImportedBundle{}, fmt.Errorf("invalid bundle ZIP: %w", err)
	}
	if len(reader.File) < 2 || len(reader.File) > MaxBundleFiles+2 {
		return ImportedBundle{}, fmt.Errorf("bundle has an invalid file count")
	}

	files := map[string][]byte{}
	var total int64
	for _, file := range reader.File {
		if err := validateBundlePath(file); err != nil {
			return ImportedBundle{}, err
		}
		if _, duplicate := files[file.Name]; duplicate {
			return ImportedBundle{}, fmt.Errorf("bundle contains duplicate path %q", file.Name)
		}
		if file.UncompressedSize64 > MaxBundleSize || total+int64(file.UncompressedSize64) > MaxBundleSize {
			return ImportedBundle{}, fmt.Errorf("bundle exceeds the decompressed size limit")
		}
		stream, err := file.Open()
		if err != nil {
			return ImportedBundle{}, err
		}
		content, readErr := io.ReadAll(io.LimitReader(stream, MaxBundleSize+1))
		closeErr := stream.Close()
		if readErr != nil {
			return ImportedBundle{}, readErr
		}
		if closeErr != nil {
			return ImportedBundle{}, closeErr
		}
		total += int64(len(content))
		files[file.Name] = content
	}

	manifestBytes, ok := files["manifest.json"]
	if !ok {
		return ImportedBundle{}, fmt.Errorf("bundle is missing manifest.json")
	}
	rulesetBytes, ok := files["ruleset.json"]
	if !ok {
		return ImportedBundle{}, fmt.Errorf("bundle is missing ruleset.json")
	}
	if len(rulesetBytes) > MaxRulesetJSONSize {
		return ImportedBundle{}, fmt.Errorf("ruleset.json exceeds the size limit")
	}

	var manifest BundleManifest
	if err := decodeStrictJSON(manifestBytes, &manifest); err != nil {
		return ImportedBundle{}, fmt.Errorf("invalid manifest.json: %w", err)
	}
	if manifest.FormatVersion != BundleFormatVersion {
		return ImportedBundle{}, fmt.Errorf("unsupported bundle format version %d", manifest.FormatVersion)
	}
	if checksum(rulesetBytes) != strings.ToLower(manifest.RulesetChecksum) {
		return ImportedBundle{}, fmt.Errorf("ruleset.json checksum does not match the manifest")
	}

	declared := map[string]BundleAssetManifest{}
	assetKeys := map[string]struct{}{}
	for _, asset := range manifest.Assets {
		if !strings.HasPrefix(asset.Path, "assets/") {
			return ImportedBundle{}, fmt.Errorf("asset %q is outside the assets directory", asset.Path)
		}
		if _, duplicate := declared[asset.Path]; duplicate {
			return ImportedBundle{}, fmt.Errorf("asset path %q is declared more than once", asset.Path)
		}
		if _, duplicate := assetKeys[asset.AssetKey]; duplicate {
			return ImportedBundle{}, fmt.Errorf("asset key %q is declared more than once", asset.AssetKey)
		}
		declared[asset.Path] = asset
		assetKeys[asset.AssetKey] = struct{}{}
		content, exists := files[asset.Path]
		if !exists {
			return ImportedBundle{}, fmt.Errorf("declared asset %q is missing", asset.Path)
		}
		if int64(len(content)) != asset.ByteSize || checksum(content) != strings.ToLower(asset.Checksum) {
			return ImportedBundle{}, fmt.Errorf("asset %q size or checksum does not match", asset.Path)
		}
		if err := validateAssetContent(asset, content); err != nil {
			return ImportedBundle{}, err
		}
	}
	for name := range files {
		if name == "manifest.json" || name == "ruleset.json" {
			continue
		}
		if _, declared := declared[name]; !declared {
			return ImportedBundle{}, fmt.Errorf("bundle contains undeclared file %q", name)
		}
	}

	var definition DefinitionV1
	if err := decodeStrictJSON(rulesetBytes, &definition); err != nil {
		return ImportedBundle{}, fmt.Errorf("invalid ruleset.json: %w", err)
	}
	report := Validate(definition, assetKeys)
	if !report.Valid() {
		return ImportedBundle{}, fmt.Errorf("ruleset validation failed: %s", report.Errors[0].Message)
	}

	assets := make(map[string][]byte, len(manifest.Assets))
	for _, asset := range manifest.Assets {
		assets[asset.AssetKey] = files[asset.Path]
	}
	return ImportedBundle{Manifest: manifest, Definition: definition, Assets: assets}, nil
}

func WriteBundle(manifest BundleManifest, definition DefinitionV1, assets map[string][]byte) ([]byte, error) {
	rulesetBytes, err := json.Marshal(definition)
	if err != nil {
		return nil, err
	}
	manifest.FormatVersion = BundleFormatVersion
	manifest.RulesetChecksum = checksum(rulesetBytes)
	for i := range manifest.Assets {
		content, ok := assets[manifest.Assets[i].AssetKey]
		if !ok {
			return nil, fmt.Errorf("asset %q is missing", manifest.Assets[i].AssetKey)
		}
		manifest.Assets[i].ByteSize = int64(len(content))
		manifest.Assets[i].Checksum = checksum(content)
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, err
	}

	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	if err := writeZipFile(writer, "manifest.json", manifestBytes); err != nil {
		return nil, err
	}
	if err := writeZipFile(writer, "ruleset.json", rulesetBytes); err != nil {
		return nil, err
	}
	for _, asset := range manifest.Assets {
		if err := writeZipFile(writer, asset.Path, assets[asset.AssetKey]); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	if output.Len() > MaxBundleSize {
		return nil, fmt.Errorf("generated bundle exceeds the size limit")
	}
	return output.Bytes(), nil
}

func validateBundlePath(file *zip.File) error {
	clean := path.Clean(file.Name)
	if file.Name == "" || strings.Contains(file.Name, "\\") || strings.HasPrefix(file.Name, "/") || clean != file.Name || clean == "." || strings.HasPrefix(clean, "../") {
		return fmt.Errorf("bundle contains unsafe path %q", file.Name)
	}
	if file.FileInfo().Mode()&os.ModeSymlink != 0 || file.FileInfo().IsDir() {
		return fmt.Errorf("bundle contains unsupported entry %q", file.Name)
	}
	return nil
}

func validateAssetContent(asset BundleAssetManifest, content []byte) error {
	switch asset.Kind {
	case "image":
		if len(content) > 2<<20 {
			return fmt.Errorf("image %q exceeds 2 MB", asset.Path)
		}
	case "audio":
		if len(content) > 5<<20 {
			return fmt.Errorf("audio %q exceeds 5 MB", asset.Path)
		}
	default:
		return fmt.Errorf("asset %q has unsupported kind %q", asset.Path, asset.Kind)
	}
	if _, _, err := validateDeclaredAsset(asset.Kind, asset.MIMEType, content); err != nil {
		return fmt.Errorf("asset %q: %w", asset.Path, err)
	}
	return nil
}

func decodeStrictJSON(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("unexpected trailing JSON")
	}
	return nil
}

func checksum(data []byte) string {
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func DefinitionChecksum(data []byte) string {
	return checksum(data)
}

func writeZipFile(writer *zip.Writer, name string, data []byte) error {
	entry, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = entry.Write(data)
	return err
}
