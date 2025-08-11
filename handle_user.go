package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handleResetUsers(res http.ResponseWriter, req *http.Request) {
	if cfg.platform != "dev" {
		res.WriteHeader(http.StatusForbidden)
		res.Write([]byte("Reset is only allowed in dev environment."))
		return
	}
	err := cfg.db.ResetUsers(req.Context())
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to reset the users", err)
		return
	}
	res.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handleCreateUsers(res http.ResponseWriter, req *http.Request) {
	type userRegistrationRequest struct {
		Email string `json:"email"`
	}
	type userRegistrationResponse struct {
		User
	}

	decoder := json.NewDecoder(req.Body)
	params := userRegistrationRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}
	user, err := cfg.db.CreateUser(req.Context(), params.Email)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't create user", err)
		return
	}
	respondWithJSON(res, http.StatusCreated, userRegistrationResponse{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}
