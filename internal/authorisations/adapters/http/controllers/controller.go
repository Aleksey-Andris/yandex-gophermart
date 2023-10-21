package controllers

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Aleksey-Andris/yandex-gophermart/internal/authorisations"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/db"
	"github.com/Aleksey-Andris/yandex-gophermart/internal/instruments/logger"
	"github.com/go-chi/chi"
)

const (
	сontentType         = "Content-Type"
	сontentTypeAppJSON  = "application/json"
	сontentTypeAppXGZIP = "application/x-gzip"
)

type Usecase interface {
	Register(ctx context.Context, auth *authorisations.Auth) (*authorisations.Auth, error)
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
	return router
}

func (c *controller) register(res http.ResponseWriter, req *http.Request) {
	ct := strings.Split(req.Header.Get(сontentType), ";")[0]
	if !(ct == сontentTypeAppJSON || ct == сontentTypeAppXGZIP) {
		c.logger.Errorf("Register: invalid content type: %s. Session ID: %s", ct, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		c.logger.Errorf("Register: failed reading body, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Failed reading body", http.StatusBadRequest)
		return
	}

	c.logger.Infof("Register: body: %s. Session ID: %s", string(body), c.logger.GetSesionID(req.Context()))

	var request authorisations.Auth
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&request); err != nil {
		c.logger.Errorf("Register: failed decoding body, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Invalid body", http.StatusBadRequest)
		return
	}
	if request.Login == "" || request.Password == "" {
		c.logger.Errorf("Register: incorrect body fields, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
		http.Error(res, "Incorrect body fields", http.StatusBadRequest)
		return
	}
	request.Password = generatePasswordHash(request.Password)

	_, err = c.usecase.Register(req.Context(), &request)
	var status int
	if err != nil {
		if !errors.Is(err, db.ErrConflict) {
			c.logger.Errorf("Register: failed to registration, err value: %s. Session ID: %s", err, c.logger.GetSesionID(req.Context()))
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
		status = http.StatusConflict
	} else {
		status = http.StatusCreated
	}

	res.WriteHeader(status)
}

func generatePasswordHash(password string) string {
	salt := "82hduhuesjdjj"
	hash := sha1.New()
	hash.Write([]byte(password))
	hashString := fmt.Sprintf("%x", hash.Sum([]byte(salt)))
	return hashString
}
