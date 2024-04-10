package banner_postgre

import (
	"context"
	"fmt"
	"log/slog"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (repo *bannerRepository) UpdateBanner(ctx context.Context, banner *banner_model.BannerUpdate) error {
	repo.logger.Debug("update banner", slog.Int("id", banner.ID))

	tx, err := repo.dbClient.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			repo.logger.Warn(err.Error())
			tx.Rollback(ctx)
		} else {
			repo.logger.Debug("commit")
			tx.Commit(ctx)
		}
	}()

	err = repo.updateOnlyBanner(ctx, tx, banner)
	if err != nil {
		return err
	}

	if banner.FeatureID != nil || banner.TagsID != nil {
		params, err := repo.GetBannerParams(ctx, banner.ID)
		if err != nil {
			return err
		}

		if err = repo.deleteCrossTable(ctx, tx, banner.ID); err != nil {
			return err
		}

		var feauterId int
		var tagsIDs []int
		if banner.FeatureID != nil {
			feauterId = *banner.FeatureID
		} else {
			feauterId = params.FeatureID
		}

		if banner.TagsID != nil {
			tagsIDs = *banner.TagsID
		} else {
			tagsIDs = params.TagIDs
		}

		if err = repo.insertCrossTable(ctx, tx, feauterId, banner.ID, tagsIDs); err != nil {
			return err
		}

	}

	return nil
}

func (repo *bannerRepository) updateOnlyBanner(ctx context.Context, tx pgx.Tx, banner *banner_model.BannerUpdate) error {
	q := `
		UPDATE banners
	`
	var tag pgconn.CommandTag
	var err error
	if banner.Content != nil {

		q += " SET content_v1 = $1, content_v2 = content_v1, content_v3 = content_v2"
		if banner.IsActive != nil {
			q += ", is_active = $2, updated_at = now() WHERE id = $3"
			repo.logger.Debug(q)

			tag, err = tx.Exec(ctx, q, banner.Content, *banner.IsActive, banner.ID)
		} else {
			q += ", updated_at = now() WHERE id = $2"
			repo.logger.Debug(q)

			tag, err = tx.Exec(ctx, q, banner.Content, banner.ID)
		}

	} else if banner.IsActive != nil {
		q += " SET is_active = $1, updated_at = now() WHERE id = $2"
		repo.logger.Debug(q)

		tag, err = tx.Exec(ctx, q, *banner.IsActive, banner.ID)
	} else {
		q += " SET updated_at = now() WHERE id = $1"
		repo.logger.Debug(q)

		tag, err = tx.Exec(ctx, q, banner.ID)
	}

	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return nil
}

func (repo *bannerRepository) deleteCrossTable(ctx context.Context, tx pgx.Tx, bannerID int) error {
	q := `
		DELETE FROM features_tags_to_banners
		WHERE banner_id = $1
	`
	repo.logger.Debug("repo query", slog.String("query", q))

	if _, err := tx.Exec(ctx, q, bannerID); err != nil {
		repo.logger.Warn(err.Error())
		return err
	}
	return nil
}

func (repo *bannerRepository) UpdateBannerVersion(ctx context.Context, id, version int) error {
	repo.logger.Debug("update banner version", slog.Int("id", id), slog.Int("version", version))

	q := fmt.Sprintf(`
		UPDATE banners
		SET content_v1 = content_v%d, content_v2 = content_v1, content_v3 = content_v2, updated_at = now()
		WHERE id = $1
	`, version)
	repo.logger.Debug(q)

	tag, err := repo.dbClient.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}

	return nil
}
