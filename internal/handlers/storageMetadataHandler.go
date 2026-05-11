package handlers

import (
	"encoding/json"
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

func Upload(w http.ResposeWriter, r *http.Request) {
	file, _, err := r.FormFile("uploaded_file")

	if err != nil {
        log.Println("Error reading form data:", err)
        http.Error(w, "Error uploading file", http.StatusBadRequest) 
        return
    }

	// Deserializing File
    bodyBytes, err := ioutil.ReadAll(file)
    if err != nil {
        log.Println("Error reading uploaded file:", err)
        http.Error(w, "Error reading uploaded file", http.StatusInternalServerError) 
        return
    }

	// Uploading File
    uploadService := services.NewUploadService()
    response, err := uploadService.ProcessFile(bodyBytes)
    if err != nil {
        log.Println("Error processing file:", err) 
        http.Error(w, "Error uploading file", http.StatusInternalServerError) 
        return
    }

    // Respond with success
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("File uploaded."))
}