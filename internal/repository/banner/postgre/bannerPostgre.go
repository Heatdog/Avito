package banner_postgre

import (
	"context"
	"fmt"
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

	id, err := repo.insertInBannerTable(ctx, banner)
	if err != nil {
		repo.logger.Warn(err.Error())

		if errTransaction := transaction.Rollback(ctx); errTransaction != nil {
			repo.logger.Warn(err.Error())
			return 0, errTransaction
		}
		return 0, err
	}

	for _, tag := range banner.TagsId {
		if err = repo.insertCrossTable(ctx, tag, banner.FeatureId, id); err != nil {
			repo.logger.Warn(err.Error())

			if errTransaction := transaction.Rollback(ctx); errTransaction != nil {
				repo.logger.Warn(err.Error())
				return 0, errTransaction
			}
			return 0, err
		}
	}

	if err = transaction.Commit(ctx); err != nil {
		repo.logger.Warn(err.Error())
		return 0, err
	}
	return id, nil
}

func (repo *bannerRepository) insertInBannerTable(ctx context.Context, banner banner_model.BannerInsert) (int, error) {
	repo.logger.Debug("insert into banners", slog.Any("banner", banner))

	q := `
		INSERT INTO banners (content, is_active)
		VALUES ($1, $2)
		RETURNING id
	`

	repo.logger.Debug("repo query", slog.String("query", q))
	row := repo.dbClient.QueryRow(ctx, q, banner.Content, banner.IsActive)

	var id int

	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *bannerRepository) insertCrossTable(ctx context.Context, tagId, feauterId, bannerId int) error {
	repo.logger.Debug("insert into features_tags_to_banners table", slog.Int("tagId", tagId),
		slog.Int("bannerId", bannerId), slog.Int("feauterId", feauterId))

	q := `
		INSERT INTO features_tags_to_banners (feature_id, tag_id, banner_id)
		VALUES ($1, $2, $3)
	`
	repo.logger.Debug("repo query", slog.String("query", q))

	tag, err := repo.dbClient.Exec(ctx, q, bannerId, tagId)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("rows affected error")
	}

	return nil
}
