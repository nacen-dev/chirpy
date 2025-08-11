package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nacen-dev/chirpy/internal/database"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleCreateChirp(res http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if !isChirpValid(params.Body) {
		respondWithError(res, http.StatusBadRequest, "Chirp is too long", nil)
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: params.UserId,
	})

	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to create the chirp", nil)
	}

	respondWithJSON(res, http.StatusCreated, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      cleanProfaneWords(chirp.Body, profaneWords),
		UserID:    chirp.UserID,
	})

}

func isChirpValid(chirp string) bool {
	const maxChirpLength = 140

	return len(chirp) <= maxChirpLength
}
func cleanProfaneWords(s string, profaneWords map[string]struct{}) string {
	words := strings.Fields(s)
	for index, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := profaneWords[loweredWord]; ok {
			words[index] = "****"
		}
	}
	return strings.Join(words, " ")
}
