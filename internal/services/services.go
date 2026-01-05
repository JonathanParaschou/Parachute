package services

type Services struct {
	Heartbeat          *HeartbeatService
	StorageMetadata, _ *StorageMetadata
}

func NewServices() *Services {
	return &Services{
		Heartbeat:       NewHeartbeatService(),
		StorageMetadata: NewStorageMetadataService(),
	}
}
