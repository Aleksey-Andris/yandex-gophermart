package controllers

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt/v4"
)

const (
	сontentType         = "Content-Type"
	сontentTypeAppJSON  = "application/json"
	сontentTypeAppXGZIP = "application/x-gzip"
)

type Usecase interface {
	Register(ctx context.Context, auth *authorisations.Auth) (*authorisations.Auth, error)
	Login(ctx context.Context, auth *authorisations.Auth) (*authorisations.Auth, error)
}

type controller struct {
	logger  *logger.Logger
	usecase Usecase
}

func New(logger *logger.Logger, usecase Usecase) *controller {
	return &controller{
		logger:  logger,
		usecase: usecase,
	}
}

func (c *controller) Init() *chi.Mux {
	router := chi.NewRouter()
	router.Post("/register", c.register)
	router.Post("/login", c.login)
	return router
}

func (c *controller) register(res http.ResponseWriter, req *http.Request) {
	ct := strings.Split(req.Header.Get(сontentType), ";")[0]
	if !(ct == сontentTypeAppJSON || ct == сontentTypeAppXGZIP) {
		c.logger.Errorf("Register: invalid content type: %s. Session ID: %s", ct, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	var request authorisations.Auth
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		c.logger.Errorf("Register: failed decoding body, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}
	if request.Login == "" || request.Password == "" {
		c.logger.Errorf("Register: incorrect body fields. Session ID: %s", c.logger.GetSesionID(req.Context()))
		http.Error(res, "Incorrect body fields", http.StatusBadRequest)
		return
	}
	request.Password = generatePasswordHash(request.Password)

	auth, err := c.usecase.Register(req.Context(), &request)
	if err != nil {
		if !errors.Is(err, db.ErrConflict) {
			c.logger.Errorf("Register: failed to registration, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusConflict)
		return
	}

	auth, err = c.usecase.Login(req.Context(), auth)
	if err != nil {
		c.logger.Errorf("Register: failed to login after registration, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Failed to login after registration", http.StatusInternalServerError)
		return
	}

	tokenVal, err := BuildJWTString(auth)
	if err != nil {
		c.logger.Errorf("Register: failed to build token after registration, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Failed to build tiken after registration", http.StatusInternalServerError)
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:     "token",
		Value:    tokenVal,
		HttpOnly: true,
	})

	res.WriteHeader(http.StatusOK)
}

func (c *controller) login(res http.ResponseWriter, req *http.Request) {
	ct := strings.Split(req.Header.Get(сontentType), ";")[0]
	if !(ct == сontentTypeAppJSON || ct == сontentTypeAppXGZIP) {
		c.logger.Errorf("Login: invalid content type: %s. Session ID: %s", ct, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	var request authorisations.Auth
	if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
		c.logger.Errorf("Login: failed decoding body, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}
	if request.Login == "" || request.Password == "" {
		c.logger.Errorf("Login: incorrect body fields. Session ID: %s", c.logger.GetSesionID(req.Context()))
		http.Error(res, "Incorrect body fields", http.StatusBadRequest)
		return
	}
	request.Password = generatePasswordHash(request.Password)

	auth, err := c.usecase.Login(req.Context(), &request)
	if err != nil {
		if !errors.Is(err, db.ErrNoRows) {
			c.logger.Errorf("Login: failed to login, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusUnauthorized)
		return
	}

	tokenVal, err := BuildJWTString(auth)
	if err != nil {
		c.logger.Errorf("Login: failed to build token after login, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Failed to build tiken after login", http.StatusInternalServerError)
		return
	}
	http.SetCookie(res, &http.Cookie{
		Name:     "token",
		Value:    tokenVal,
		HttpOnly: true,
	})

	res.WriteHeader(http.StatusOK)
}

func generatePasswordHash(password string) string {
	salt := "82hduhuesjdjj"
	hash := sha1.New()
	hash.Write([]byte(password))
	hashString := fmt.Sprintf("%x", hash.Sum([]byte(salt)))
	return hashString
}

func BuildJWTString(auth *authorisations.Auth) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authorisations.TokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(authorisations.TokenExp)),
		},
		UserID: auth.ID,
	})

	tokenString, err := token.SignedString([]byte(authorisations.SigningKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
