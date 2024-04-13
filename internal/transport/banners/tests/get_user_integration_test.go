package banner_handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/queryparams"
	banner_postgre "github.com/Heatdog/Avito/internal/repository/banner/postgre"
	banner_service "github.com/Heatdog/Avito/internal/service/bannerservice"
	banners_transport "github.com/Heatdog/Avito/internal/transport/banners"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	hashicorp_lru "github.com/Heatdog/Avito/pkg/cache/hashi_corp"
	simpletoken "github.com/Heatdog/Avito/pkg/token/simple_token"
	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func TestGetUserBanner(t *testing.T) {
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

	tokenProvider := simpletoken.NewSimpleTokenProvider()

	logger.Debug("register middlewre")
	middleware := middleware_transport.NewMiddleware(logger, tokenProvider)

	bannerRepo := banner_postgre.NewBannerRepository(logger, dbMock)
	bannerService := banner_service.NewBannerService(logger, bannerRepo, cache, tokenProvider)
	bannerHandler := banners_transport.NewBannersHandler(logger, bannerService, middleware)
	router := mux.NewRouter()

	bannerHandler.Register(router)

	type mockBehavior func(banners *banner_model.Banner, params queryparams.BannerUserParams, err error)

	testTable := []struct {
		name   string
		path   string
		token  string
		params queryparams.BannerUserParams

		respBanners *banner_model.Banner
		statusCode  int
		err         error

		mockFunc mockBehavior
	}{
		{
			name:  "ok",
			path:  "/user_banner?tag_id=1&feature_id=1&use_last_revision=true&version=1",
			token: "user_token",
			params: queryparams.BannerUserParams{
				TagID:            "1",
				FeatureID:        "1",
				UseLastrRevision: "true",
				Version:          "1",
				Token:            "user_token",
			},

			respBanners: &banner_model.Banner{
				ID:        1,
				TagsID:    []int{1, 2, 3},
				FeatureID: 4,
				ContentV1: `{"title":"good_title3"}`,
				ContentV2: nil,
				ContentV3: nil,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(banner *banner_model.Banner, params queryparams.BannerUserParams, _ error) {
				row := pgxmock.NewRows([]string{"id", "content_v1", "content_v2", "content_v3", "is_active"})
				row.AddRow(banner.ID, banner.ContentV1, banner.ContentV2, banner.ContentV3, banner.IsActive)

				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active
					FROM banners b JOIN features_tags_to_banners ftb`).
					WithArgs(params.FeatureID, params.TagID).
					WillReturnRows(row)
			},
		},
		{
			name:  "admin token",
			path:  "/user_banner?tag_id=1&feature_id=1&use_last_revision=true&version=1",
			token: "admin_token",
			params: queryparams.BannerUserParams{
				TagID:            "1",
				FeatureID:        "1",
				UseLastrRevision: "true",
				Version:          "1",
				Token:            "user_token",
			},

			respBanners: &banner_model.Banner{
				ID:        1,
				TagsID:    []int{1, 2, 3},
				FeatureID: 4,
				ContentV1: `{"title":"good_title3"}`,
				ContentV2: nil,
				ContentV3: nil,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(banner *banner_model.Banner, params queryparams.BannerUserParams, _ error) {
				row := pgxmock.NewRows([]string{"id", "content_v1", "content_v2", "content_v3", "is_active"})
				row.AddRow(banner.ID, banner.ContentV1, banner.ContentV2, banner.ContentV3, banner.IsActive)

				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active
					FROM banners b JOIN features_tags_to_banners ftb`).
					WithArgs(params.FeatureID, params.TagID).
					WillReturnRows(row)
			},
		},
		{
			name:  "banner not found",
			path:  "/user_banner?tag_id=5&feature_id=5&use_last_revision=false&version=1",
			token: "user_token",
			params: queryparams.BannerUserParams{
				TagID:            "5",
				FeatureID:        "5",
				UseLastrRevision: "true",
				Version:          "1",
				Token:            "user_token",
			},

			respBanners: nil,
			statusCode:  http.StatusNotFound,
			err:         nil,

			mockFunc: func(_ *banner_model.Banner, params queryparams.BannerUserParams, _ error) {
				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active
					FROM banners b JOIN features_tags_to_banners ftb`).
					WithArgs(params.FeatureID, params.TagID).
					WillReturnError(pgx.ErrNoRows)
			},
		},
		{
			name:  "use cache",
			path:  "/user_banner?tag_id=1&feature_id=1&use_last_revision=false&version=1",
			token: "admin_token",
			params: queryparams.BannerUserParams{
				TagID:            "1",
				FeatureID:        "1",
				UseLastrRevision: "false",
				Version:          "1",
				Token:            "user_token",
			},

			respBanners: &banner_model.Banner{
				ID:        1,
				TagsID:    []int{1, 2, 3},
				FeatureID: 4,
				ContentV1: `{"title":"good_title3"}`,
				ContentV2: nil,
				ContentV3: nil,
				IsActive:  true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(banner *banner_model.Banner, _ queryparams.BannerUserParams, _ error) {
				for _, tag := range banner.TagsID {
					if _, err := cache.Add(context.Background(), banner_model.BannerKey{
						TagID:     strconv.Itoa(tag),
						FeatureID: strconv.Itoa(banner.FeatureID),
					}, banner); err != nil {
						return
					}
				}
			},
		},
		{
			name:  "internal error",
			path:  "/user_banner?tag_id=1&feature_id=1&use_last_revision=true&version=1",
			token: "admin_token",
			params: queryparams.BannerUserParams{
				TagID:            "1",
				FeatureID:        "1",
				UseLastrRevision: "true",
				Version:          "1",
				Token:            "user_token",
			},

			respBanners: nil,
			statusCode:  http.StatusInternalServerError,
			err:         fmt.Errorf("internal error"),

			mockFunc: func(_ *banner_model.Banner, params queryparams.BannerUserParams, err error) {
				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active
					FROM banners b JOIN features_tags_to_banners ftb`).
					WithArgs(params.FeatureID, params.TagID).
					WillReturnError(err)
			},
		},
		{
			name:   "no params",
			path:   "/user_banner",
			token:  "user_token",
			params: queryparams.BannerUserParams{},

			respBanners: nil,
			statusCode:  http.StatusBadRequest,
			err:         fmt.Errorf("Key: 'BannerUserParams.TagID' Error:Field validation for 'TagID' failed on the 'required' tag\nKey: 'BannerUserParams.FeatureID' Error:Field validation for 'FeatureID' failed on the 'required' tag"),

			mockFunc: func(_ *banner_model.Banner, _ queryparams.BannerUserParams, _ error) {},
		},
		{
			name:   "no token",
			path:   "/user_banner",
			token:  "12321",
			params: queryparams.BannerUserParams{},

			respBanners: nil,
			statusCode:  http.StatusUnauthorized,
			err:         nil,

			mockFunc: func(_ *banner_model.Banner, _ queryparams.BannerUserParams, _ error) {},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, testCase.path, nil)

			r.Header.Set("token", testCase.token)

			testCase.mockFunc(testCase.respBanners, testCase.params, testCase.err)

			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			var expected []byte
			if testCase.respBanners != nil {
				expected, err = json.Marshal(testCase.respBanners.ContentV1)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				if testCase.err != nil {
					expected, err = json.Marshal(struct {
						Err string `json:"error"`
					}{
						Err: testCase.err.Error(),
					})
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			require.Equal(t, testCase.statusCode, w.Code)
			require.Equal(t, string(expected), string(data))
		})
	}
}
