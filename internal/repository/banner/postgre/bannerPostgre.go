package banner_postgre

import (
	"context"
	"log/slog"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	banner_repository "github.com/Heatdog/Avito/internal/repository/banner"
	"github.com/Heatdog/Avito/pkg/client/postgre"
)

type bannerRepository struct {
	logger   *slog.Logger
	dbClient postgre.Client
}

func NewBannerRepository(logger *slog.Logger, dbClient postgre.Client) banner_repository.BannerRepository {
	return &bannerRepository{
		logger:   logger,
		dbClient: dbClient,
	}
}

func (repo *bannerRepository) InsertBanner(ctx context.Context, banner banner_model.BannerInsert) (int, error) {
	repo.logger.Debug("insert banner repository")

	repo.logger.Debug("begin transaction")
	transaction, err := repo.dbClient.Begin(ctx)
	if err != nil {
		repo.logger.Error(err.Error())
		return 0, err
	}

	repo.logger.Debug("insert banner", slog.Any("banner", banner))

	for _, tag := range banner.TagsId {

	}
}

func (repo *bannerRepository) insertBannersTags(ctx context.Context, tagId, bannerId int) error {
	repo.logger.Debug("insert into banners_to_tags", slog.Int("tagId", tagId), slog.Int("bannerId", bannerId))

	q := `
		INSERT INTO banners_to_tags (banner_id, tag_id)
		VALUES ($1, $2)
	`
	repo.logger.Debug("repo query", slog.String("query", q))

	tag, err := repo.dbClient.Exec(ctx, q, bannerId, tagId)
	if err != nil {
		repo.logger.Warn(err.Error())
		return err
	}

}
