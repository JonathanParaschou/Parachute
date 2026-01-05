//go:build windows

package services

import (
	"fmt"
	"os"
	"strings"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows"
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
}

func NewStorageMetadataService() *StorageMetadata {
	return &StorageMetadata{}
}

func (s *StorageMetadata) GetMetadata() (StorageMetadata, error) {
	root := defaultWindowsPath() // e.g. "C:\\"
	total, free, avail, platform, err := diskUsageBytes(root)
	if err != nil {
		return *s, fmt.Errorf("storage metadata error: %v", err)
	}

	used := total - free
	s.Platform = platform
	s.TotalStorage = total
	s.UsedStorage = used
	s.FreeStorage = free
	s.AvailableStorage = avail
	s.Drives, _ = s.GetDrives()

	return *s, nil
}

func (s *StorageMetadata) GetDrives() ([]Drive, error) {
	// Prefer modern Storage provider (best for SSD/HDD + bus)
	storage, storageErr := queryMSFTPhysicalDisks()
	// Fallback / enrichment
	win32, win32Err := queryWin32DiskDrives()

	// If both fail, return an error
	if storageErr != nil && win32Err != nil {
		return nil, fmt.Errorf("WMI failed: MSFT_PhysicalDisk=%v; Win32_DiskDrive=%v", storageErr, win32Err)
	}

	// Index Win32 by index
	win32ByIndex := map[uint32]win32DiskDrive{}
	for _, d := range win32 {
		win32ByIndex[d.Index] = d
	}

	// If storage works, use it as primary output
	if storageErr == nil && len(storage) > 0 {
		out := make([]Drive, 0, len(storage))
		for _, pd := range storage {
			w := win32ByIndex[pd.DeviceId]

			model := strings.TrimSpace(pd.FriendlyName)
			if model == "" {
				model = strings.TrimSpace(w.Model)
			}

			serial := strings.TrimSpace(pd.SerialNumber)
			if serial == "" {
				serial = strings.TrimSpace(w.SerialNumber)
			}

			size := pd.Size
			if size == 0 {
				if parsed, ok := parseUint64(w.Size); ok {
					size = parsed
				}
			}

			name := strings.TrimSpace(w.DeviceID)
			if name == "" {
				// typical format: \\.\PHYSICALDRIVE0
				name = fmt.Sprintf(`\\.\PHYSICALDRIVE%d`, pd.DeviceId)
			}

			out = append(out, Drive{
				Name:      name,
				Model:     model,
				Serial:    serial,
				SizeBytes: size,
				Bus:       busTypeString(pd.BusType, w.InterfaceType),
				IsSSD:     mediaTypeToIsSSD(pd.MediaType),
			})
		}

		s.Drives = out
		return out, nil
	}

	// Win32-only fallback (SSD/HDD unknown)
	out := make([]Drive, 0, len(win32))
	for _, d := range win32 {
		sizeBytes, _ := parseUint64(d.Size)
		out = append(out, Drive{
			Name:      strings.TrimSpace(d.DeviceID),
			Model:     strings.TrimSpace(d.Model),
			Serial:    strings.TrimSpace(d.SerialNumber),
			SizeBytes: sizeBytes,
			Bus:       strings.TrimSpace(d.InterfaceType),
			IsSSD:     nil,
		})
	}

	s.Drives = out
	return out, nil
}

func diskUsageBytes(path string) (total, free, avail uint64, platform string, err error) {
	p, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, 0, "windows", err
	}

	// GetDiskFreeSpaceEx returns:
	// avail = free bytes available to the caller
	// total = total bytes
	// free  = total free bytes on the disk
	if err = windows.GetDiskFreeSpaceEx(p, &avail, &total, &free); err != nil {
		return 0, 0, 0, "windows", err
	}

	return total, free, avail, "windows", nil
}

func defaultWindowsPath() string {
	// Prefer SYSTEMDRIVE if set (usually "C:")
	if d := strings.TrimSpace(os.Getenv("SYSTEMDRIVE")); d != "" {
		if !strings.HasSuffix(d, `\`) {
			d += `\`
		}
		return d
	}
	return `C:\`
}

/* ---------------- WMI: physical drive metadata ---------------- */

type msftPhysicalDisk struct {
	DeviceId     uint32
	FriendlyName string
	SerialNumber string
	Size         uint64
	BusType      uint16
	MediaType    uint16
}

type win32DiskDrive struct {
	Index         uint32
	Model         string
	SerialNumber  string // may be empty on some systems
	Size          string // bytes as string
	DeviceID      string // \\.\PHYSICALDRIVE0
	InterfaceType string // "SCSI", "IDE", "USB" (rough)
}

func queryMSFTPhysicalDisks() ([]msftPhysicalDisk, error) {
	var dst []msftPhysicalDisk
	q := wmi.CreateQuery(&dst, "")
	// Storage namespace
	if err := wmi.QueryNamespace(q, &dst, `ROOT\Microsoft\Windows\Storage`); err != nil {
		return nil, err
	}
	return dst, nil
}

func queryWin32DiskDrives() ([]win32DiskDrive, error) {
	var dst []win32DiskDrive
	q := wmi.CreateQuery(&dst, "")
	if err := wmi.Query(q, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}

// MSFT_PhysicalDisk MediaType (common values):
// 0 = Unspecified, 3 = HDD, 4 = SSD, 5 = SCM
func mediaTypeToIsSSD(mediaType uint16) *bool {
	switch mediaType {
	case 4, 5:
		v := true
		return &v
	case 3:
		v := false
		return &v
	default:
		return nil
	}
}

// BusType mapping best-effort; fallback to Win32 InterfaceType when unknown.
func busTypeString(bus uint16, fallback string) string {
	switch bus {
	case 1:
		return "SCSI"
	case 2:
		return "ATAPI"
	case 3:
		return "ATA"
	case 4:
		return "IEEE 1394"
	case 6:
		return "Fibre Channel"
	case 7:
		return "USB"
	case 8:
		return "RAID"
	case 9:
		return "iSCSI"
	case 10:
		return "SAS"
	case 11:
		return "SATA"
	case 14:
		return "Virtual"
	case 16:
		return "Storage Spaces"
	case 17:
		return "NVMe"
	default:
		f := strings.TrimSpace(fallback)
		if f == "" {
			return "Unknown"
		}
		return f
	}
}

func parseUint64(s string) (uint64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	var n uint64
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + uint64(c-'0')
	}
	return n, true
}
