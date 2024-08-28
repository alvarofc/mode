package api

import (
	"encoding/json"
	"net/http"

	"github.com/alvarofc/mode/storage"
)

type Server struct {
	listenAddr string
	store      storage.Storage
	s3         storage.S3
}

func NewServer(listenAddr string, pg storage.Storage, s3 storage.S3) *Server {
	return &Server{
		listenAddr: listenAddr,
		store:      pg,
		s3:         s3,
	}
}

func (s *Server) Start() error {
	// Routes that only need logging
	http.HandleFunc("POST /signup", s.loggingMiddleware(s.handleSignUp))
	http.HandleFunc("POST /signin", s.loggingMiddleware(s.handleSignIn))

	// Routes that need both logging and authentication
	http.HandleFunc("GET /user", s.combineMiddleware(s.handleGetUserById, s.loggingMiddleware, s.authMiddleware))
	http.HandleFunc("GET /photo/{key}", s.combineMiddleware(s.handleGetPhotoByKey, s.loggingMiddleware, s.authMiddleware))
	http.HandleFunc("GET /user/{user_id}/photos", s.combineMiddleware(s.handleGetLastXPhotosForUser, s.loggingMiddleware, s.authMiddleware))
	http.HandleFunc("GET /user/{user_id}/photo", s.combineMiddleware(s.handleGetLastPhotoForUser, s.loggingMiddleware, s.authMiddleware))

	return http.ListenAndServe(s.listenAddr, nil)
}

func (s *Server) handleGetUserById(w http.ResponseWriter, r *http.Request) {
	user, err := s.store.GetUserById(10)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}
