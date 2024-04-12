package bannerpostgre

import (
	"log/slog"

	banner_repository "github.com/Heatdog/Avito/internal/repository/banner"
	"github.com/Heatdog/Avito/pkg/client"
)

type bannerRepository struct {
	logger   *slog.Logger
	dbClient client.Client
}

func NewBannerRepository(logger *slog.Logger, dbClient client.Client) banner_repository.BannerRepository {
	return &bannerRepository{
		logger:   logger,
		dbClient: dbClient,
	}
}
