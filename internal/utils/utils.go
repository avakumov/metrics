package utils

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
