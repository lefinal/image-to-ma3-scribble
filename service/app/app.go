package app

import (
	"context"
	"errors"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/lefinal/image-to-ma3-scribble/web"
	"github.com/lefinal/meh"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"net/http"
	"time"
)

type Config struct {
	Logger            *zap.Logger
	HTTPAPIListenAddr string
	PotraceFilename   string
}

type App struct {
	logger     *zap.Logger
	config     Config
	httpClient *http.Client
}

func New(config Config) *App {
	return &App{
		logger:     config.Logger,
		config:     config,
		httpClient: &http.Client{},
	}
}

func (app *App) Run(ctx context.Context) error {
	app.logger.Info("startup")
	defer app.logger.Info("shutdown")

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	apiLogger := app.logger.Named("http")
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodPost, http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
	}))
	r.Use(web.RequestDebugLogger(apiLogger))
	builder := web.HandlerBuilder{Logger: apiLogger}
	r.POST("/api/v1/png-to-ma3-scribble/preview", builder.GinHandler(app.handlePNGToMA3ScribblePreview()))
	r.POST("/api/v1/png-to-ma3-scribble", builder.GinHandler(app.handlePNGToMA3Scribble()))

	httpServer := http.Server{
		Addr:           app.config.HTTPAPIListenAddr,
		Handler:        r,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		app.logger.Info("serve http", zap.String("listen_addr", app.config.HTTPAPIListenAddr))
		err := httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return meh.NewInternalErrFromErr(err, "serve http", nil)
		}
		return nil
	})
	eg.Go(func() error {
		<-egCtx.Done()
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err := httpServer.Shutdown(timeout)
		if err != nil {
			return meh.NewInternalErrFromErr(err, "shutdown http server", nil)
		}
		return nil
	})

	return eg.Wait()
}
