package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Kamciooo1226/chirpy/internal/auth"
)

type RefreshResponse struct {
	Token string `json:"token"`
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	userID, err := cfg.db.GetUserIDFromRefreshToken(r.Context(), token)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusUnauthorized, "Invalid or expired token", nil)
		return
	}

	newToken, err := auth.MakeJWT(userID, cfg.secret, time.Hour)
	if err != nil {
		log.Printf("An error occurred when creating JWT: %v", err)
		respondWithError(w, http.StatusInternalServerError, "An error occurred when creating JWT", nil)
		return
	}

	respondWithJSON(w, http.StatusOK, RefreshResponse{
		Token: newToken,
	})

}
