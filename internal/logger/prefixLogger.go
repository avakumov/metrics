package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

// Используется исключительно для разработки
func PrefixLogger(prefix string) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")

	// Создаем кастомный encoder
	encoder := &prefixConsoleEncoder{
		Encoder: zapcore.NewConsoleEncoder(config.EncoderConfig),
		prefix:  "[" + prefix + "] ",
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zapcore.DebugLevel,
	)

	return zap.New(core, zap.AddCaller())
}

type prefixConsoleEncoder struct {
	zapcore.Encoder
	prefix string
}

func (e *prefixConsoleEncoder) Clone() zapcore.Encoder {
	return &prefixConsoleEncoder{
		Encoder: e.Encoder.Clone(),
		prefix:  e.prefix,
	}
}

func (e *prefixConsoleEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// Получаем сгенерированную строку от родительского encoder
	buf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	// Получаем строку
	str := buf.String()
	buf.Free()

	// Находим позицию времени (формат "15:04:05")
	// В строке будет что-то вроде: "15:04:05\tINFO\tmessage..."
	// Нам нужно вставить префикс перед временем

	// Создаем новую строку с префиксом в начале
	prefixed := e.prefix + str

	// Создаем новый буфер
	newBuf := buffer.NewPool().Get()
	newBuf.AppendString(prefixed)

	return newBuf, nil
}
