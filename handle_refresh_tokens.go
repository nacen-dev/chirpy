package main

import (
	"net/http"
	"time"

	"github.com/nacen-dev/chirpy/internal/auth"
)

func (cfg *apiConfig) handleRefresh(res http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)

	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "Missing token", err)
		return
	}

	refreshTokenData, err := cfg.db.GetUserFromRefreshToken(req.Context(), token)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "Unable to get user from the refresh token", err)
		return
	}

	if refreshTokenData.RevokedAt.Valid || time.Now().After(refreshTokenData.ExpiresAt) {
		respondWithError(res, http.StatusUnauthorized, "Token is expired or revoked", err)
		return
	}

	accessToken, err := auth.MakeJWT(refreshTokenData.UserID, cfg.jwtSecret, time.Duration(auth.DefaultExpirationInHours)*time.Hour)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "Unable to create token", err)
		return
	}
	type response struct {
		Token string `json:"token"`
	}
	respondWithJSON(res, http.StatusOK, response{
		Token: accessToken,
	})

}

func (cfg *apiConfig) handleRevoke(res http.ResponseWriter, req *http.Request) {
	token, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(res, http.StatusUnauthorized, "No token found", err)
		return
	}
	err = cfg.db.RevokeRefreshToken(req.Context(), token)
	if err != nil {
		respondWithError(res, http.StatusInternalServerError, "unable to revoke the token", err)
	}
	res.WriteHeader(http.StatusNoContent)
}
