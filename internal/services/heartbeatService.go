package services

import "time"

type HeartbeatService struct{}

func NewHeartbeatService() *HeartbeatService {
	return &HeartbeatService{}
}

func (s *HeartbeatService) Ping() string {
	return "System Operational: " + time.Now().Format(time.RFC3339)
}
