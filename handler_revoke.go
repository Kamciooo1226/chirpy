package main

import (
	"log"
	"net/http"

	"github.com/Kamciooo1226/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusUnauthorized, err.Error(), err)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusInternalServerError, "An error occurred while revoking token", nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
