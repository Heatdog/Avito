package banner_postgre

import (
	"context"
	"log/slog"

	"github.com/Heatdog/Avito/internal/models/query_params"
	"github.com/jackc/pgx/v5"
)

func (repo *bannerRepository) DeleteBanner(ctx context.Context, id int) (bool, error) {
	q := `
		DELETE FROM banners
		WHERE id = $1
	`
	repo.logger.Debug("repo query", slog.String("query", q))

	tag, err := repo.dbClient.Exec(ctx, q, id)
	if err != nil {
		repo.logger.Warn(err.Error())
		return false, err
	}

	if tag.RowsAffected() < 1 {
		return false, nil
	}
	return true, nil
}

func (repo *bannerRepository) DeleteBanners(ctx context.Context, params *query_params.DeleteBannerParams) error {
	tx, err := repo.dbClient.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		repo.logger.Warn("tx begin error", tx)
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

	bannersID, err := repo.getBannersID(ctx, params)
	if err != nil {
		return err
	}

	for _, id := range bannersID {
		if _, err := repo.DeleteBanner(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (repo *bannerRepository) getBannersID(ctx context.Context, params *query_params.DeleteBannerParams) ([]int, error) {
	q := `
		SELECT banner_id
		FROM features_tags_to_banners
		WHERE
	`
	var res []int
	var rows pgx.Rows
	var err error
	if params.FeatureID != nil && params.TagID != nil {
		q += " tag_id = $1 OR feature_id = $2"
		rows, err = repo.dbClient.Query(ctx, q, params.TagID, params.FeatureID)
	} else if params.FeatureID != nil {
		q += " feature_id = $1"
		rows, err = repo.dbClient.Query(ctx, q, params.FeatureID)
	} else if params.TagID != nil {
		q += " tag_id = $1"
		rows, err = repo.dbClient.Query(ctx, q, params.TagID)
	} else {
		return res, nil
	}

	if err != nil {
		return nil, err
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
