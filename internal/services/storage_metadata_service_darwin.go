//go:build darwin

package services

import (
	"fmt"
	"golang.org/x/sys/unix"
)

type Drive struct {
	Name      string
	Model     string
	Serial    string
	SizeBytes uint64
	Bus       string
	IsSSD     *bool
}

type StorageMetadata struct {
	Platform         string
	TotalStorage     uint64
	UsedStorage      uint64
	FreeStorage      uint64
	AvailableStorage uint64
	Drives           []Drive
	DiskPath  string
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
