package downstream

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	gormopentracing "gorm.io/plugin/opentracing"

	"opentracing-playground/database"
	"opentracing-playground/logging"
	"opentracing-playground/models/user"
	"opentracing-playground/pkg/jaeger"
)

// Server is a HTTP server.
type Server struct {
}

// RegisterMiddleware registers middleware for all endpoints.
func (s *Server) RegisterMiddleware(r *gin.Engine) {
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	config.AllowAllOrigins = true

	r.Use(cors.New(config))
}

// RegisterEndpoint installs api representation layer processing function.
func (s *Server) RegisterEndpoint(r *gin.Engine) {
	r.GET("ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, "pong")
	})
	r.POST("user", func(ctx *gin.Context) {
		var req struct {
			Email    string `json:"email" binding:"required"`
			Password string `json:"password" binding:"required"`
			Name     string `json:"name" binding:"required"`
		}

		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}

		bs, err := json.Marshal(req)
		if err != nil {
			logging.Get().Error(errors.Wrap(err, "json marshal failed"))
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		httpReq, err := http.NewRequest(http.MethodPost,
			"http://upstream:8787/user",
			bytes.NewBuffer(bs))
		if err != nil {
			logging.Get().Error(errors.Wrap(err, "http gen request failed"))
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		var client http.Client
		resp, err := client.Do(httpReq)
		if err != nil {
			logging.Get().Error(errors.Wrap(err, "proxy to upstream failed"))
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			logging.Get().Errorf("invalid status: %v %v: %v",
				resp.StatusCode, http.StatusText(resp.StatusCode), string(body))
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		var newUser user.User
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&newUser); err != nil {
			logging.Get().Error(errors.Wrap(err, "json decode failed"))
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		db := database.GetDB(database.Default)
		// if err := db.Use(gormopentracing.New()); err != nil {
		// 	logging.Get().Error(err)
		// 	ctx.JSON(http.StatusInternalServerError, nil)
		// 	return
		// }

		if err := db.Create(&newUser).Error; err != nil {
			logging.Get().Error(err)
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.JSON(http.StatusOK, nil)
	})
}

// Start starts HTTP server.
func (s *Server) Start(ctx context.Context, apiAddr string) {
	if err := jaeger.InitGlobalTracer("downstream"); err != nil {
		log.Fatal("jaeger.InitGlobalTracer: ", err)
	}

	db := database.GetDB(database.Default)
	if err := db.Use(gormopentracing.New()); err != nil {
		log.Fatal("use gorm opentracing: ", err)
	}

	gin.ForceConsoleColor()

	// setup gin.
	apiEngine := gin.New()
	apiEngine.RedirectTrailingSlash = true

	s.RegisterMiddleware(apiEngine)

	// setup endpoint.
	s.RegisterEndpoint(apiEngine)

	srv := &http.Server{
		Addr:    apiAddr,
		Handler: apiEngine,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server Shutdown: ", err)
		}
	}()

	logging.Get().Info("starts serving...")
	if err := srv.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("listen: %s\n", err)
	}
}
