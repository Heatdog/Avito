package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/Heatdog/Avito/docs"
	"github.com/hashicorp/golang-lru/v2/expirable"

	"github.com/Heatdog/Avito/internal/config"
	"github.com/Heatdog/Avito/internal/migrations"
	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	banner_postgre "github.com/Heatdog/Avito/internal/repository/banner/postgre"
	banner_service "github.com/Heatdog/Avito/internal/service/banner"
	banners_transport "github.com/Heatdog/Avito/internal/transport/banners"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	hashicorp_lru "github.com/Heatdog/Avito/pkg/cache/hashi_corp"
	"github.com/Heatdog/Avito/pkg/client/postgre"
	"github.com/Heatdog/Avito/pkg/token/simple_token"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// swag init --pd -g internal/app/app.go

// @title Сервис баннеров
// @description API сервер для сервиса баннеров

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apiKey ApiKeyAuth
// @in header
// @name token
func App() {
	opt := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opt))
	slog.SetDefault(logger)

	ctx := context.Background()

	logger.Info("reading server config files")
	cfg := config.NewConfigStorage(logger)

	logger.Info("connecting to DataBase")
	dbClient, err := postgre.NewPostgreClient(ctx, cfg.Postgre)
	if err != nil {
		logger.Error("connection to PostgreSQL failed", slog.Any("error", err))
		panic(err)
	}
	defer dbClient.Close()

	logger.Info("init db")
	if err = migrations.InitDb(dbClient); err != nil {
		logger.Error(err.Error())
	}

	logger.Info("init cache")
	cacheLRU := expirable.NewLRU[banner_model.BannerKey, *banner_model.Banner](cfg.Cache.Size, nil,
		time.Minute*time.Duration(cfg.Cache.TTL))
	cache := hashicorp_lru.NewLRU(logger, cacheLRU)

	tokenProvider := simple_token.NewSimpleTokenProvider()

	logger.Debug("register middlewre")
	middleware := middleware_transport.NewMiddleware(logger, tokenProvider)

	router := mux.NewRouter()
	router.Use(middleware.Logging)

	logger.Debug("register banners handler")
	bannerRepo := banner_postgre.NewBannerRepository(logger, dbClient)
	bannerService := banner_service.NewBannerService(logger, bannerRepo, cache, tokenProvider)
	bannerHandler := banners_transport.NewBannersHandler(logger, bannerService, middleware)
	bannerHandler.Register(router)

	logger.Info("adding swagger documentation")
	host := fmt.Sprintf("%s:%d", cfg.Server.IP, cfg.Server.Port)
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%d/swagger/doc.json", cfg.Server.Port)),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	logger.Info("listen tcp", slog.String("host", host))

	if err := http.ListenAndServe(host, router); err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}
