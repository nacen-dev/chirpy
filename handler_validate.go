package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

func handleValidateChirp(res http.ResponseWriter, req *http.Request) {
	type chirpBody struct {
		Body string `json:"body"`
	}
	type responseBody struct {
		CleanedBody string `json:"cleaned_body"`
	}
	decoder := json.NewDecoder(req.Body)
	params := chirpBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	const maxChirpLength = 140
	if len(params.Body) > maxChirpLength {
		respondWithError(res, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}

	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	respondWithJSON(res, http.StatusOK, responseBody{
		CleanedBody: handleProfaneWords(params.Body, profaneWords),
	})
}

func handleProfaneWords(s string, profaneWords map[string]struct{}) string {
	words := strings.Fields(s)
	for index, word := range words {
		loweredWord := strings.ToLower(word)
		if _, ok := profaneWords[loweredWord]; ok {
			words[index] = "****"
		}
	}
	return strings.Join(words, " ")
}
