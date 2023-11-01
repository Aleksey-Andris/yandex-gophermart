package controllers

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations/mocks"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func Test_Controller_register(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l, err := logger.New("develop")
	if err != nil {
		log.Fatalf("Logger_test: failed creating: %s", err)
	}
	defer l.Sync()

	storage := mocks.NewMockStorage(c)
	usecase := authorisations.New(l, storage)
	controller := New(l, usecase)

	router := chi.NewRouter()
	router.Use(l.WithLogging)
	router.Mount("/api/user/", controller.Init())

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
			requestURL:         "/api/user/register",
			requestContentType: "application/json",
			reguestBody:        []byte(`{"login": "any_user","password": "123"} `),
			expectedStatusCode: http.StatusOK,
			mocBehavior: func(s *mocks.MockStorage) {
				auth := &authorisations.Auth{
					ID:       1,
					Login:    "any_user",
					Password: generatePasswordHash("123"),
				}
				s.EXPECT().Register(gomock.Any(), gomock.Any()).Return(auth, nil)
				s.EXPECT().Login(gomock.Any(), gomock.Any()).Return(auth, nil)
			},
		},

		{
			name:               "Conflict",
			requestURL:         "/api/user/register",
			requestContentType: "application/json",
			reguestBody:        []byte(`{"login": "any_user","password": "123"} `),
			expectedStatusCode: http.StatusConflict,
			mocBehavior: func(s *mocks.MockStorage) {
				s.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, db.ErrConflict)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mocBehavior(storage)

			req, err := http.NewRequest(http.MethodPost, testServ.URL+tc.requestURL, bytes.NewBufferString(string(tc.reguestBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tc.requestContentType)

			res, err := testServ.Client().Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
		})
	}
}

func Test_Controller_login(t *testing.T) {
	c := gomock.NewController(t)
	defer c.Finish()

	l, err := logger.New("develop")
	if err != nil {
		log.Fatalf("Logger_test: failed creating: %s", err)
	}
	defer l.Sync()

	storage := mocks.NewMockStorage(c)
	usecase := authorisations.New(l, storage)
	controller := New(l, usecase)

	router := chi.NewRouter()
	router.Use(l.WithLogging)
	router.Mount("/api/user/", controller.Init())

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
			requestURL:         "/api/user/login",
			requestContentType: "application/json",
			reguestBody:        []byte(`{"login": "any_user","password": "123"} `),
			expectedStatusCode: http.StatusOK,
			mocBehavior: func(s *mocks.MockStorage) {
				auth := &authorisations.Auth{
					ID:       1,
					Login:    "any_user",
					Password: generatePasswordHash("123"),
				}
				s.EXPECT().Login(gomock.Any(), gomock.Any()).Return(auth, nil)
			},
		},

		{
			name:               "User no existe",
			requestURL:         "/api/user/login",
			requestContentType: "application/json",
			reguestBody:        []byte(`{"login": "any_user","password": "123"} `),
			expectedStatusCode: http.StatusUnauthorized,
			mocBehavior: func(s *mocks.MockStorage) {
				s.EXPECT().Login(gomock.Any(), gomock.Any()).Return(nil, db.ErrNoRows)
			},
		}, 
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.mocBehavior(storage)

			req, err := http.NewRequest(http.MethodPost, testServ.URL+tc.requestURL, bytes.NewBufferString(string(tc.reguestBody)))
			require.NoError(t, err)
			req.Header.Set("Content-Type", tc.requestContentType)

			res, err := testServ.Client().Do(req)
			require.NoError(t, err)
			defer res.Body.Close()

			assert.Equal(t, tc.expectedStatusCode, res.StatusCode)
		})
	}
}
