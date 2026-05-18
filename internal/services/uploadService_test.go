package services

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"parachute/internal/config"
)

func TestUploadPersistsObjectAndMetadata(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("APPDATA", configHome)
	t.Setenv("XDG_CONFIG_HOME", configHome)

	rootBase := t.TempDir()
	root, err := config.AddStorageRoot(rootBase, 1024*1024)
	if err != nil {
		t.Fatalf("add storage root: %v", err)
	}

	service := NewUploadService()
	record, err := service.ProcessFile(bytes.NewBufferString("hello parachute"), "../sample file.txt", "text/plain")
	if err != nil {
		t.Fatalf("process file: %v", err)
	}

	if record.RootID != root.ID {
		t.Fatalf("root ID = %q, want %q", record.RootID, root.ID)
	}
	if record.OriginalName != "../sample file.txt" {
		t.Fatalf("original name = %q", record.OriginalName)
	}
	if filepath.Base(record.StoredName) != record.StoredName {
		t.Fatalf("stored name includes path elements: %q", record.StoredName)
	}

	data, err := os.ReadFile(record.ObjectPath)
	if err != nil {
		t.Fatalf("read object: %v", err)
	}
	if string(data) != "hello parachute" {
		t.Fatalf("object content = %q", string(data))
	}

	if _, err := os.Stat(record.MetadataPath); err != nil {
		t.Fatalf("metadata was not persisted: %v", err)
	}

	uploads, err := service.ListUploads()
	if err != nil {
		t.Fatalf("list uploads: %v", err)
	}
	if len(uploads) != 1 {
		t.Fatalf("uploads len = %d, want 1", len(uploads))
	}
	if uploads[0].ID != record.ID {
		t.Fatalf("listed upload ID = %q, want %q", uploads[0].ID, record.ID)
	}
}
