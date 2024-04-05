package banner_repository

import (
	"context"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
)

type BannerRepository interface {
	InsertBanner(ctx context.Context, banner banner_model.BannerInsert) (int, error)
}
