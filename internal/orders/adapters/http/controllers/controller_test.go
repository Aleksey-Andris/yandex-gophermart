package controllers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations/adapters/http/controllers"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/orders/mocks"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Controller_addOne(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l, err := logger.New("develop")
	if err != nil {
		log.Fatalf("Logger_test: failed creating: %s", err)
	}
	defer l.Sync()

	storage := mocks.NewMockStorage(c)
	usecase := orders.New(l, storage)
	controller := New(l, usecase)

	router := chi.NewRouter()
	router.Use(l.WithLogging)
	router.Mount("/api/user/orders", controller.Init())

	testServ := httptest.NewServer(router)
	defer testServ.Close()

	type mocBehavior func(s *mocks.MockStorage)
	tests := []struct {
		name               string
		requestURL         string
		requestContentType string
		reguestBody        []byte
		expectedStatusCode int
		mocBehavior        mocBehavior
	}{
		{
			name:               "Correct case",
			requestURL:         "/api/user/orders",
			requestContentType: "text/plain",
			reguestBody:        []byte(`12345678903`),
			expectedStatusCode: http.StatusAccepted,
			mocBehavior: func(s *mocks.MockStorage) {
				ord := &orders.Order{
					ID:  1,
					Num: "12345678903",
				}
				s.EXPECT().AddOne(gomock.Any(), gomock.Any()).Return(ord, nil)
			},
		},
		{
			name:               "Invalid nums format - not Loon",
			requestURL:         "/api/user/orders",
			requestContentType: "text/plain",
			reguestBody:        []byte(`12345678903100500`),
			expectedStatusCode: http.StatusUnprocessableEntity,
			mocBehavior:        func(s *mocks.MockStorage) {},
		},
		{
			name:               "Invalid nums format - not num",
			requestURL:         "/api/user/orders",
			requestContentType: "text/plain",
			reguestBody:        []byte(`1234567890sd3100500`),
			expectedStatusCode: http.StatusUnprocessableEntity,
			mocBehavior:        func(s *mocks.MockStorage) {},
		},
		{
			name:               "Conflict",
			requestURL:         "/api/user/orders",
			requestContentType: "text/plain",
			reguestBody:        []byte(`12345678903`),
			expectedStatusCode: http.StatusConflict,
			mocBehavior: func(s *mocks.MockStorage) {
				s.EXPECT().AddOne(gomock.Any(), gomock.Any()).Return(nil, db.ErrConflict)
			},
		},
		{
			name:               "Row exist",
			requestURL:         "/api/user/orders",
			requestContentType: "text/plain",
			reguestBody:        []byte(`12345678903`),
			expectedStatusCode: http.StatusOK,
			mocBehavior: func(s *mocks.MockStorage) {
				s.EXPECT().AddOne(gomock.Any(), gomock.Any()).Return(nil, db.ErrRowExist)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mocBehavior(storage)

			req, err := http.NewRequest(http.MethodPost, testServ.URL+tc.requestURL, bytes.NewBufferString(string(tc.reguestBody)))
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

func Test_Controller_getAll(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l, err := logger.New("develop")
	if err != nil {
		log.Fatalf("Logger_test: failed creating: %s", err)
	}
	defer l.Sync()

	storage := mocks.NewMockStorage(c)
	usecase := orders.New(l, storage)
	controller := New(l, usecase)

	router := chi.NewRouter()
	router.Use(l.WithLogging)
	router.Mount("/api/user/orders", controller.Init())

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
			requestURL:         "/api/user/orders",
			expectedStatusCode: http.StatusOK,
			wantAuth:           true,
			expectedErr:        false,
			mocBehavior: func(s *mocks.MockStorage) {
				ords := []orders.Order{{
					ID:  1,
					Num: "12345678903",
				}}
				s.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(ords, nil)
			},
		},
		{
			name:               "No auths",
			requestURL:         "/api/user/orders",
			expectedStatusCode: http.StatusUnauthorized,
			wantAuth:           false,
			expectedErr:        true,
			mocBehavior:        func(s *mocks.MockStorage) {},
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
				err = json.NewDecoder(res.Body).Decode(&[]orders.Order{})
				require.NoError(t, err)
			}
		})
	}
}
