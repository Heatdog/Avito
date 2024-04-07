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
	InsertBanner(context context.Context, banner banner_model.BannerInsert) (int, error)
	GetUserBanner(context context.Context, params query_params.BannerUserParams) (interface{}, error)
	GetBanners(context context.Context, params query_params.BannerParams) ([]banner_model.Banner, error)
	DeleteBanner(context context.Context, id int) (bool, error)
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

func (service *bannerService) InsertBanner(context context.Context, banner banner_model.BannerInsert) (int, error) {
	service.logger.Debug("insert banner serivce")

	return service.repo.InsertBanner(context, banner)
}

func (service *bannerService) GetUserBanner(context context.Context,
	params query_params.BannerUserParams) (interface{}, error) {

	service.logger.Debug("get user banner service")

	if !params.UseLastrRevision {
		banner, ok := service.cache.Get(banner_model.BannerKey{
			TagID:     params.TagID,
			FeatureID: params.FeatureID,
		})
		if ok {
			if !banner.IsActive && !service.tokenProvider.VerifyOnAdmin(params.Token) {
				return "", pgx.ErrNoRows
			}
			return banner.Content, nil
		}
	}

	banner, err := service.repo.GetUserBanner(context, params.TagID, params.FeatureID)
	if err != nil {
		return "", err
	}
	if !banner.IsActive && !service.tokenProvider.VerifyOnAdmin(params.Token) {
		return "", pgx.ErrNoRows
	}
	go service.cache.Add(banner_model.BannerKey{
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

func (service *bannerService) DeleteBanner(context context.Context, id int) (bool, error) {
	service.logger.Debug("delete banner", slog.Int("id", id))

	params, err := service.repo.GetBannerParams(context, id)
	if err != nil {
		return false, err
	}

	ok, err := service.repo.DeleteBanner(context, id)
	if err != nil {
		return false, err
	}
	if !ok {
		return false, nil
	}

	for _, tagID := range params.TagIDs {
		service.cache.Remove(banner_model.BannerKey{
			FeatureID: params.FeatureID,
			TagID:     tagID,
		})
	}

	return true, nil
}
