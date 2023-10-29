package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations/adapters/http/controllers"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/balances/mocks"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Controller_get(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l, err := logger.New("develop")
	if err != nil {
		log.Fatalf("Logger_test: failed creating: %s", err)
	}
	defer l.Sync()

	storage := mocks.NewMockStorage(c)
	usecase := balances.New(l, storage)
	controller := New(l, usecase)

	router := chi.NewRouter()
	router.Use(l.WithLogging)
	router.Mount("/api/user/balance/", controller.Init())

	testServ := httptest.NewServer(router)
	defer testServ.Close()

	type mocBehavior func(s *mocks.MockStorage)
	tests := []struct {
		name               string
		requestURL         string
		expectedStatusCode int
		expectedErr        bool
		mocBehavior        mocBehavior
	}{
		{
			name:               "Correct case",
			requestURL:         "/api/user/balance/",
			expectedStatusCode: http.StatusOK,
			expectedErr:        false,
			mocBehavior: func(s *mocks.MockStorage) {
				balance := &balances.Balance{
					Current:   100500,
					Withdrawn: 10,
				}
				s.EXPECT().Get(gomock.Any(), gomock.Any()).Return(balance, nil)
			},
		},
		{
			name:               "Intrnal err",
			requestURL:         "/api/user/balance/",
			expectedStatusCode: http.StatusInternalServerError,
			expectedErr:        true,
			mocBehavior: func(s *mocks.MockStorage) {
				s.EXPECT().Get(gomock.Any(), gomock.Any()).Return(nil, errors.New("any err"))
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mocBehavior(storage)

			req, err := http.NewRequest(http.MethodGet, testServ.URL+tc.requestURL, bytes.NewBufferString(""))
			require.NoError(t, err)
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

			res, err := testServ.Client().Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
			if !tc.expectedErr {
				err = json.NewDecoder(res.Body).Decode(&balances.Balance{})
				require.NoError(t, err)
			}
		})
	}
}

func Test_Controller_spend(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l, err := logger.New("develop")
	if err != nil {
		log.Fatalf("Logger_test: failed creating: %s", err)
	}
	defer l.Sync()

	storage := mocks.NewMockStorage(c)
	usecase := balances.New(l, storage)
	controller := New(l, usecase)

	router := chi.NewRouter()
	router.Use(l.WithLogging)
	router.Mount("/api/user/balance/", controller.Init())

	testServ := httptest.NewServer(router)
	defer testServ.Close()

	type mocBehavior func(s *mocks.MockStorage)
	tests := []struct {
		name               string
		requestURL         string
		requestContentType string
		reguestBody        string
		expectedStatusCode int
		mocBehavior        mocBehavior
	}{
		{
			name:               "Correct case",
			requestURL:         "/api/user/balance/withdraw",
			requestContentType: "application/json",
			reguestBody:        `{"order": "4026843483168683","sum": 100500}`,
			expectedStatusCode: http.StatusOK,
			mocBehavior: func(s *mocks.MockStorage) {
				op := &balances.Operation{
					Ord:    "4026843483168683",
					Amount: 100500,
				}
				s.EXPECT().Spend(gomock.Any(), gomock.Any()).Return(op, nil)
			},
		},
		{
			name:               "Not enough balls",
			requestURL:         "/api/user/balance/withdraw",
			requestContentType: "application/json",
			reguestBody:        `{"order": "4026843483168683","sum": 100500}`,
			expectedStatusCode: http.StatusPaymentRequired,
			mocBehavior: func(s *mocks.MockStorage) {
				op := &balances.Operation{
					Ord:    "4026843483168683",
					Amount: 100500,
					Result: balances.ResultNotEnough,
				}
				s.EXPECT().Spend(gomock.Any(), gomock.Any()).Return(op, nil)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mocBehavior(storage)

			req, err := http.NewRequest(http.MethodPost, testServ.URL+tc.requestURL, bytes.NewBufferString(tc.reguestBody))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tc.requestContentType)
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

			res, err := testServ.Client().Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
		})
	}
}
