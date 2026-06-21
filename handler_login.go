package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Kamciooo1226/chirpy/internal/auth"
	"github.com/google/uuid"
)

type UserLogin struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "Error decoding parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		log.Printf("Error retrieving user data for email: %s", params.Email)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error comparing provided password and hashed password for email: %s", params.Email)
		respondWithError(w, http.StatusInternalServerError, "An error occured while trying to log in", nil)
		return
	}

	if !match {
		log.Printf("Unsuccessful login attempt for email: %s", params.Email)
		respondWithError(w, http.StatusUnauthorized, "Incorrect email or password", nil)
		return
	}

	const maxExpiryTime int = 3600
	expiresIn := params.ExpiresInSeconds
	if expiresIn == 0 || expiresIn > maxExpiryTime {
		expiresIn = maxExpiryTime
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(expiresIn)*time.Second)
	if err != nil {
		log.Printf("An error occurred when creating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "An error occurred when creating JWT", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, UserLogin{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
	})

}
