package banner_postgre

import (
	"context"
	"fmt"
	"log/slog"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/jackc/pgx/v5"
)

func (repo *bannerRepository) InsertBanner(ctx context.Context, banner *banner_model.BannerInsert) (int, error) {
	repo.logger.Debug("insert banner repository")

	repo.logger.Debug("begin transaction")
	transaction, err := repo.dbClient.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		repo.logger.Error(err.Error())
		return 0, err
	}
	defer func() {
		if err != nil {
			repo.logger.Debug("rollback")
			transaction.Rollback(ctx)
		} else {
			repo.logger.Debug("commit")
			transaction.Commit(ctx)
		}
	}()

	repo.logger.Debug("insert banner", slog.Any("banner", banner))

	id, err := repo.insertInBannerTable(ctx, transaction, banner)
	if err != nil {
		repo.logger.Warn(err.Error())
		return 0, err
	}

	if err = repo.insertCrossTable(ctx, transaction, banner.FeatureID, id, banner.TagsID); err != nil {
		repo.logger.Warn(err.Error())
		return 0, err
	}
	return id, nil
}

func (repo *bannerRepository) insertInBannerTable(ctx context.Context, transaction pgx.Tx,
	banner *banner_model.BannerInsert) (int, error) {
	repo.logger.Debug("insert into banners", slog.Any("banner", banner))

	q := `
		INSERT INTO banners (content, is_active)
		VALUES ($1, $2)
		RETURNING id
	`

	repo.logger.Debug("repo query", slog.String("query", q))
	row := transaction.QueryRow(ctx, q, banner.Content, banner.IsActive)

	var id int

	if err := row.Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

func (repo *bannerRepository) insertCrossTable(ctx context.Context, transaction pgx.Tx, featureID, bannerId int,
	tagIDs []int) error {
	repo.logger.Debug("insert into features_tags_to_banners table", slog.Any("tagIds", tagIDs),
		slog.Int("bannerId", bannerId), slog.Int("feauterIds", featureID))

	q := `
		INSERT INTO features_tags_to_banners (feature_id, tag_id, banner_id)
		VALUES ($1, $2, $3)
	`
	repo.logger.Debug("repo query", slog.String("query", q))

	for _, tagId := range tagIDs {
		tag, err := transaction.Exec(ctx, q, featureID, tagId, bannerId)
		if err != nil {
			return err
		}

		if tag.RowsAffected() != 1 {
			return fmt.Errorf("rows affected error")
		}
	}

	return nil
}
