package handlers

import (
	"encoding/json"
	"net/http"
	"parachute/internal/services"
)

func StorageMetadata(w http.ResponseWriter, r *http.Request) {
	storageMetadataService := services.NewStorageMetadataService()
	metadata, err := storageMetadataService.GetMetadata()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metadata)
}
