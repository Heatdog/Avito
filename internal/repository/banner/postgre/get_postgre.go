package banner_postgre

import (
	"context"
	"fmt"
	"log/slog"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/query_params"
	"github.com/jackc/pgx/v5"
)

func (repo *bannerRepository) GetUserBanner(ctx context.Context, tagID, feautureID string) (banner_model.Banner, error) {
	repo.logger.Debug("get user banner repository")

	q := `
		SELECT b.id, b.content, b.is_active
		FROM banners b
		JOIN features_tags_to_banners ftb ON ftb.banner_id = b.id
		WHERE ftb.feature_id = $1 AND ftb.tag_id = $2
	`
	repo.logger.Debug("repo query", slog.String("query", q))
	row := repo.dbClient.QueryRow(ctx, q, feautureID, tagID)

	var banner banner_model.Banner
	if err := row.Scan(&banner.ID, &banner.Content, &banner.IsActive); err != nil {
		repo.logger.Warn(err.Error())
		return banner_model.Banner{}, err
	}

	return banner, nil
}

func (repo *bannerRepository) GetBanners(ctx context.Context, params *query_params.BannerParams) ([]banner_model.Banner,
	error) {

	repo.logger.Debug("get banners repository")
	banners, err := repo.getOnlyBanners(ctx, params)
	if err != nil {
		repo.logger.Warn(err.Error())
		return nil, err
	}

	var res []banner_model.Banner

	for _, banner := range banners {
		bannerParams, err := repo.GetBannerParams(ctx, banner.ID)
		if err != nil {
			repo.logger.Warn(err.Error())
			return nil, err
		}
		banner = banners[banner.ID]
		banner.FeatureID = bannerParams.FeatureID
		banner.TagsID = bannerParams.TagIDs
		res = append(res, banner)
	}

	return res, nil
}

func (repo *bannerRepository) getOnlyBanners(ctx context.Context, params *query_params.BannerParams) (
	map[int]banner_model.Banner, error) {

	var rows pgx.Rows
	var err error
	q := repo.makeQueryBanner(params)
	if params.FeatureID != nil {
		if params.TagID != nil {
			rows, err = repo.dbClient.Query(ctx, q, &params.FeatureID, &params.TagID)
		} else {
			rows, err = repo.dbClient.Query(ctx, q, &params.FeatureID)
		}
	} else {
		if params.TagID != nil {
			rows, err = repo.dbClient.Query(ctx, q, &params.TagID)
		} else {
			rows, err = repo.dbClient.Query(ctx, q)
		}
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	banners := make(map[int]banner_model.Banner)
	for rows.Next() {

		var banner banner_model.Banner
		if err = rows.Scan(&banner.ID, &banner.Content, &banner.IsActive, &banner.CreatedAt,
			&banner.UpdatedAt); err != nil {

			return nil, err
		}
		banners[banner.ID] = banner
	}
	return banners, nil
}

func (repo *bannerRepository) makeQueryBanner(params *query_params.BannerParams) string {
	q := `
		SELECT b.id, b.content, b.is_active, b.created_at, b.updated_at
		FROM banners b
	`
	if params.FeatureID != nil {
		q += "JOIN features_tags_to_banners ftb ON ftb.feature_id = $1 AND ftb.banner_id = b.id"
		if params.TagID != nil {
			q += " AND ftb.tag_id = $2"
		}
	} else {
		if params.TagID != nil {
			q += "JOIN features_tags_to_banners ftb ON ftb.tag_id = $1 AND ftb.banner_id = b.id"
		}
	}
	q += " ORDER BY b.updated_at DESC"

	if params.Limit != nil {
		q += fmt.Sprintf(` LIMIT %d`, *params.Limit)
	}
	if params.Offset != nil {
		q += fmt.Sprintf(` OFFSET %d`, *params.Offset)
	}

	repo.logger.Debug("repo query", slog.String("query", q))
	return q
}

func (repo *bannerRepository) GetBannerParams(ctx context.Context, bannerID int) (banner_model.BannerParams,
	error) {
	q := `
		SELECT feature_id, tag_id
		FROM features_tags_to_banners
		WHERE banner_id = $1
	`
	repo.logger.Debug("repo query", slog.String("query", q))
	rows, err := repo.dbClient.Query(ctx, q, bannerID)
	if err != nil {
		return banner_model.BannerParams{}, err
	}
	defer rows.Close()

	var banerKeys banner_model.BannerParams

	for rows.Next() {
		var tagID int
		if err = rows.Scan(&banerKeys.FeatureID, &tagID); err != nil {
			return banner_model.BannerParams{}, err
		}

		banerKeys.TagIDs = append(banerKeys.TagIDs, tagID)
	}

	return banerKeys, nil
}
