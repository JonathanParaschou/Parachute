package services

type Services struct {
	Heartbeat       *HeartbeatService
	StorageMetadata *StorageMetadataService
}

func NewServices() *Services {
	return &Services{
		Heartbeat:       NewHeartbeatService(),
		StorageMetadata: NewStorageMetadataService(),
	}
}
