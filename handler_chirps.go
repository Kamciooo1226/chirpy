package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/Kamciooo1226/chirpy/internal/auth"
	"github.com/Kamciooo1226/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

var badWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusUnauthorized, "Authorization failed", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusUnauthorized, "Authorization failed", nil)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		respondWithError(w, http.StatusBadRequest, "Error decoding parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	cleaned_body := cleanBody(params.Body, badWords)

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned_body,
		UserID: userID,
	})
	if err != nil {
		log.Printf("Failed to create chirp.")
		respondWithError(w, http.StatusInternalServerError, "Failed to create chirp.", nil)
		return
	}

	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func cleanBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")

	for i, word := range words {
		if _, exists := badWords[strings.ToLower(word)]; exists {
			words[i] = "****"
		}
	}

	clean_body := strings.Join(words, " ")
	return clean_body
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorID := r.URL.Query().Get("author_id")
	sortParam := r.URL.Query().Get("sort")
	var chirps []database.Chirp
	var err error
	allChirps := []Chirp{}

	if authorID == "" {
		chirps, err = cfg.db.GetAllChirps(r.Context())
	} else {
		userUUID, err := uuid.Parse(authorID)
		if err != nil {
			log.Printf("Error parsing userID string: %v into UUID", authorID)
			respondWithError(w, http.StatusBadRequest, "An error occured when parsing author_id", nil)
			return
		}
		chirps, err = cfg.db.GetChirpsByUserID(r.Context(), userUUID)
	}

	if err != nil {
		msg := "Error retrieving chirps from the database"
		log.Print(msg)
		respondWithError(w, http.StatusInternalServerError, msg, nil)
		return
	}

	for _, chirp := range chirps {
		allChirps = append(allChirps, Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}

	// DB query returns the chirps ordered by create_time ascending, so we ONLY check if query param sort=desc and reverse the slice if so
	//  otherwise we return it sorted correctly by default
	if sortParam == "desc" {
		slices.Reverse(allChirps)
	}

	respondWithJSON(w, http.StatusOK, allChirps)

}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		msg := fmt.Sprintf("The specified chirp_id: %v is not a valid chirp ID", r.PathValue("chirpID"))
		log.Print(msg)
		respondWithError(w, http.StatusBadRequest, msg, nil)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		msg := fmt.Sprintf("The specified chirp_id: %v does not exist", chirpID)
		log.Print(msg)
		respondWithError(w, http.StatusNotFound, msg, nil)
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})

}

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, r *http.Request) {
	bearerToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusUnauthorized, "Authorization failed", err)
		return
	}

	userID, err := auth.ValidateJWT(bearerToken, cfg.secret)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusUnauthorized, "Authorization failed", nil)
		return
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		msg := fmt.Sprintf("The specified chirp_id: %v is not a valid chirp ID", r.PathValue("chirpID"))
		log.Print(msg)
		respondWithError(w, http.StatusBadRequest, msg, nil)
		return
	}

	chirp, err := cfg.db.GetChirp(r.Context(), chirpID)
	if err != nil {
		msg := fmt.Sprintf("The specified chirp_id: %v does not exist", chirpID)
		log.Print(msg)
		respondWithError(w, http.StatusNotFound, msg, nil)
		return
	}

	if userID != chirp.UserID {
		log.Printf("Forbidden: user %v attempted to delete chirp %v owned by %v", userID, chirpID, chirp.UserID)
		respondWithError(w, http.StatusForbidden, "You can't delete this chirp", nil)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		log.Print(err)
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("An error occurred while deleting chirp: %v", chirp), nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
