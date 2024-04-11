package banner_handler_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	banner_postgre "github.com/Heatdog/Avito/internal/repository/banner/postgre"
	banner_service "github.com/Heatdog/Avito/internal/service/banner"
	banners_transport "github.com/Heatdog/Avito/internal/transport/banners"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	hashicorp_lru "github.com/Heatdog/Avito/pkg/cache/hashi_corp"
	"github.com/Heatdog/Avito/pkg/token/simple_token"
	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func Int(i int) *int    { return &i }
func Bool(b bool) *bool { return &b }

func TestGetBanners(t *testing.T) {
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

	type queryParams struct {
		TagID     *int
		FeatureID *int
	}
	type mockBehavior func(dbMock pgxmock.PgxPoolIface, banners []banner_model.Banner, params queryParams)

	testTable := []struct {
		name   string
		path   string
		token  string
		params queryParams

		respBanners []banner_model.Banner
		statusCode  int
		err         error

		mockFunc mockBehavior
	}{
		{
			name:   "ok",
			path:   "/banner",
			token:  "admin_token",
			params: queryParams{},

			respBanners: []banner_model.Banner{
				{
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
				{
					ID:        2,
					TagsID:    []int{4},
					FeatureID: 5,
					ContentV1: `{"title":"good_title1"}`,
					ContentV2: `{"title":"good_title2"}`,
					ContentV3: `{"title":"good_title3"}`,
					IsActive:  false,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(dbMock pgxmock.PgxPoolIface, banners []banner_model.Banner, params queryParams) {

				rows := pgxmock.NewRows([]string{"id", "content_v1", "content_v2", "content_v3", "is_active",
					"created_at", "updated_at"})
				for _, banner := range banners {
					rows.AddRow(banner.ID, banner.ContentV1, banner.ContentV2, banner.ContentV3, banner.IsActive,
						banner.CreatedAt, banner.UpdatedAt)
				}

				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active, b.created_at, 
				b.updated_at FROM banners b`).
					WillReturnRows(rows)

				var tagFeature []*pgxmock.Rows
				for _, banner := range banners {
					bannerRows := pgxmock.NewRows([]string{"feature_id", "tag_id"})
					for _, tag := range banner.TagsID {
						bannerRows.AddRow(banner.FeatureID, tag)
					}
					tagFeature = append(tagFeature, bannerRows)
				}

				for i, banner := range banners {
					dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners WHERE banner_id").
						WithArgs(banner.ID).
						WillReturnRows(tagFeature[i])
				}
			},
		},
		{
			name:  "ok with all params",
			path:  "/banner?tag_id=1&feature_id=1&limit=1&offset=1",
			token: "admin_token",
			params: queryParams{
				TagID:     Int(1),
				FeatureID: Int(1),
			},

			respBanners: []banner_model.Banner{
				{
					ID:        1,
					TagsID:    []int{1},
					FeatureID: 1,
					ContentV1: `{"title":"good_title3"}`,
					ContentV2: nil,
					ContentV3: nil,
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(dbMock pgxmock.PgxPoolIface, banners []banner_model.Banner, params queryParams) {

				rows := pgxmock.NewRows([]string{"id", "content_v1", "content_v2", "content_v3", "is_active",
					"created_at", "updated_at"})
				for _, banner := range banners {
					rows.AddRow(banner.ID, banner.ContentV1, banner.ContentV2, banner.ContentV3, banner.IsActive,
						banner.CreatedAt, banner.UpdatedAt)
				}

				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active, b.created_at, 
				b.updated_at FROM banners b JOIN features_tags_to_banners ftb`).
					WithArgs(&params.FeatureID, &params.TagID).
					WillReturnRows(rows)

				var tagFeature []*pgxmock.Rows
				for _, banner := range banners {
					bannerRows := pgxmock.NewRows([]string{"feature_id", "tag_id"})
					for _, tag := range banner.TagsID {
						bannerRows.AddRow(banner.FeatureID, tag)
					}
					tagFeature = append(tagFeature, bannerRows)
				}

				for i, banner := range banners {
					dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners WHERE banner_id").
						WithArgs(banner.ID).
						WillReturnRows(tagFeature[i])
				}
			},
		},
		{
			name:  "only feature",
			path:  "/banner?feature_id=1",
			token: "admin_token",
			params: queryParams{
				FeatureID: Int(1),
			},

			respBanners: []banner_model.Banner{
				{
					ID:        1,
					TagsID:    []int{1},
					FeatureID: 1,
					ContentV1: `{"title":"good_title3"}`,
					ContentV2: nil,
					ContentV3: nil,
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(dbMock pgxmock.PgxPoolIface, banners []banner_model.Banner, params queryParams) {

				rows := pgxmock.NewRows([]string{"id", "content_v1", "content_v2", "content_v3", "is_active",
					"created_at", "updated_at"})
				for _, banner := range banners {
					rows.AddRow(banner.ID, banner.ContentV1, banner.ContentV2, banner.ContentV3, banner.IsActive,
						banner.CreatedAt, banner.UpdatedAt)
				}

				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active, b.created_at, 
				b.updated_at FROM banners b JOIN features_tags_to_banners ftb`).
					WithArgs(&params.FeatureID).
					WillReturnRows(rows)

				var tagFeature []*pgxmock.Rows
				for _, banner := range banners {
					bannerRows := pgxmock.NewRows([]string{"feature_id", "tag_id"})
					for _, tag := range banner.TagsID {
						bannerRows.AddRow(banner.FeatureID, tag)
					}
					tagFeature = append(tagFeature, bannerRows)
				}

				for i, banner := range banners {
					dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners WHERE banner_id").
						WithArgs(banner.ID).
						WillReturnRows(tagFeature[i])
				}
			},
		},
		{
			name:   "only limits",
			path:   "/banner?limit=1&offset=1",
			token:  "admin_token",
			params: queryParams{},

			respBanners: []banner_model.Banner{
				{
					ID:        1,
					TagsID:    []int{1},
					FeatureID: 1,
					ContentV1: `{"title":"good_title3"}`,
					ContentV2: nil,
					ContentV3: nil,
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(dbMock pgxmock.PgxPoolIface, banners []banner_model.Banner, params queryParams) {

				rows := pgxmock.NewRows([]string{"id", "content_v1", "content_v2", "content_v3", "is_active",
					"created_at", "updated_at"})
				for _, banner := range banners {
					rows.AddRow(banner.ID, banner.ContentV1, banner.ContentV2, banner.ContentV3, banner.IsActive,
						banner.CreatedAt, banner.UpdatedAt)
				}

				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active, b.created_at, 
				b.updated_at FROM banners b`).
					WillReturnRows(rows)

				var tagFeature []*pgxmock.Rows
				for _, banner := range banners {
					bannerRows := pgxmock.NewRows([]string{"feature_id", "tag_id"})
					for _, tag := range banner.TagsID {
						bannerRows.AddRow(banner.FeatureID, tag)
					}
					tagFeature = append(tagFeature, bannerRows)
				}

				for i, banner := range banners {
					dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners WHERE banner_id").
						WithArgs(banner.ID).
						WillReturnRows(tagFeature[i])
				}
			},
		},
		{
			name:  "only tag",
			path:  "/banner?tag_id=1",
			token: "admin_token",
			params: queryParams{
				TagID: Int(1),
			},

			respBanners: []banner_model.Banner{
				{
					ID:        1,
					TagsID:    []int{1},
					FeatureID: 1,
					ContentV1: `{"title":"good_title3"}`,
					ContentV2: nil,
					ContentV3: nil,
					IsActive:  true,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(dbMock pgxmock.PgxPoolIface, banners []banner_model.Banner, params queryParams) {

				rows := pgxmock.NewRows([]string{"id", "content_v1", "content_v2", "content_v3", "is_active",
					"created_at", "updated_at"})
				for _, banner := range banners {
					rows.AddRow(banner.ID, banner.ContentV1, banner.ContentV2, banner.ContentV3, banner.IsActive,
						banner.CreatedAt, banner.UpdatedAt)
				}

				dbMock.ExpectQuery(`SELECT b.id, b.content_v1, b.content_v2, b.content_v3, b.is_active, b.created_at, 
				b.updated_at FROM banners b JOIN features_tags_to_banners ftb`).
					WithArgs(&params.TagID).
					WillReturnRows(rows)

				var tagFeature []*pgxmock.Rows
				for _, banner := range banners {
					bannerRows := pgxmock.NewRows([]string{"feature_id", "tag_id"})
					for _, tag := range banner.TagsID {
						bannerRows.AddRow(banner.FeatureID, tag)
					}
					tagFeature = append(tagFeature, bannerRows)
				}

				for i, banner := range banners {
					dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners WHERE banner_id").
						WithArgs(banner.ID).
						WillReturnRows(tagFeature[i])
				}
			},
		},
		{
			name:   "Unauthorized",
			path:   "/banner?tag_id=1&feature_id=1&limit=1&offset=1",
			params: queryParams{},
			token:  "123213213",

			respBanners: nil,
			statusCode:  http.StatusUnauthorized,
			err:         nil,

			mockFunc: func(dbMock pgxmock.PgxPoolIface, banners []banner_model.Banner, params queryParams) {},
		},
		{
			name:   "Forbidden",
			path:   "/banner?tag_id=1&feature_id=1&limit=1&offset=1",
			params: queryParams{},
			token:  "user_token",

			respBanners: nil,
			statusCode:  http.StatusForbidden,
			err:         nil,

			mockFunc: func(dbMock pgxmock.PgxPoolIface, banners []banner_model.Banner, params queryParams) {},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {

			r := httptest.NewRequest(http.MethodGet, testCase.path, nil)

			r.Header.Set("token", testCase.token)

			testCase.mockFunc(dbMock, testCase.respBanners, testCase.params)

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
				expected, err = json.Marshal(testCase.respBanners)
				if err != nil {
					t.Fatal(err)
				}
			}

			require.Equal(t, testCase.err, err)
			require.Equal(t, testCase.statusCode, w.Code)
			require.Equal(t, string(expected), string(data))
		})
	}
}
