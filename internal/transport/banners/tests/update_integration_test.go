package banner_handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	banner_postgre "github.com/Heatdog/Avito/internal/repository/banner/postgre"
	banner_service "github.com/Heatdog/Avito/internal/service/bannerservice"
	banners_transport "github.com/Heatdog/Avito/internal/transport/banners"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	hashicorp_lru "github.com/Heatdog/Avito/pkg/cache/hashi_corp"
	simpletoken "github.com/Heatdog/Avito/pkg/token/simple_token"
	"github.com/gorilla/mux"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v2"
	"github.com/stretchr/testify/require"
)

func TestUpdateBanner(t *testing.T) {
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

	type RespID struct {
		ID int `json:"banner_id"`
	}

	type mockBehavior func(banner banner_model.BannerUpdate, err error)

	testTable := []struct {
		name      string
		path      string
		token     string
		reqBanner banner_model.BannerUpdate

		bannerID *RespID

		statusCode int
		err        error

		mockFunc mockBehavior
	}{
		{
			name:  "ok",
			path:  "/banner/1",
			token: "admin_token",
			reqBanner: banner_model.BannerUpdate{
				TagsID:    &[]int{1, 2, 3},
				FeatureID: Int(1),
				Content: map[string]interface{}{
					"title": "123",
					"text":  "456",
				},
				IsActive: Bool(false),
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(banner banner_model.BannerUpdate, _ error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				dbMock.ExpectExec("UPDATE banners").
					WithArgs(banner.Content, *banner.IsActive, banner.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				row := pgxmock.NewRows([]string{"feature_id", "tag_id"})
				row.AddRow(1, 2)

				dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners").
					WithArgs(banner.ID).
					WillReturnRows(row)

				dbMock.ExpectExec("DELETE FROM features_tags_to_banners").
					WithArgs(banner.ID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))

				for _, tagID := range *banner.TagsID {
					dbMock.ExpectExec("INSERT INTO features_tags_to_banners").
						WithArgs(*banner.FeatureID, tagID, banner.ID).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
				}
			},
		},
		{
			name:  "not found",
			path:  "/banner/21321321312",
			token: "admin_token",
			reqBanner: banner_model.BannerUpdate{
				FeatureID: Int(2),
			},

			bannerID: &RespID{
				ID: 21321321312,
			},

			statusCode: http.StatusNotFound,
			err:        nil,

			mockFunc: func(banner banner_model.BannerUpdate, _ error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectRollback()

				dbMock.ExpectExec("UPDATE banners").
					WithArgs(banner.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 0))
			},
		},
		{
			name:  "active field update only",
			path:  "/banner/1",
			token: "admin_token",
			reqBanner: banner_model.BannerUpdate{
				IsActive: Bool(false),
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(banner banner_model.BannerUpdate, _ error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				dbMock.ExpectExec("UPDATE banners").
					WithArgs(*banner.IsActive, banner.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))
			},
		},
		{
			name:  "tag field update only",
			path:  "/banner/1",
			token: "admin_token",
			reqBanner: banner_model.BannerUpdate{
				TagsID: &[]int{1, 2, 3},
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(banner banner_model.BannerUpdate, _ error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				dbMock.ExpectExec("UPDATE banners").
					WithArgs(banner.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				row := pgxmock.NewRows([]string{"feature_id", "tag_id"})
				row.AddRow(1, 2)

				dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners").
					WithArgs(banner.ID).
					WillReturnRows(row)

				dbMock.ExpectExec("DELETE FROM features_tags_to_banners").
					WithArgs(banner.ID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))

				for _, tagID := range *banner.TagsID {
					dbMock.ExpectExec("INSERT INTO features_tags_to_banners").
						WithArgs(1, tagID, banner.ID).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
				}
			},
		},
		{
			name:  "feature_id update only",
			path:  "/banner/1",
			token: "admin_token",
			reqBanner: banner_model.BannerUpdate{
				FeatureID: Int(2),
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusOK,
			err:        nil,

			mockFunc: func(banner banner_model.BannerUpdate, _ error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				dbMock.ExpectExec("UPDATE banners").
					WithArgs(banner.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				row := pgxmock.NewRows([]string{"feature_id", "tag_id"})
				row.AddRow(1, 2)

				dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners").
					WithArgs(banner.ID).
					WillReturnRows(row)

				dbMock.ExpectExec("DELETE FROM features_tags_to_banners").
					WithArgs(banner.ID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))

				dbMock.ExpectExec("INSERT INTO features_tags_to_banners").
					WithArgs(*banner.FeatureID, 2, banner.ID).
					WillReturnResult(pgxmock.NewResult("INSERT", 1))
			},
		},
		{
			name:  "dublicated keys error",
			path:  "/banner/1",
			token: "admin_token",
			reqBanner: banner_model.BannerUpdate{
				FeatureID: Int(2),
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusInternalServerError,
			err:        fmt.Errorf("dublicated keys"),

			mockFunc: func(banner banner_model.BannerUpdate, err error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectRollback()

				dbMock.ExpectExec("UPDATE banners").
					WithArgs(banner.ID).
					WillReturnResult(pgxmock.NewResult("UPDATE", 1))

				row := pgxmock.NewRows([]string{"feature_id", "tag_id"})
				row.AddRow(1, 2)

				dbMock.ExpectQuery("SELECT feature_id, tag_id FROM features_tags_to_banners").
					WithArgs(banner.ID).
					WillReturnRows(row)

				dbMock.ExpectExec("DELETE FROM features_tags_to_banners").
					WithArgs(banner.ID).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))

				dbMock.ExpectExec("INSERT INTO features_tags_to_banners").
					WithArgs(*banner.FeatureID, 2, banner.ID).
					WillReturnError(err)
			},
		},
		{
			name:  "internal error",
			path:  "/banner/1",
			token: "admin_token",
			reqBanner: banner_model.BannerUpdate{
				FeatureID: Int(2),
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusInternalServerError,
			err:        fmt.Errorf("internal error"),

			mockFunc: func(banner banner_model.BannerUpdate, err error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectRollback()

				dbMock.ExpectExec("UPDATE banners").
					WithArgs(banner.ID).
					WillReturnError(err)
			},
		},
		{
			name:     "Forbidden",
			path:     "/banner/1",
			token:    "user_token",
			bannerID: nil,

			statusCode: http.StatusForbidden,
			err:        nil,

			mockFunc: func(_ banner_model.BannerUpdate, _ error) {},
		},
		{
			name:     "Unauthorized",
			path:     "/banner/1",
			token:    "123",
			bannerID: nil,

			statusCode: http.StatusUnauthorized,
			err:        nil,

			mockFunc: func(_ banner_model.BannerUpdate, _ error) {},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			body, err := json.Marshal(testCase.reqBanner)
			if err != nil {
				t.Fatal(err)
			}

			if testCase.bannerID != nil {
				testCase.reqBanner.ID = testCase.bannerID.ID
			}

			r := httptest.NewRequest(http.MethodPatch, testCase.path, bytes.NewBuffer(body))

			r.Header.Set("token", testCase.token)

			if testCase.bannerID != nil {
				testCase.mockFunc(testCase.reqBanner, testCase.err)
			}

			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			resp := w.Result()
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			var expected []byte

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

			require.Equal(t, testCase.statusCode, w.Code)
			require.Equal(t, string(expected), string(data))
		})
	}
}
