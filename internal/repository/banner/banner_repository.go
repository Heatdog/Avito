package banner_repository

import (
	"context"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/query_params"
)

type BannerRepository interface {
	InsertBanner(ctx context.Context, banner *banner_model.BannerInsert) (int, error)
	GetUserBanner(ctx context.Context, tagID, feautureID string) (banner_model.Banner, error)
	GetBanners(ctx context.Context, params *query_params.BannerParams) ([]banner_model.Banner, error)
	GetBannerParams(ctx context.Context, id int) (banner_model.BannerParams, error)
	DeleteBanner(ctx context.Context, id int) (bool, error)
	UpdateBanner(ctx context.Context, banner *banner_model.BannerUpdate) error
	DeleteBanners(ctx context.Context, params query_params.DeleteBannerParams) error
	UpdateBannerVersion(ctx context.Context, id, version int) error
}
