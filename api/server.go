package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/lushenle/chatgpt-web/db/sqlc"
	"github.com/lushenle/chatgpt-web/token"
	"github.com/lushenle/chatgpt-web/util"
)

// Server servers HTTP requests for chatgpt service
type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

// NewServer creates a new HTTP sever and setup routing
func NewServer(config util.Config, store db.Store) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.ChatGPT.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
	}

	server.setupRouter()
	return server, nil
}

func (server *Server) setupRouter() {
	router := gin.Default()

	router.Use(corsMiddleware())
	router.POST("/register", server.createUser)
	router.POST("/login", server.loginUser)
	router.POST("/tokens/renew_access", server.renewAccessToken)

	authRoutes := router.Group("/").Use(authMiddleware(server.tokenMaker))

	authRoutes.GET("/", server.index)
	authRoutes.POST("/completion", server.completion)

	server.router = router

}

// Start runs the HTTP server on a specific  address
func (server *Server) Start(address string) error {
	router := server.router
	// initializes the loading path of the HTML template
	router.LoadHTMLGlob("resources/view/*")
	// initialize static file
	router.StaticFS("/assets", http.Dir("static/assets"))
	router.StaticFile("favicon.ico", "static/favicon.ico")

	server.router = router
	return server.router.Run(address)
}

func (server *Server) responseJson(ctx *gin.Context, code int, errorMsg string, data interface{}) {
	ctx.JSON(code, gin.H{
		"code":     code,
		"errorMsg": errorMsg,
		"data":     data,
	})
	ctx.Abort()
}

func (server *Server) index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"title": "main",
	})
}
