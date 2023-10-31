package controllers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations/adapters/http/controllers"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/withdrawals"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/withdrawals/mocks"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Controller_getAll(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l, err := logger.New("develop")
	if err != nil {
		log.Fatalf("Logger_test: failed creating: %s", err)
	}
	defer l.Sync()

	storage := mocks.NewMockStorage(c)
	usecase := withdrawals.New(l, storage)
	controller := New(l, usecase)

	router := chi.NewRouter()
	router.Use(l.WithLogging)
	router.Mount("/api/user/withdrawals", controller.Init())

	testServ := httptest.NewServer(router)
	defer testServ.Close()

	type mocBehavior func(s *mocks.MockStorage)
	tests := []struct {
		name               string
		requestURL         string
		expectedStatusCode int
		wantAuth           bool
		expectedErr        bool
		mocBehavior        mocBehavior
	}{
		{
			name:               "Correct case",
			requestURL:         "/api/user/withdrawals",
			expectedStatusCode: http.StatusOK,
			wantAuth:           true,
			expectedErr:        false,
			mocBehavior: func(s *mocks.MockStorage) {
				ords := []withdrawals.Withdrowal{{
					Ord:    "12345678903",
					Amount: 100,
					Data:   time.Now(),
				}}
				s.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(ords, nil)
			},
		},
		{
			name:               "No auths",
			requestURL:         "/api/user/withdrawals",
			expectedStatusCode: http.StatusUnauthorized,
			wantAuth:           false,
			expectedErr:        true,
			mocBehavior: func(s *mocks.MockStorage) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mocBehavior(storage)

			req, err := http.NewRequest(http.MethodGet, testServ.URL+tc.requestURL, bytes.NewBufferString(""))
			require.NoError(t, err)
			if tc.wantAuth {
				token, _ := controllers.BuildJWTString(&authorisations.Auth{
					ID:       1,
					Login:    "some login",
					Password: "some password",
				})
				req.AddCookie(&http.Cookie{
					Name:     "token",
					Value:    token,
					HttpOnly: true,
				})
			}

			res, err := testServ.Client().Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			if !tc.expectedErr {
				err = json.NewDecoder(res.Body).Decode(&[]withdrawals.Withdrowal{})
				require.NoError(t, err)
			}
		})
	}
}
