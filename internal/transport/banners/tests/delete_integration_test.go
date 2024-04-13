package banner_handler_test

import (
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
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func TestDeleteBanner(t *testing.T) {
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

	type mockBehavior func(id int, err error)

	testTable := []struct {
		name  string
		path  string
		token string

		bannerID int

		statusCode int
		err        error

		mockFunc mockBehavior
	}{
		{
			name:  "ok",
			path:  "/banner/1",
			token: "admin_token",

			bannerID:   1,
			statusCode: http.StatusNoContent,
			err:        nil,

			mockFunc: func(id int, _ error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				dbMock.ExpectExec("DELETE FROM banners").
					WithArgs(id).
					WillReturnResult(pgxmock.NewResult("DELETE", 1))
			},
		},
		{
			name:     "Forbidden",
			path:     "/banner/1",
			token:    "user_token",
			bannerID: 1,

			statusCode: http.StatusForbidden,
			err:        nil,

			mockFunc: func(_ int, _ error) {},
		},
		{
			name:     "Unauthorized",
			path:     "/banner/1",
			token:    "123",
			bannerID: 1,

			statusCode: http.StatusUnauthorized,
			err:        nil,

			mockFunc: func(_ int, _ error) {},
		},
		{
			name:     "not found",
			path:     "/banner/1",
			token:    "admin_token",
			bannerID: 1,

			statusCode: http.StatusNotFound,
			err:        nil,

			mockFunc: func(id int, _ error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectCommit()

				dbMock.ExpectExec("DELETE FROM banners").
					WithArgs(id).
					WillReturnResult(pgxmock.NewResult("DELETE", 0))
			},
		},
		{
			name:     "internal error",
			path:     "/banner/1",
			token:    "admin_token",
			bannerID: 1,

			statusCode: http.StatusInternalServerError,
			err:        fmt.Errorf("internal error"),

			mockFunc: func(id int, err error) {
				dbMock.ExpectBeginTx(pgx.TxOptions{})
				defer dbMock.ExpectRollback()

				dbMock.ExpectExec("DELETE FROM banners").
					WithArgs(id).
					WillReturnError(err)
			},
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodDelete, testCase.path, nil)

			r.Header.Set("token", testCase.token)

			testCase.mockFunc(testCase.bannerID, testCase.err)

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
