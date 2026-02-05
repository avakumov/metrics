package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Простой constraint для числовых типов
type Number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func Float64Ptr[T Number](value T) *float64 {
	f := float64(value)
	return &f
}

func HashHMACSHA256(data []byte, key string) (string, error) {
	// Создаём HMAC с алгоритмом SHA256 и секретным ключом
	h := hmac.New(sha256.New, []byte(key))
	// Записываем данные для хеширования
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	// Получаем итоговый HMAC
	dst := h.Sum(nil)
	// Конвертируем в hex строку
	return hex.EncodeToString(dst), nil
}
