package handlers

import (
	"encoding/json"
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
	roots := cfg.StorageRoots
	if roots == nil {
		roots = []config.StorageRoot{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(roots)
}

func Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("uploaded_file")
	if err != nil {
		log.Println("Error reading form data:", err)
		http.Error(w, "Error uploading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadService := services.NewUploadService()
	record, err := uploadService.ProcessFile(file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		log.Println("Error processing file:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(record)
}

func Uploads(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uploadService := services.NewUploadService()
	uploads, err := uploadService.ListUploads()
	if err != nil {
		log.Println("Error listing uploads:", err)
		http.Error(w, "Error listing uploads", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(uploads)
}
