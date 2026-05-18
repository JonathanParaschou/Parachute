package services

type Services struct {
	Heartbeat       *HeartbeatService
	StorageMetadata *StorageMetadata
}

func NewServices() *Services {
	return &Services{
		Heartbeat:       NewHeartbeatService(),
		StorageMetadata: NewStorageMetadataService(),
	}
}
