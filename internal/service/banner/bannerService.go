package banner_service

import (
	"context"
	"log/slog"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	banner_repository "github.com/Heatdog/Avito/internal/repository/banner"
)

type BannerService interface {
	InsertBanner(context context.Context, banner banner_model.BannerInsert) (int, error)
	GetUserBanner(context context.Context, params banner_model.BannerUserParams) (string, error)
	GetBanners(context context.Context, params banner_model.BannerParams) ([]banner_model.Banner, error)
}

type bannerService struct {
	logger *slog.Logger
	repo   banner_repository.BannerRepository
}

func NewBannerService(logger *slog.Logger, repo banner_repository.BannerRepository) BannerService {
	return &bannerService{
		logger: logger,
		repo:   repo,
	}
}

func (service *bannerService) InsertBanner(context context.Context, banner banner_model.BannerInsert) (int, error) {
	service.logger.Debug("insert banner serivce")

	return service.repo.InsertBanner(context, banner)
}

func (service *bannerService) GetUserBanner(context context.Context, params banner_model.BannerUserParams) (string, error) {
	service.logger.Debug("get user banner service")

	return service.repo.GetUserBanner(context, params)
}

func (service *bannerService) GetBanners(context context.Context, params banner_model.BannerParams) ([]banner_model.Banner,
	error) {

	service.logger.Debug("get banners")
	return service.repo.GetBanners(context, params)
}
