//go:build darwin

package services

import (
	"fmt"
	"golang.org/x/sys/unix"
)

type Drive struct {
	Name      string `json:"name"`
	Model     string `json:"model"`
	Serial    string `json:"serial"`
	SizeBytes uint64 `json:"size_bytes"`
	Bus       string `json:"bus"`
	IsSSD     *bool  `json:"is_ssd"`
}

type StorageMetadata struct {
	Platform         string  `json:"platform"`
	TotalStorage     uint64  `json:"total_storage"`
	UsedStorage      uint64  `json:"used_storage"`
	FreeStorage      uint64  `json:"free_storage"`
	AvailableStorage uint64  `json:"available_storage"`
	Drives           []Drive `json:"drives"`
	DiskPath         string  `json:"disk_path"`
}

func NewStorageMetadataService() *StorageMetadata {
	return &StorageMetadata{}
}

func (s *StorageMetadata) GetMetadata() (StorageMetadata, error) {
	total, free, avail, platform, err := diskUsageBytes("/")
	if err != nil {
		return *s, fmt.Errorf("storage metadata error: %v", err)
	}

	s.Platform = platform
	s.TotalStorage = total
	s.UsedStorage = total - free
	s.FreeStorage = free
	s.AvailableStorage = avail
	s.Drives = []Drive{}

	return *s, nil
}

func diskUsageBytes(path string) (total, free, avail uint64, platform string, err error) {
	var st unix.Statfs_t
	if err = unix.Statfs(path, &st); err != nil {
		return 0, 0, 0, "darwin", err
	}

	bsize := uint64(st.Bsize)
	total = st.Blocks * bsize
	free = st.Bfree * bsize
	avail = st.Bavail * bsize

	return total, free, avail, "darwin", nil
}
