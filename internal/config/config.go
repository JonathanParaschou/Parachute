package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AppDirName     = "parachute"
	ConfigFileName = "config.json"
	RootDirName    = "ParachuteStorage"
	RootManifest   = "parachute-root.json"
)

type Config struct {
	StorageRoots []StorageRoot `json:"storage_roots"`
}

type StorageRoot struct {
	ID         string    `json:"id"`
	Path       string    `json:"path"`
	LimitBytes uint64    `json:"limit_bytes"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
}

type RootManifestData struct {
	ID         string    `json:"id"`
	LimitBytes uint64    `json:"limit_bytes"`
	CreatedAt  time.Time `json:"created_at"`
}

func Path() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, AppDirName, ConfigFileName), nil
}

func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func AddStorageRoot(path string, limitBytes uint64) (StorageRoot, error) {
	if strings.TrimSpace(path) == "" {
		return StorageRoot{}, errors.New("storage path is required")
	}
	if limitBytes == 0 {
		return StorageRoot{}, errors.New("limit must be greater than zero")
	}

	managedPath, err := NormalizeStoragePath(path)
	if err != nil {
		return StorageRoot{}, err
	}

	cfg, err := Load()
	if err != nil {
		return StorageRoot{}, err
	}

	for _, root := range cfg.StorageRoots {
		if samePath(root.Path, managedPath) {
			return StorageRoot{}, fmt.Errorf("storage root already exists: %s", managedPath)
		}
	}

	root := StorageRoot{
		ID:         newRootID(managedPath),
		Path:       managedPath,
		LimitBytes: limitBytes,
		Enabled:    true,
		CreatedAt:  time.Now().UTC(),
	}

	if err := initializeRoot(root); err != nil {
		return StorageRoot{}, err
	}

	cfg.StorageRoots = append(cfg.StorageRoots, root)
	if err := Save(cfg); err != nil {
		return StorageRoot{}, err
	}

	return root, nil
}

func RemoveStorageRoot(path string) (StorageRoot, error) {
	managedPath, err := NormalizeStoragePath(path)
	if err != nil {
		return StorageRoot{}, err
	}

	cfg, err := Load()
	if err != nil {
		return StorageRoot{}, err
	}

	next := cfg.StorageRoots[:0]
	var removed StorageRoot
	for _, root := range cfg.StorageRoots {
		if samePath(root.Path, managedPath) {
			removed = root
			continue
		}
		next = append(next, root)
	}
	if removed.Path == "" {
		return StorageRoot{}, fmt.Errorf("storage root not found: %s", managedPath)
	}

	cfg.StorageRoots = next
	if err := Save(cfg); err != nil {
		return StorageRoot{}, err
	}
	return removed, nil
}

func NormalizeStoragePath(path string) (string, error) {
	cleaned := filepath.Clean(path)
	if filepath.Base(cleaned) != RootDirName {
		cleaned = filepath.Join(cleaned, RootDirName)
	}
	return filepath.Abs(cleaned)
}

func initializeRoot(root StorageRoot) error {
	for _, name := range []string{"objects", "metadata", "temp"} {
		if err := os.MkdirAll(filepath.Join(root.Path, name), 0o755); err != nil {
			return err
		}
	}

	manifest := RootManifestData{
		ID:         root.ID,
		LimitBytes: root.LimitBytes,
		CreatedAt:  root.CreatedAt,
	}
	b, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root.Path, RootManifest), append(b, '\n'), 0o644)
}

func samePath(a, b string) bool {
	return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
}

func newRootID(path string) string {
	name := strings.ToLower(filepath.Base(filepath.Dir(path)))
	name = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, name)
	name = strings.Trim(name, "-")
	if name == "" {
		name = "root"
	}
	return fmt.Sprintf("%s-%d", name, time.Now().UTC().UnixNano())
}
