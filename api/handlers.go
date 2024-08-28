package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) handleGetPhotoByKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	// Get header data
	contentType := r.URL.Query().Get("Image-Type")

	// Use content type from header if provided, otherwise default to "image/avif"
	if contentType == "" {
		contentType = "image/big"
	}
	var photo []byte
	var err error

	switch contentType {
	case "image/small":
		w.Header().Set("Content-Type", "image/png")
		photo, err = s.s3.DownloadSmallPhotoByKey(key)
	default:
		w.Header().Set("Content-Type", "image/png")
		photo, err = s.s3.DownloadPhotoByKey(key)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(photo)

}

func (s *Server) handleGetLastXPhotosForUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	photoNum := r.URL.Query().Get("photo_num")
	photoCount, err := strconv.ParseInt(photoNum, 10, 64)
	if err != nil {
		http.Error(w, "Invalid photo_num parameter", http.StatusBadRequest)
		return
	}

	photos, err := s.s3.GetLastXPhotosForUser(userID, photoCount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photos)
}

func (s *Server) handleGetLastPhotoForUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")

	photo, err := s.s3.GetLastPhotoForUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photo)
}
