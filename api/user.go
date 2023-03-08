package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	db "github.com/lushenle/chatgpt-web/db/sqlc"
	"github.com/lushenle/chatgpt-web/util"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"passwordChangedAt"`
	CreatedAt         time.Time `json:"createdAt"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.responseJson(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.store.CreateUser(ctx, arg)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			case "unique_violation":
				server.responseJson(ctx, http.StatusForbidden, err.Error(), nil)
				return
			}
		}
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	rsp := newUserResponse(user)
	server.responseJson(ctx, http.StatusOK, "", rsp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		server.responseJson(ctx, http.StatusBadRequest, err.Error(), nil)
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			server.responseJson(ctx, http.StatusNotFound, err.Error(), nil)
			return
		}
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		server.responseJson(ctx, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.ChatGPT.AccessTokenDuration,
	)
	if err != nil {
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.ChatGPT.RefreshTokenDuration,
	)
	if err != nil {
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		server.responseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	rsp := loginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  newUserResponse(user),
	}
	server.responseJson(ctx, http.StatusOK, "", rsp)
}
