package banner_postgre

import (
	"context"
	"fmt"
	"log/slog"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/query_params"
	banner_repository "github.com/Heatdog/Avito/internal/repository/banner"
	"github.com/Heatdog/Avito/pkg/client"
	"github.com/jackc/pgx/v5"
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

func (repo *bannerRepository) InsertBanner(ctx context.Context, banner banner_model.BannerInsert) (int, error) {
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

	for _, tag := range banner.TagsID {
		if err = repo.insertCrossTable(ctx, transaction, tag, banner.FeatureID, id); err != nil {
			repo.logger.Warn(err.Error())
			return 0, err
		}
	}
	return id, nil
}

func (repo *bannerRepository) insertInBannerTable(ctx context.Context, transaction pgx.Tx,
	banner banner_model.BannerInsert) (int, error) {
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

func (repo *bannerRepository) insertCrossTable(ctx context.Context, transaction pgx.Tx, tagId, feauterId,
	bannerId int) error {
	repo.logger.Debug("insert into features_tags_to_banners table", slog.Int("tagId", tagId),
		slog.Int("bannerId", bannerId), slog.Int("feauterId", feauterId))

	q := `
		INSERT INTO features_tags_to_banners (feature_id, tag_id, banner_id)
		VALUES ($1, $2, $3)
	`
	repo.logger.Debug("repo query", slog.String("query", q))

	tag, err := transaction.Exec(ctx, q, feauterId, tagId, bannerId)
	if err != nil {
		return err
	}

	if tag.RowsAffected() != 1 {
		return fmt.Errorf("rows affected error")
	}

	return nil
}

func (repo *bannerRepository) GetUserBanner(ctx context.Context, tagID, feautureID int) (banner_model.Banner, error) {
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

func (repo *bannerRepository) GetBanners(ctx context.Context, params query_params.BannerParams) ([]banner_model.Banner,
	error) {

	repo.logger.Debug("get banners repository")
	banners, err := repo.getOnlyBanners(ctx, params)
	if err != nil {
		repo.logger.Warn(err.Error())
		return nil, err
	}

	for i, banner := range banners {
		bannerParams, err := repo.GetBannerParams(ctx, banner.ID)
		if err != nil {
			repo.logger.Warn(err.Error())
			return nil, err
		}
		banners[i].FeatureID = bannerParams.FeatureID
		banners[i].TagsID = bannerParams.TagIDs
	}

	return banners, nil
}

func (repo *bannerRepository) getOnlyBanners(ctx context.Context, params query_params.BannerParams) ([]banner_model.Banner,
	error) {

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

	var banners []banner_model.Banner
	for rows.Next() {

		var banner banner_model.Banner
		if err = rows.Scan(&banner.ID, &banner.Content, &banner.IsActive, &banner.CreatedAt,
			&banner.UpdatedAt); err != nil {

			return nil, err
		}

		banners = append(banners, banner)
	}
	return banners, nil
}

func (repo *bannerRepository) makeQueryBanner(params query_params.BannerParams) string {
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
	q += " ORDER BY b.updated_at"

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