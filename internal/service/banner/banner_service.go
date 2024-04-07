package banner_service

import (
	"context"
	"log/slog"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/query_params"
	banner_repository "github.com/Heatdog/Avito/internal/repository/banner"
	"github.com/Heatdog/Avito/pkg/cache"
	"github.com/jackc/pgx/v5"
)

type BannerService interface {
	InsertBanner(context context.Context, banner banner_model.BannerInsert) (int, error)
	GetUserBanner(context context.Context, params query_params.BannerUserParams) (interface{}, error)
	GetBanners(context context.Context, params query_params.BannerParams) ([]banner_model.Banner, error)
}

type bannerService struct {
	logger *slog.Logger
	repo   banner_repository.BannerRepository
	cache  cache.Cache[banner_model.CacheKey, *banner_model.Banner]
}

func NewBannerService(logger *slog.Logger, repo banner_repository.BannerRepository,
	cache cache.Cache[banner_model.CacheKey, *banner_model.Banner]) BannerService {
	return &bannerService{
		logger: logger,
		repo:   repo,
		cache:  cache,
	}
}

func (service *bannerService) InsertBanner(context context.Context, banner banner_model.BannerInsert) (int, error) {
	service.logger.Debug("insert banner serivce")

	return service.repo.InsertBanner(context, banner)
}

func (service *bannerService) GetUserBanner(context context.Context,
	params query_params.BannerUserParams) (interface{}, error) {

	service.logger.Debug("get user banner service")

	if !params.UseLastrRevision {
		banner, ok := service.cache.Get(banner_model.CacheKey{
			TagID:     params.TagID,
			FeatureID: params.FeatureID,
		})
		if ok {
			if !banner.IsActive {
				return "", pgx.ErrNoRows
			}
			return banner.Content, nil
		}
	}

	banner, err := service.repo.GetUserBanner(context, params.TagID, params.FeatureID)
	if err != nil {
		return "", err
	}
	if !banner.IsActive {
		return "", pgx.ErrNoRows
	}
	go service.cache.Add(banner_model.CacheKey{
		TagID:     params.TagID,
		FeatureID: params.FeatureID,
	}, &banner)

	return banner.Content, nil
}

func (service *bannerService) GetBanners(context context.Context, params query_params.BannerParams) ([]banner_model.Banner,
	error) {

	service.logger.Debug("get banners")
	return service.repo.GetBanners(context, params)
}
