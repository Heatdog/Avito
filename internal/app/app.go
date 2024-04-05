package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/Heatdog/Avito/docs"

	"github.com/Heatdog/Avito/internal/config"
	"github.com/Heatdog/Avito/internal/migrations"
	banner_postgre "github.com/Heatdog/Avito/internal/repository/banner/postgre"
	banner_service "github.com/Heatdog/Avito/internal/service/banner"
	banners_transport "github.com/Heatdog/Avito/internal/transport/banners"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	"github.com/Heatdog/Avito/pkg/client/postgre"
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
// @name Authorization
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
		panic(err)
	}

	router := mux.NewRouter()

	logger.Debug("register middlewre")
	middleware := middleware_transport.NewMiddleware(logger)

	logger.Debug("register banners handler")
	bannerRepo := banner_postgre.NewBannerRepository(logger, dbClient)
	bannerService := banner_service.NewBannerService(logger, bannerRepo)
	bannerHandler := banners_transport.NewBunnersHandler(logger, bannerService, middleware)
	bannerHandler.Register(router)

	logger.Info("adding swagger documentation")
	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	)).Methods(http.MethodGet)

	host := fmt.Sprintf("%s:%s", cfg.Server.IP, cfg.Server.Port)
	logger.Info("listen tcp", slog.String("host", host))

	if err := http.ListenAndServe(host, router); err != nil {
		logger.Error(err.Error())
		panic(err)
	}
}
