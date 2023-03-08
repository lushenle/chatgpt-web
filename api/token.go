package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (server *Server) renewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.responseJson(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	refreshPayload, err := server.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		server.responseJson(ctx, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	session, err := server.store.GetSession(ctx, refreshPayload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			server.responseJson(ctx, http.StatusNotFound, err.Error(), nil)
			return
		}
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		server.responseJson(ctx, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	if session.Username != refreshPayload.Username {
		err := fmt.Errorf("incorrect session user")
		server.responseJson(ctx, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		server.responseJson(ctx, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		server.responseJson(ctx, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		refreshPayload.Username,
		server.config.ChatGPT.AccessTokenDuration,
	)
	if err != nil {
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}
	ctx.JSON(http.StatusOK, rsp)
	server.responseJson(ctx, http.StatusOK, "", rsp)
}
