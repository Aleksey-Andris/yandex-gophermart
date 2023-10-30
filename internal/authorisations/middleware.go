package authorisations

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	SigningKey = "ushjdhui38487"
	TokenExp   = time.Hour * 3
	userCTX ygmAuthContext = "YGMUserID"
)

type ygmAuthContext string

type TokenClaims struct {
	jwt.RegisteredClaims
	UserID int64
}

func UserIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		cookieToken, err := req.Cookie("token")
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				http.Error(res, "Token not present", http.StatusUnauthorized)
				return
			}
			http.Error(res, "Failed to get token", http.StatusInternalServerError)
			return
		}

		auth, valid, err := parseToken(cookieToken.Value)
		if err != nil {
			http.Error(res, "Failed to parse token", http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(res, "Invalid token", http.StatusUnauthorized)
			return
		}

		request := req.WithContext(context.WithValue(req.Context(), userCTX, auth.UserID))
		next.ServeHTTP(res, request)
	})
}

func parseToken(tokenString string) (*TokenClaims, bool, error) {
	claims := &TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %s", t.Header["alg"])
			}
			return []byte(SigningKey), nil
		})

	if err != nil {
		return nil, false, err
	}

	return claims, token.Valid, err
}

func GetUserID(ctx context.Context) int64 {
	ctxVal := ctx.Value(userCTX)
	userID, _ := ctxVal.(int64)
	return userID
}
