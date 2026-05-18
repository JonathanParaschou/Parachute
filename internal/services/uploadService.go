package services

import "fmt"

type UploadService struct{}

func NewUploadService() *UploadService {
	return &UploadService{}
}

func (s *UploadService) ProcessFile(fileData []byte) (int, error) {
	if len(fileData) == 0 {
		return 0, fmt.Errorf("uploaded file is empty")
	}
	return len(fileData), nil
}
