package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"parachute/internal/config"
	"parachute/internal/services"
)

func StorageMetadata(w http.ResponseWriter, r *http.Request) {
	storageMetadataService := services.NewStorageMetadataService()
	metadata, err := storageMetadataService.GetMetadata()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metadata)
}

func StorageRoots(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cfg.StorageRoots)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("uploaded_file")
	if err != nil {
		log.Println("Error reading form data:", err)
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Deserializing File
	bodyBytes, err := io.ReadAll(file)
	if err != nil {
		log.Println("Error reading uploaded file:", err)
		http.Error(w, "Error reading uploaded file", http.StatusInternalServerError)
		return
	}

	// Uploading File
	uploadService := services.NewUploadService()
	bytesUploaded, err := uploadService.ProcessFile(bodyBytes)
	if err != nil {
		log.Println("Error processing file:", err)
		http.Error(w, "Error uploading file", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded (%d bytes).", bytesUploaded)
}
