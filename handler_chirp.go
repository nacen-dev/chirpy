package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nacen-dev/chirpy/internal/auth"
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
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)

	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "Unable to get bearer token", err)
		return
	}

	userIdFromJWT, err := auth.ValidateJWT(token, cfg.jwtSecret)

	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "Invalid jwt", err)
		return
	}

	if !isChirpValid(params.Body) {
		respondWithError(res, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	chirp, err := cfg.db.CreateChirp(req.Context(), database.CreateChirpParams{
		Body:   params.Body,
		UserID: userIdFromJWT,
	})

	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to create the chirp", nil)
		return
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

func (cfg *apiConfig) handleGetChirps(res http.ResponseWriter, req *http.Request) {
	author_id := req.URL.Query().Get("author_id")
	sortingOrder := req.URL.Query().Get("sort")
	var parsed_author_id uuid.UUID
	var err error

	if author_id != "" {
		parsed_author_id, err = uuid.Parse(author_id)
		if err != nil {
			respondWithError(res, http.StatusBadRequest, "Invalid author id", err)
			return
		}
	}

	chirps, chirpsError := cfg.db.GetChirps(req.Context(), database.GetChirpsParams{
		AuthorID: parsed_author_id,
		OrderBy:  sortingOrder,
	})

	if chirpsError != nil {
		respondWithError(res, http.StatusInternalServerError, "Unable to get chirps", err)
		return
	}

	response := []Chirp{}
	for _, dbChirp := range chirps {
		response = append(response, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		})
	}
	respondWithJSON(res, http.StatusOK, response)
}

func (cfg *apiConfig) handleGetChirpById(res http.ResponseWriter, req *http.Request) {
	chirpId, err := uuid.Parse(req.PathValue("chirpId"))
	if err != nil {
		respondWithError(res, http.StatusBadRequest, "Invalid chirp id", nil)
		return
	}

	dbChirp, err := cfg.db.GetChirpById(req.Context(), chirpId)
	if err != nil {
		respondWithError(res, http.StatusNotFound, "Couldn't retrieve chirp", nil)
		return
	}

	respondWithJSON(res, http.StatusOK, Chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	})
}

func (cfg *apiConfig) handleDeleteChirpById(res http.ResponseWriter, req *http.Request) {
	chirpId, err := uuid.Parse(req.PathValue("chirpId"))
	if err != nil {
		respondWithError(res, http.StatusBadRequest, "Chirp Id is missing", err)
		return
	}

	accessToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "No token found", err)
		return
	}

	userIdFromJWT, err := auth.ValidateJWT(accessToken, cfg.jwtSecret)

	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "Invalid jwt", err)
		return
	}

	chirp, err := cfg.db.GetChirpById(req.Context(), chirpId)
	if err != nil {
		respondWithError(res, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if chirp.UserID != userIdFromJWT {
		respondWithError(res, http.StatusForbidden, "Unauthorized", err)
		return
	}

	err = cfg.db.DeleteChirpById(req.Context(), chirp.ID)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Unable to delete chirp...", err)
		return
	}
	res.WriteHeader(http.StatusNoContent)
}
