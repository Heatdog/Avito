package bannerpostgre

import (
	"context"
	"log/slog"

	"github.com/Heatdog/Avito/internal/models/queryparams"
	"github.com/jackc/pgx/v5"
)

func (repo *bannerRepository) DeleteBanner(ctx context.Context, id int) (bool, error) {
	tx, err := repo.dbClient.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		repo.logger.Warn(err.Error())
		return false, err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			repo.logger.Warn(err.Error())
		}
	}()

	res, err := repo.deleteBanner(ctx, tx, id)
	if err != nil {
		repo.logger.Warn(err.Error())
		return false, err
	}

	if err = tx.Commit(ctx); err != nil {
		repo.logger.Warn(err.Error())
		return false, err
	}

	return res, nil
}

func (repo *bannerRepository) DeleteBanners(ctx context.Context, params queryparams.DeleteBannerParams) error {
	tx, err := repo.dbClient.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		repo.logger.Warn(err.Error())
		return err
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			repo.logger.Warn(err.Error())
		}
	}()

	bannersID, err := repo.getBannersID(ctx, tx, &params)
	if err != nil {
		repo.logger.Warn(err.Error())
		return err
	}

	for _, id := range bannersID {
		if _, err := repo.deleteBanner(ctx, tx, id); err != nil {
			repo.logger.Warn(err.Error())
			return err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		repo.logger.Warn(err.Error())
		return err
	}

	return nil
}

func (repo *bannerRepository) getBannersID(ctx context.Context, tx pgx.Tx,
	params *queryparams.DeleteBannerParams) ([]int, error) {
	var (
		res  []int
		rows pgx.Rows
		err  error
	)

	if params.FeatureID != nil && params.TagID != nil {
		q := `
			SELECT banner_id
			FROM features_tags_to_banners
			WHERE tag_id = $1 OR feature_id = $2
			`

		rows, err = tx.Query(ctx, q, params.TagID, params.FeatureID)

		if err != nil {
			return nil, err
		}
	} else if params.FeatureID != nil {
		q := `
			SELECT banner_id
			FROM features_tags_to_banners
			WHERE feature_id = $1
			`

		rows, err = tx.Query(ctx, q, params.FeatureID)

		if err != nil {
			return nil, err
		}
	} else if params.TagID != nil {
		q := `
			SELECT banner_id
			FROM features_tags_to_banners
			WHERE tag_id = $1
			`

		rows, err = tx.Query(ctx, q, params.TagID)

		if err != nil {
			return nil, err
		}
	} else {
		return res, nil
	}

	for rows.Next() {
		var id int

		if err = rows.Scan(&id); err != nil {
			return res, err
		}

		res = append(res, id)
	}

	return res, nil
}

func (repo *bannerRepository) deleteBanner(ctx context.Context, tx pgx.Tx, id int) (bool, error) {
	q := `
		DELETE FROM banners
		WHERE id = $1
	`
	repo.logger.Debug("repo query", slog.String("query", q))

	tag, err := tx.Exec(ctx, q, id)
	if err != nil {
		repo.logger.Warn(err.Error())
		return false, err
	}

	if tag.RowsAffected() < 1 {
		return false, nil
	}

	return true, nil
}
