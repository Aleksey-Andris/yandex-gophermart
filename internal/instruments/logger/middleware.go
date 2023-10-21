package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	sessionCTX YGMLogContext = "YGMSessionID"
)

type YGMLogContext string

type responseData struct {
	status int
	size   int
}

type logginResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (res *logginResponseWriter) Write(b []byte) (int, error) {
	size, err := res.ResponseWriter.Write(b)
	res.responseData.size += size
	return size, err
}

func (res *logginResponseWriter) WriteHeader(statusCode int) {
	res.ResponseWriter.WriteHeader(statusCode)
	res.responseData.status = statusCode
}

func (l *Logger) WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		start := time.Now()
		sessionID := uuid.New()
		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lRes := logginResponseWriter{
			ResponseWriter: res,
			responseData:   responseData,
		}
		request := req.WithContext(context.WithValue(req.Context(), sessionCTX, sessionID))
		next.ServeHTTP(&lRes, request)
		duration := time.Since(start)

		l.Info( 
			"SessionID: ", sessionID.String(),
			" URI: ", req.RequestURI,
			" Method: ", req.Method,
			" Status: ", responseData.status,
			" Duartion: ", duration,
			" Size: ", responseData.size,
		)
	})
}

func (l *Logger) GetSesionID(ctx context.Context) string {
	ctxVal := ctx.Value(sessionCTX)
	sessionID, _ := ctxVal.(uuid.UUID)
	return sessionID.String()
}
