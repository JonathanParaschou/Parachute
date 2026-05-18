package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"parachute/internal/config"
)

type UploadService struct{}

type UploadRecord struct {
	ID           string    `json:"id"`
	OriginalName string    `json:"original_name"`
	StoredName   string    `json:"stored_name"`
	RootID       string    `json:"root_id"`
	RootPath     string    `json:"root_path"`
	ObjectPath   string    `json:"object_path"`
	MetadataPath string    `json:"metadata_path"`
	SizeBytes    int64     `json:"size_bytes"`
	ContentType  string    `json:"content_type"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

func NewUploadService() *UploadService {
	return &UploadService{}
}

func (s *UploadService) ProcessFile(r io.Reader, originalName, contentType string) (UploadRecord, error) {
	if r == nil {
		return UploadRecord{}, fmt.Errorf("uploaded file is empty")
	}

	cfg, err := config.Load()
	if err != nil {
		return UploadRecord{}, fmt.Errorf("load config: %w", err)
	}

	root, err := selectStorageRoot(cfg.StorageRoots)
	if err != nil {
		return UploadRecord{}, err
	}

	now := time.Now().UTC()
	id, err := newUploadID(now)
	if err != nil {
		return UploadRecord{}, err
	}

	safeName := safeFilename(originalName)
	storedName := id + "-" + safeName
	objectDir := filepath.Join(root.Path, "objects", now.Format("2006"), now.Format("01"))
	tempDir := filepath.Join(root.Path, "temp")
	metadataDir := filepath.Join(root.Path, "metadata")
	for _, dir := range []string{objectDir, tempDir, metadataDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return UploadRecord{}, fmt.Errorf("prepare storage directory: %w", err)
		}
	}

	tempPath := filepath.Join(tempDir, storedName+".uploading")
	objectPath := filepath.Join(objectDir, storedName)
	metadataPath := filepath.Join(metadataDir, id+".json")

	tempFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return UploadRecord{}, fmt.Errorf("create temp upload: %w", err)
	}

	size, copyErr := io.Copy(tempFile, r)
	closeErr := tempFile.Close()
	if copyErr != nil {
		_ = os.Remove(tempPath)
		return UploadRecord{}, fmt.Errorf("write upload: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(tempPath)
		return UploadRecord{}, fmt.Errorf("close upload: %w", closeErr)
	}
	if size == 0 {
		_ = os.Remove(tempPath)
		return UploadRecord{}, fmt.Errorf("uploaded file is empty")
	}
	if root.LimitBytes > 0 {
		used, err := rootObjectsSize(root.Path)
		if err != nil {
			_ = os.Remove(tempPath)
			return UploadRecord{}, fmt.Errorf("check storage usage: %w", err)
		}
		if uint64(size)+used > root.LimitBytes {
			_ = os.Remove(tempPath)
			return UploadRecord{}, fmt.Errorf("storage root limit exceeded")
		}
	}

	if err := os.Rename(tempPath, objectPath); err != nil {
		_ = os.Remove(tempPath)
		return UploadRecord{}, fmt.Errorf("persist upload: %w", err)
	}

	record := UploadRecord{
		ID:           id,
		OriginalName: originalName,
		StoredName:   storedName,
		RootID:       root.ID,
		RootPath:     root.Path,
		ObjectPath:   objectPath,
		MetadataPath: metadataPath,
		SizeBytes:    size,
		ContentType:  contentType,
		UploadedAt:   now,
	}

	b, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return UploadRecord{}, fmt.Errorf("encode upload metadata: %w", err)
	}
	if err := os.WriteFile(metadataPath, append(b, '\n'), 0o644); err != nil {
		return UploadRecord{}, fmt.Errorf("persist upload metadata: %w", err)
	}

	return record, nil
}

func (s *UploadService) ListUploads() ([]UploadRecord, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	records := []UploadRecord{}
	for _, root := range cfg.StorageRoots {
		if !root.Enabled {
			continue
		}
		metadataDir := filepath.Join(root.Path, "metadata")
		entries, err := os.ReadDir(metadataDir)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("read upload metadata: %w", err)
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
				continue
			}
			b, err := os.ReadFile(filepath.Join(metadataDir, entry.Name()))
			if err != nil {
				return nil, fmt.Errorf("read upload metadata file: %w", err)
			}
			var record UploadRecord
			if err := json.Unmarshal(b, &record); err != nil {
				return nil, fmt.Errorf("decode upload metadata file: %w", err)
			}
			records = append(records, record)
		}
	}

	sort.Slice(records, func(i, j int) bool {
		return records[i].UploadedAt.After(records[j].UploadedAt)
	})
	return records, nil
}

func selectStorageRoot(roots []config.StorageRoot) (config.StorageRoot, error) {
	for _, root := range roots {
		if root.Enabled {
			return root, nil
		}
	}
	return config.StorageRoot{}, fmt.Errorf("no enabled storage root configured")
}

func safeFilename(name string) string {
	name = strings.TrimSpace(filepath.Base(name))
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "upload.bin"
	}

	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	cleaned := strings.Trim(b.String(), ".-")
	if cleaned == "" {
		return "upload.bin"
	}
	return cleaned
}

func newUploadID(t time.Time) (string, error) {
	var random [6]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", t.Format("20060102T150405000000000Z"), hex.EncodeToString(random[:])), nil
}

func rootObjectsSize(rootPath string) (uint64, error) {
	var total uint64
	objectsDir := filepath.Join(rootPath, "objects")
	err := filepath.WalkDir(objectsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += uint64(info.Size())
		return nil
	})
	if os.IsNotExist(err) {
		return 0, nil
	}
	return total, err
}
