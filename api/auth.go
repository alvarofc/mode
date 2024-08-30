package api

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/alvarofc/mode/types"
	"github.com/golang-jwt/jwt"

	"golang.org/x/crypto/bcrypt"
)

// Global variables for RSA keys
var (
	signKey   *rsa.PrivateKey
	verifyKey *rsa.PublicKey
)

// InitializeKeys loads the RSA keys from environment variables
func InitializeKeys() error {
	log.Println("Starting to initialize keys...")

	// Read private key from environment variable
	privateKeyPEM := os.Getenv("RSA_PRIVATE_KEY")
	if privateKeyPEM == "" {
		return fmt.Errorf("RSA_PRIVATE_KEY environment variable is not set")
	}

	var err error
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Read public key from environment variable
	publicKeyPEM := os.Getenv("RSA_PUBLIC_KEY")
	if publicKeyPEM == "" {
		return fmt.Errorf("RSA_PUBLIC_KEY environment variable is not set")
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	if signKey == nil {
		return fmt.Errorf("signKey is nil after initialization")
	}
	if verifyKey == nil {
		return fmt.Errorf("verifyKey is nil after initialization")
	}

	log.Println("Keys initialized successfully")
	return nil
}

func (s *Server) handleSignUp(w http.ResponseWriter, r *http.Request) {
	var user types.User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validate user input
	if user.Email == "" || user.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Create user
	err = s.store.CreateUser(user.Email, user.Password)
	if err != nil {
		http.Error(w, "Error creating user: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}

func (s *Server) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var creds types.User
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := s.store.GetUserByEmail(creds.Email)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   strconv.Itoa(user.ID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(signKey)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	// Get the host from the request
	host := r.Host
	// If you're behind a proxy, you might need to use:
	// host := r.Header.Get("X-Forwarded-Host")

	isSecure := r.TLS != nil
	// If you're behind a proxy, you might need to check a header instead:
	// isSecure := r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     "mode_session",
		Value:    tokenString,
		Expires:  expirationTime,
		HttpOnly: true,
		Secure:   isSecure,             // Only set Secure flag if using HTTPS
		SameSite: http.SameSiteLaxMode, // Try Lax mode
		Path:     "/",
		Domain:   host, // Set the domain explicitly
	})

	// Log cookie setting for debugging
	log.Printf("Setting cookie for domain: %s, secure: %v", host, isSecure)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Successfully signed in"})
}

// VerifyToken verifies the JWT token
func VerifyToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return verifyKey, nil
	})
}
