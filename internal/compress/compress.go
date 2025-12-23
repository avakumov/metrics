package compress

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipDecoderMiddleware распаковывает gzip-сжатые POST-запросы
func GzipDecoderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что это POST запрос и есть заголовок Content-Encoding: gzip
		contentType := r.Header.Get("Content-Type")
		if contentType == "application/json" || contentType == "text/html" {

			if r.Method == http.MethodPost &&
				strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {

				// Создаем gzip reader
				gz, err := gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "failed to create gzip reader: "+err.Error(),
						http.StatusBadRequest)
					return
				}
				defer gz.Close()

				// Заменяем тело запроса распакованными данными
				r.Body = io.NopCloser(gz)

				// Удаляем заголовок Content-Encoding, так как данные теперь распакованы
				r.Header.Del("Content-Encoding")
			}
		}

		next.ServeHTTP(w, r)
	})
}

// GzipResponseWriter для кодирования ответов
type GzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (g *GzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

// GzipEncoderMiddleware - ДЛЯ ИСХОДЯЩИХ ДАННЫХ
func GzipEncoderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем поддержку gzip клиентом
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Создаем gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Устанавливаем заголовки
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Создаем обертку для ResponseWriter
		gzw := &GzipResponseWriter{
			Writer:         gz,
			ResponseWriter: w,
		}

		next.ServeHTTP(gzw, r)
	})
}
