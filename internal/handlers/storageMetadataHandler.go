package handlers

import (
	"net/http"
	"parachute/internal/services"
)

func StorageMetadata(w http.ResponseWriter, r *http.Request) {
	storageMetadataService := services.NewStorageMetadataService()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(storageMetadataService.GetMetadata()))
}
