package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nacen-dev/chirpy/internal/auth"
	"github.com/nacen-dev/chirpy/internal/database"
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
		Email    string `json:"email"`
		Password string `json:"password"`
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

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to hash the password", nil)
		return
	}

	user, err := cfg.db.CreateUser(req.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	})

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

func (cfg *apiConfig) handleLogin(res http.ResponseWriter, req *http.Request) {
	type userLogin struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	type response struct {
		User
		Token        string `json:"token"`
		RefreshToken string `json:"refresh_token"`
	}
	decoder := json.NewDecoder(req.Body)
	params := userLogin{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(req.Context(), params.Email)
	checkPasswordHashErr := auth.CheckPasswordHash(params.Password, user.HashedPassword)

	if err != nil || checkPasswordHashErr != nil {
		respondWithError(res, http.StatusUnauthorized, "Incorrect email or password", err)
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, time.Duration(auth.DefaultExpirationInHours)*time.Hour)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to get token", err)
		return
	}

	refreshToken, _ := auth.MakeRefreshToken()
	cfg.db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
		RevokedAt: sql.NullTime{},
	})

	respondWithJSON(res, http.StatusOK, response{
		Token: accessToken,
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handleUpdateUser(res http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "Malformed or missing token", err)
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	userId, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "invalid token", err)
		return
	}

	user, err := cfg.db.GetUserById(req.Context(), userId)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to retrieve the user", err)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to hash the password", err)
		return
	}

	updatedUserData, err := cfg.db.UpdateUser(req.Context(), database.UpdateUserParams{
		NewEmail:    params.Email,
		NewPassword: hashedPassword,
		OldEmail:    user.Email,
	})
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to update the user's data", err)
		return
	}

	respondWithJSON(res, http.StatusOK, User{
		ID:        updatedUserData.ID,
		CreatedAt: updatedUserData.CreatedAt,
		UpdatedAt: updatedUserData.UpdatedAt,
		Email:     updatedUserData.Email,
	})
}
