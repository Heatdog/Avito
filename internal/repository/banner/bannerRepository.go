package banner_repository

import (
	"context"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
)

type BannerRepository interface {
	InsertBanner(ctx context.Context, banner banner_model.BannerInsert) (int, error)
	GetUserBanner(ctx context.Context, params banner_model.BannerUserParams) (string, error)
	GetBanners(ctx context.Context, params banner_model.BannerParams) ([]banner_model.Banner, error)
}
