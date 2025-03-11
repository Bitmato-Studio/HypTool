package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// ==================== DATA STRUCTURES ====================

// Blueprint holds your main application data.
type Blueprint struct {
	ID      string     `json:"id"`
	Version int        `json:"version"`
	Name    string     `json:"name"`
	Image   *ImageData `json:"image"`
	Author  string     `json:"author"`
	URL     string     `json:"url"`
	Desc    string     `json:"desc"`
	Model   string     `json:"model"`
	Script  string     `json:"script"`
	Props   PropsMap   `json:"props"`

	Preload bool `json:"preload"`
	Public  bool `json:"public"`
	Locked  bool `json:"locked"`
	Unique  bool `json:"unique"`
	Frozen  bool `json:"frozen"` // If Locked == true, set Frozen = true
}

// ImageData corresponds to blueprint.image.
type ImageData struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type AppMetaData struct {
	RequiredMods []string          `json:"required_mods,omitempty"`
	SourceMap    map[string]string `json:"source_map,omitempty"`
}

// PropsMap allows both string values and nested objects.
type PropsMap map[string]any

// Asset holds metadata plus the actual in-memory file data.
type Asset struct {
	Type     string        `json:"type"`
	URL      string        `json:"url"`
	Size     int           `json:"size"`
	Mime     string        `json:"mime"`
	FileData []byte        `json:"-"` // The raw embedded bytes (not in JSON)
	MemFile  *bytes.Reader `json:"-"` // In-memory reader (not in JSON)
}

// HypeHeader is the JSON structure stored in the .hyp file.
type HypeHeader struct {
	Blueprint *Blueprint   `json:"blueprint"`
	Assets    []Asset      `json:"assets"`
	Meta      *AppMetaData `json:"meta,omitempty"`
}

// ==================== EXPORT ====================

// ExportApp takes a Blueprint and returns a single `.hyp` byte slice.
func ExportApp(bp *Blueprint, existingAssets []Asset) ([]byte, string, error) {
	// If locked, set frozen
	if bp.Locked {
		bp.Frozen = true
	}

	// Choose a filename
	filename := "app.hyp"
	if bp.Name != "" {
		filename = bp.Name + ".hyp"
	}

	// Gather assets by extracting correct file data & updating URLs to their SHA-256 hashes
	fmt.Printf("Size of assets %d\n", len(existingAssets))

	// Build the JSON header (metadata only)
	header := HypeHeader{
		Blueprint: bp,
		Assets:    make([]Asset, len(existingAssets)),
	}
	for i, a := range existingAssets {
		fmt.Printf("Bundling: %s\n", a.URL)
		header.Assets[i] = Asset{
			Type: a.Type,
			URL:  a.URL,
			Size: len(a.FileData),
			Mime: a.Mime,
		}
	}

	// 1) Encode header to JSON
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal header: %w", err)
	}

	// 2) Construct the final `.hyp` data
	headerLen := uint32(len(headerBytes))
	finalData := make([]byte, 4+headerLen)
	binary.LittleEndian.PutUint32(finalData[0:4], headerLen)
	copy(finalData[4:], headerBytes)

	// 3) Append raw file data in sequence
	for _, a := range existingAssets {
		finalData = append(finalData, a.FileData...)
	}

	return finalData, filename, nil
}

func resolveMime(asset *Asset) string {
	var mimeType string

	switch asset.Type {
	case "script":
		mimeType = "application/javascript"
	case "avatar":
		mimeType = "application/octet-stream"
	case "model":
		mimeType = "model/gltf-binary"
	case "emote":
		mimeType = "model/gltf-binary"
	case "texture":
		parts := strings.Split(asset.URL, ".")
		ext := parts[len(parts)-1]
		mimeType := "image/" + ext
		return mimeType
	default:
		mimeType := "application/octet-stream"
		return mimeType
	}

	asset.Mime = mimeType
	return mimeType
}

func resolvePath(asset *Asset) string {
	url := "asset://" + hashBytes(asset.FileData)

	switch asset.Type {
	case "script":
		url += ".js"
		break
	case "avatar":
		url += ".vrm"
		break
	case "model":
		url += ".glb"
		break
	case "emote":
		url += ".glb"
		break
	case "texture":
		ext := strings.Split(asset.URL, ".")[1]
		url += "." + ext
		break
	case "hdr":
		url += ".hdr"
		break
	case "audio":
		url += ".mp3"
		break
	default:
		fmt.Printf("Failed to get file extension for type %s\n", asset.Type)

	}

	asset.URL = url
	return url
}

// ==================== HASHING ====================

// hashFile computes the SHA-256 hash of a file.
func hashFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// hashBytes computes the SHA-256 hash of a byte slice.
func hashBytes(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// ==================== IMPORT ====================

// ImportApp reads `.hyp` data, extracts the JSON header, then loads all asset data into memory.
func ImportApp(blob []byte) (*Blueprint, []Asset, error) {
	if len(blob) < 4 {
		return nil, nil, errors.New("invalid .hyp data: missing header length")
	}

	// 1) Read the 4-byte header length
	headerLen := binary.LittleEndian.Uint32(blob[0:4])
	if len(blob) < int(4+headerLen) {
		return nil, nil, errors.New("invalid .hyp data: truncated header JSON")
	}

	// 2) Parse the JSON portion
	headerBytes := blob[4 : 4+headerLen]
	var hdr HypeHeader
	if err := json.Unmarshal(headerBytes, &hdr); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}
	if hdr.Blueprint == nil {
		return nil, nil, errors.New("header missing blueprint")
	}

	// 3) Extract asset data
	assets := make([]Asset, len(hdr.Assets))
	pos := 4 + headerLen

	for i, meta := range hdr.Assets {
		if pos+uint32(meta.Size) > uint32(len(blob)) {
			return nil, nil, errors.New("invalid .hyp data: not enough bytes for asset data")
		}
		dataChunk := blob[pos : pos+uint32(meta.Size)]
		pos += uint32(meta.Size)

		// Create an Asset with in-memory file representation
		assets[i] = Asset{
			Type:     meta.Type,
			URL:      meta.URL,
			Size:     meta.Size,
			Mime:     meta.Mime,
			FileData: dataChunk,
			MemFile:  bytes.NewReader(dataChunk),
		}
	}

	return hdr.Blueprint, assets, nil
}

func AddAssetToGroup(assets *[]Asset, data []byte, fType string) (Asset, error) {

	if assets == nil {
		return Asset{}, fmt.Errorf("Failed to add data to assets as assets is nil")
	}

	newAsset := Asset{
		Type:     fType,
		Size:     len(data),
		FileData: data,
		MemFile:  bytes.NewReader(data),
	}

	resolveMime(&newAsset)

	resolvePath(&newAsset)

	*assets = append(*assets, newAsset)
	fmt.Printf("Added %s to assets\n", newAsset.URL)
	return newAsset, nil
}
