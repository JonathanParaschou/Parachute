package services

type Services struct {
	Heartbeat       *HeartbeatService
	StorageMetadata *StorageMetadata
	VPN             *VPNService
}

func NewServices() *Services {
	vpn, _ := NewVPNService() // Ignore error for now, handle in app
	return &Services{
		Heartbeat:       NewHeartbeatService(),
		StorageMetadata: NewStorageMetadataService(),
		VPN:             vpn,
	}
}
