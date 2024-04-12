package banner_handler_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/query_params"
	banner_postgre "github.com/Heatdog/Avito/internal/repository/banner/postgre"
	banner_service "github.com/Heatdog/Avito/internal/service/banner"
	banners_transport "github.com/Heatdog/Avito/internal/transport/banners"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	hashicorp_lru "github.com/Heatdog/Avito/pkg/cache/hashi_corp"
	"github.com/Heatdog/Avito/pkg/token/simple_token"
	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestMultyDeleteBanner(t *testing.T) {
	dbMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer dbMock.Close()

	opt := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelError,
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, opt))
	slog.SetDefault(logger)

	cacheLRU := expirable.NewLRU[banner_model.BannerKey, *banner_model.Banner](0, nil,
		time.Minute*time.Duration(5))
	cache := hashicorp_lru.NewLRU(logger, cacheLRU)

	tokenProvider := simple_token.NewSimpleTokenProvider()

	logger.Debug("register middlewre")
	middleware := middleware_transport.NewMiddleware(logger, tokenProvider)

	bannerRepo := banner_postgre.NewBannerRepository(logger, dbMock)
	bannerService := banner_service.NewBannerService(logger, bannerRepo, cache, tokenProvider)
	bannerHandler := banners_transport.NewBannersHandler(logger, bannerService, middleware)
	router := mux.NewRouter()

	bannerHandler.Register(router)

	type mockBehavior func(params *query_params.DeleteBannerParams, deletedBanners []int, err error)

	testTable := []struct {
		name  string
		path  string
		token string

		params         *query_params.DeleteBannerParams
		deletedBanners []int

		statusCode int
		err        error

		mockFunc mockBehavior
	}{
		{
			name:  "ok",
			path:  "/banner",
			token: "admin_token",

			params: &query_params.DeleteBannerParams{
				TagID:     Int(1),
				FeatureID: Int(2),
			},
			deletedBanners: []int{1, 2, 3},

			statusCode: http.StatusAccepted,
			err:        nil,

			mockFunc: func(params *query_params.DeleteBannerParams, deletedBanners []int, err error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				rows := pgxmock.NewRows([]string{"banner_id"})
				for _, id := range deletedBanners {
					rows.AddRow(id)
				}

				dbMock.ExpectQuery("SELECT banner_id FROM features_tags_to_banners").
					WithArgs(*params.TagID, *params.FeatureID).
					WillReturnRows(rows)

				for _, id := range deletedBanners {
					dbMock.ExpectExec("DELETE FROM banners").
						WithArgs(id).
						WillReturnResult(pgxmock.NewResult("DELETE", 1))
				}
			},
		},
		{
			name:  "internal error",
			path:  "/banner",
			token: "admin_token",

			params: &query_params.DeleteBannerParams{
				TagID:     Int(1),
				FeatureID: Int(2),
			},
			deletedBanners: []int{1, 2, 3},

			statusCode: http.StatusAccepted,
			err:        fmt.Errorf("internal error"),

			mockFunc: func(params *query_params.DeleteBannerParams, deletedBanners []int, err error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectRollback()

				dbMock.ExpectQuery("SELECT banner_id FROM features_tags_to_banners").
					WithArgs(*params.TagID, *params.FeatureID).
					WillReturnError(err)
			},
		},
		{
			name:  "tag only",
			path:  "/banner",
			token: "admin_token",

			params: &query_params.DeleteBannerParams{
				TagID: Int(1),
			},
			deletedBanners: []int{1, 2, 3},

			statusCode: http.StatusAccepted,
			err:        nil,

			mockFunc: func(params *query_params.DeleteBannerParams, deletedBanners []int, err error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				rows := pgxmock.NewRows([]string{"banner_id"})
				for _, id := range deletedBanners {
					rows.AddRow(id)
				}

				dbMock.ExpectQuery("SELECT banner_id FROM features_tags_to_banners").
					WithArgs(*params.TagID).
					WillReturnRows(rows)

				for _, id := range deletedBanners {
					dbMock.ExpectExec("DELETE FROM banners").
						WithArgs(id).
						WillReturnResult(pgxmock.NewResult("DELETE", 1))
				}
			},
		},
		{
			name:  "Forbidden",
			path:  "/banner",
			token: "user_token",

			params:         nil,
			deletedBanners: nil,

			statusCode: http.StatusForbidden,
			err:        nil,

			mockFunc: func(params *query_params.DeleteBannerParams, deletedBanners []int, err error) {},
		},
		{
			name:  "Unauthorized",
			path:  "/banner/1",
			token: "123",

			params:         nil,
			deletedBanners: nil,

			statusCode: http.StatusUnauthorized,
			err:        nil,

			mockFunc: func(params *query_params.DeleteBannerParams, deletedBanners []int, err error) {},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {

			r := httptest.NewRequest(http.MethodDelete, testCase.path, nil)

			r.Header.Set("token", testCase.token)

			testCase.mockFunc(testCase.params, testCase.deletedBanners, testCase.err)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			require.Equal(t, testCase.statusCode, w.Code)
		})
	}
}
