//go:build linux

package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

type StorageMetadataService struct {
	Platform     string
	TotalStorage uint64
	UsedStorage  uint64
	FreeStorage  uint64
	Drives       []Drive
}

func NewStorageMetadataService() *StorageMetadataService {
	return &StorageMetadataService{}
}

func (s *StorageMetadataService) GetMetadata() string {
	total, free, avail, platform, err := diskUsageBytes("/")
	if err != nil {
		return fmt.Sprintf("Storage metadata error: %v", err)
	}

	used := total - free

	s.Platform = platform
	s.TotalStorage = total
	s.UsedStorage = used
	s.FreeStorage = free

	return fmt.Sprintf(
		"Storage Metadata OK (%s): total=%d used=%d free=%d available=%d",
		s.Platform, total, used, free, avail,
	)
}

func (s *StorageMetadataService) GetDrives() ([]Drive, error) {
	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return nil, err
	}

	var out []Drive
	for _, e := range entries {
		name := e.Name()

		// Skip loop/ram devices (common)
		if strings.HasPrefix(name, "loop") || strings.HasPrefix(name, "ram") {
			continue
		}

		base := filepath.Join("/sys/block", name)

		rotStr, err := readFileTrim(filepath.Join(base, "queue/rotational"))
		var isSSD *bool
		if err == nil {
			rot := rotStr == "1"
			v := !rot
			isSSD = &v
		}

		model, _ := readFileTrim(filepath.Join(base, "device/model"))

		// size is in 512-byte sectors on most devices
		secStr, err := readFileTrim(filepath.Join(base, "size"))
		var sizeBytes uint64
		if err == nil {
			secs, _ := strconv.ParseUint(secStr, 10, 64)
			sizeBytes = secs * 512
		}

		// Bus type heuristic
		bus := ""
		if strings.HasPrefix(name, "nvme") {
			bus = "nvme"
		} else if strings.HasPrefix(name, "sd") {
			bus = "scsi/sata/usb"
		}

		out = append(out, Drive{
			Name:      name,
			Model:     strings.TrimSpace(model),
			SizeBytes: sizeBytes,
			Bus:       bus,
			IsSSD:     isSSD,
		})
	}

	s.Drives = out
	return out, nil
}

func diskUsageBytes(path string) (total, free, avail uint64, platform string, err error) {
	var st unix.Statfs_t
	if err = unix.Statfs(path, &st); err != nil {
		return 0, 0, 0, "linux", err
	}

	bsize := uint64(st.Bsize)
	total = st.Blocks * bsize
	free = st.Bfree * bsize
	avail = st.Bavail * bsize

	return total, free, avail, "linux", nil
}

func readFileTrim(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
