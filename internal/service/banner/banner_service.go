package banner_service

import (
	"context"
	"log/slog"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/query_params"
	banner_repository "github.com/Heatdog/Avito/internal/repository/banner"
	"github.com/Heatdog/Avito/pkg/cache"
	"github.com/Heatdog/Avito/pkg/token"
	"github.com/jackc/pgx/v5"
)

type BannerService interface {
	InsertBanner(context context.Context, banner *banner_model.BannerInsert) (int, error)
	GetUserBanner(context context.Context, params *query_params.BannerUserParams) (interface{}, error)
	GetBanners(context context.Context, params *query_params.BannerParams) ([]banner_model.Banner, error)
	DeleteBanner(context context.Context, id int) (bool, error)
	UpdateBanner(context context.Context, banner *banner_model.BannerUpdate) error
	DeleteBanners(context context.Context, params *query_params.DeleteBannerParams)
}

type bannerService struct {
	logger        *slog.Logger
	repo          banner_repository.BannerRepository
	cache         cache.Cache[banner_model.BannerKey, *banner_model.Banner]
	tokenProvider token.TokenProvider
}

func NewBannerService(logger *slog.Logger, repo banner_repository.BannerRepository,
	cache cache.Cache[banner_model.BannerKey, *banner_model.Banner], tokenProvider token.TokenProvider) BannerService {
	return &bannerService{
		logger:        logger,
		repo:          repo,
		cache:         cache,
		tokenProvider: tokenProvider,
	}
}

func (service *bannerService) InsertBanner(ctx context.Context, banner *banner_model.BannerInsert) (int, error) {
	service.logger.Debug("insert banner serivce")

	return service.repo.InsertBanner(ctx, banner)
}

func (service *bannerService) GetUserBanner(ctx context.Context,
	params *query_params.BannerUserParams) (interface{}, error) {

	service.logger.Debug("get user banner service")

	if params.UseLastrRevision == "false" {
		banner, ok, err := service.cache.Get(ctx, banner_model.BannerKey{
			TagID:     params.TagID,
			FeatureID: params.FeatureID,
		})
		if err != nil {
			return "", err
		}
		if ok {
			if !banner.IsActive && !service.tokenProvider.VerifyOnAdmin(params.Token) {
				return "", pgx.ErrNoRows
			}
			return banner.Content, nil
		}
	}

	banner, err := service.repo.GetUserBanner(ctx, params.TagID, params.FeatureID)
	if err != nil {
		return "", err
	}
	if !banner.IsActive && !service.tokenProvider.VerifyOnAdmin(params.Token) {
		return "", pgx.ErrNoRows
	}
	go service.cache.Add(context.Background(), banner_model.BannerKey{
		TagID:     params.TagID,
		FeatureID: params.FeatureID,
	}, &banner)

	return banner.Content, nil
}

func (service *bannerService) GetBanners(context context.Context, params *query_params.BannerParams) ([]banner_model.Banner,
	error) {

	service.logger.Debug("get banners")
	return service.repo.GetBanners(context, params)
}

func (service *bannerService) DeleteBanner(context context.Context, id int) (bool, error) {
	service.logger.Debug("delete banner", slog.Int("id", id))

	return service.repo.DeleteBanner(context, id)
}

func (service *bannerService) UpdateBanner(context context.Context, banner *banner_model.BannerUpdate) error {
	service.logger.Debug("update banner", slog.Int("id", banner.ID))

	return service.repo.UpdateBanner(context, banner)
}

func (service *bannerService) DeleteBanners(context context.Context, params *query_params.DeleteBannerParams) {
	service.logger.Debug("delete banner params", slog.Any("params", params))

	service.repo.DeleteBanners(context, params)
}
