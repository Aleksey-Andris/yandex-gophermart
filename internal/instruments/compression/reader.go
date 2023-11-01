package compression

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	contentEncoding = "Content-Encoding"
)

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

func Decompress(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if isContentGzip(req) {
			cr, err := newCompressReader(req.Body)
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			req.Body = cr
			defer cr.Close()
		}
		next.ServeHTTP(res, req)
	})
}

func isContentGzip(r *http.Request) bool {
	for _, s := range strings.Split(r.Header.Get(contentEncoding), ",") {
		if s == "gzip" {
			return true
		}
	}
	return false
}
