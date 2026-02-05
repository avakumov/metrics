package middlewares

import (
	"bytes"
	"io"
	"net/http"

	"github.com/avakumov/metrics/internal/logger"
	"github.com/avakumov/metrics/internal/utils"
	"go.uber.org/zap"
)

func CheckHashMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hash := r.Header.Get("HashSHA256")
			if key != "" {
				data, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "error read body", http.StatusInternalServerError)
					return
				}
				// ВОССТАНАВЛИВАЕМ тело для дальнейшего использования
				r.Body = io.NopCloser(bytes.NewReader(data))

				hashSHA256, err := utils.HashHMACSHA256(data, key)

				if err != nil {
					logger.Log.Error("failed calculate hash", zap.Error(err))
					http.Error(w, "failed calculate hash: "+err.Error(),
						http.StatusBadRequest)
					return
				}
				if hash != hashSHA256 {
					http.Error(w, "hashSHA256 is wrong", http.StatusBadRequest)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AddHashHeaderMiddleware(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// 1. Создаем ДВА буфера: один для хеширования, один для ответа
			hashBuffer := &bytes.Buffer{}
			responseBuffer := &bytes.Buffer{}

			// 2. Создаем writer, который пишет в оба буфера
			hw := &hashingWriter{
				ResponseWriter: w,
				hashBuffer:     hashBuffer,
				responseBuffer: responseBuffer,
			}

			// 3. ВЫПОЛНЯЕМ ОБРАБОТЧИК (все пишется в буферы)
			next.ServeHTTP(hw, r)

			// 4. Теперь у нас есть полный ответ в буферах
			if hashBuffer.Len() > 0 {
				// Вычисляем хеш
				hash, err := utils.HashHMACSHA256(hashBuffer.Bytes(), key)
				if err != nil {
					// Отправляем без хеша при ошибке
					w.WriteHeader(hw.statusCode)
					_, err := w.Write(responseBuffer.Bytes())
					if err != nil {
						logger.Log.Error("error write body", zap.Error(err))
					}
					return
				}

				// 5. Устанавливаем заголовок ПЕРЕД отправкой
				w.Header().Set("HashSHA256", hash)

				// 6. Отправляем статус
				w.WriteHeader(hw.statusCode)

				// 7. Отправляем тело
				_, err = w.Write(responseBuffer.Bytes())
				if err != nil {
					logger.Log.Error("error write body", zap.Error(err))
				}
			} else {
				// Пустой ответ
				w.WriteHeader(hw.statusCode)
			}
		})
	}
}

type hashingWriter struct {
	http.ResponseWriter
	hashBuffer     *bytes.Buffer
	responseBuffer *bytes.Buffer
	statusCode     int
	wroteHeader    bool
}

func (hw *hashingWriter) Write(p []byte) (int, error) {
	if !hw.wroteHeader {
		hw.WriteHeader(http.StatusOK)
	}

	// Пишем в оба буфера
	hw.hashBuffer.Write(p)
	return hw.responseBuffer.Write(p)
}

func (hw *hashingWriter) WriteHeader(statusCode int) {
	if !hw.wroteHeader {
		hw.statusCode = statusCode
		hw.wroteHeader = true
		// НЕ вызываем hw.ResponseWriter.WriteHeader() здесь!
		// Отложим до вычисления хеша
	}
}

func (hw *hashingWriter) Header() http.Header {
	return hw.ResponseWriter.Header()
}
