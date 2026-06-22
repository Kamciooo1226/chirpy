package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Kamciooo1226/chirpy/internal/auth"
	"github.com/Kamciooo1226/chirpy/internal/database"
	"github.com/google/uuid"
)

type UserLogin struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
}

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
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

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		log.Printf("An error occurred when creating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "An error occurred when creating JWT", nil)
		return
	}

	refreshToken, err := cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     auth.MakeRefreshToken(),
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
	})
	if err != nil {
		log.Printf("An error occurred when creating Refresh Token: %v", err)
		respondWithError(w, http.StatusInternalServerError, "An error occurred when creating Refresh Token", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, UserLogin{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken.Token,
		IsChirpyRed:  user.IsChirpyRed,
	})

}
