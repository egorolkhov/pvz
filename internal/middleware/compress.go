package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

type CompressWrite struct {
	http.ResponseWriter
	zw *gzip.Writer
}

func (c *CompressWrite) Write(b []byte) (int, error) {
	res, err := c.zw.Write(b)
	return res, err
}

type CompressRead struct {
	io.ReadCloser
	zr *gzip.Reader
}

func (c *CompressRead) Read(p []byte) (int, error) {
	res, err := c.zr.Read(p)
	return res, err
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gw, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return gw
	},
}
var gzipReaderPool = sync.Pool{
	New: func() interface{} {
		return &gzip.Reader{}
	},
}

func CompressMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(w)

		defer func() {
			gz.Close()
			gzipWriterPool.Put(gz)
		}()

		cw := &CompressWrite{
			ResponseWriter: w,
			zw:             gz,
		}
		cw.Header().Set("Content-Encoding", "gzip")
		cw.Header().Del("Content-Length")

		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzR := gzipReaderPool.Get().(*gzip.Reader)
			err := gzR.Reset(r.Body)
			if err != nil {
				http.Error(w, "failed to create gzip reader", http.StatusInternalServerError)
				return
			}
			defer func() {
				gzR.Close()
				gzipReaderPool.Put(gzR)
			}()
			r.Body = &CompressRead{
				ReadCloser: r.Body,
				zr:         gzR,
			}
		}

		next.ServeHTTP(cw, r)
	})
}
