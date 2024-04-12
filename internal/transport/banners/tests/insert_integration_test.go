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

func TestInsertBanner(t *testing.T) {
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

	type mockBehavior func(banner banner_model.BannerInsert, id int, err error)

	testTable := []struct {
		name      string
		path      string
		token     string
		reqBanner banner_model.BannerInsert

		bannerID *RespID

		statusCode int
		err        error

		mockFunc mockBehavior
	}{
		{
			name:  "ok",
			path:  "/banner",
			token: "admin_token",
			reqBanner: banner_model.BannerInsert{
				TagsID:    []int{1, 2, 3},
				FeatureID: 1,
				Content: map[string]interface{}{
					"title": "123",
					"text":  "456",
				},
				IsActive: true,
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusCreated,
			err:        nil,

			mockFunc: func(banner banner_model.BannerInsert, id int, _ error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				row := pgxmock.NewRows([]string{"id"})
				row.AddRow(id)

				dbMock.ExpectQuery("INSERT INTO banners").
					WithArgs(banner.Content, banner.IsActive).
					WillReturnRows(row)

				for _, tag := range banner.TagsID {
					dbMock.ExpectExec("INSERT INTO features_tags_to_banners").
						WithArgs(banner.FeatureID, tag, id).
						WillReturnResult(pgxmock.NewResult("INSERT", 1))
				}
			},
		},
		{
			name:     "Forbidden",
			path:     "/banner",
			token:    "user_token",
			bannerID: nil,

			statusCode: http.StatusForbidden,
			err:        nil,

			mockFunc: func(_ banner_model.BannerInsert, _ int, _ error) {},
		},
		{
			name:     "Unauthorized",
			path:     "/banner",
			token:    "123",
			bannerID: nil,

			statusCode: http.StatusUnauthorized,
			err:        nil,

			mockFunc: func(_ banner_model.BannerInsert, _ int, _ error) {},
		},
		{
			name:     "validation error",
			path:     "/banner",
			token:    "admin_token",
			bannerID: nil,
			reqBanner: banner_model.BannerInsert{
				TagsID:    []int{1, 2, 3},
				FeatureID: 1,
				Content:   "13231",
				IsActive:  true,
			},

			statusCode: http.StatusBadRequest,
			err:        fmt.Errorf(`Key: 'BannerInsert.Content' Error:Field validation for 'Content' failed on the 'json' tag`),

			mockFunc: func(_ banner_model.BannerInsert, _ int, _ error) {},
		},
		{
			name:  "dublication error",
			path:  "/banner",
			token: "admin_token",
			reqBanner: banner_model.BannerInsert{
				TagsID:    []int{1, 2, 3},
				FeatureID: 1,
				Content: map[string]interface{}{
					"title": "123",
					"text":  "456",
				},
				IsActive: true,
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusInternalServerError,
			err:        fmt.Errorf("ERROR: duplicate key value violates unique constraint \"features_tags_to_banners_pk\" (SQLSTATE 23505)"),

			mockFunc: func(banner banner_model.BannerInsert, id int, err error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectRollback()

				row := pgxmock.NewRows([]string{"id"})
				row.AddRow(id)

				dbMock.ExpectQuery("INSERT INTO banners").
					WithArgs(banner.Content, banner.IsActive).
					WillReturnRows(row)

				dbMock.ExpectExec("INSERT INTO features_tags_to_banners").
					WithArgs(banner.FeatureID, banner.TagsID[0], id).
					WillReturnError(err)
			},
		},
		{
			name:  "internal error",
			path:  "/banner",
			token: "admin_token",
			reqBanner: banner_model.BannerInsert{
				TagsID:    []int{1, 2, 3},
				FeatureID: 1,
				Content: map[string]interface{}{
					"title": "123",
					"text":  "456",
				},
				IsActive: true,
			},

			bannerID: &RespID{
				ID: 1,
			},

			statusCode: http.StatusInternalServerError,
			err:        fmt.Errorf("internal error"),

			mockFunc: func(banner banner_model.BannerInsert, _ int, err error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectRollback()

				dbMock.ExpectQuery("INSERT INTO banners").
					WithArgs(banner.Content, banner.IsActive).
					WillReturnError(err)
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			body, err := json.Marshal(testCase.reqBanner)
			if err != nil {
				t.Fatal(err)
			}

			r := httptest.NewRequest(http.MethodPost, testCase.path, bytes.NewBuffer(body))

			r.Header.Set("token", testCase.token)

			if testCase.bannerID != nil {
				testCase.mockFunc(testCase.reqBanner, testCase.bannerID.ID, testCase.err)
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
			if testCase.bannerID != nil && testCase.err == nil {
				expected, err = json.Marshal(testCase.bannerID)
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
