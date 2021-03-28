package logger

import (
	fmt "fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/******************************************************************************
 * Типы
 */

// Color represents a text color.
type Color uint8

/******************************************************************************
 * Константы
 */

const (
	// Black Color = iota + 30
	Red Color = iota + 31
	Green
	Yellow
	Blue
	// Magenta
	// Cyan
	// White
)

/******************************************************************************
 * Переменные
 */

var (
	_levelToColor = map[zapcore.Level]Color{
		zapcore.DebugLevel:  Blue,
		zapcore.InfoLevel:   Green,
		zapcore.WarnLevel:   Yellow,
		zapcore.ErrorLevel:  Red,
		zapcore.DPanicLevel: Red,
		zapcore.PanicLevel:  Red,
		zapcore.FatalLevel:  Red,
	}
	_unknownLevelColor = Red

	_levelToLowercaseColorString = make(map[zapcore.Level]string, len(_levelToColor))
	_levelToCapitalColorString   = make(map[zapcore.Level]string, len(_levelToColor))
)

/******************************************************************************
 * Методы
 */

func (c Color) Add(s string) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(c), s)
}

/******************************************************************************
 * Функции
 */

func CapitalColorLevelEncoder(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	s, ok := _levelToCapitalColorString[l]
	if !ok {
		s = _unknownLevelColor.Add(l.CapitalString())
	}
	enc.AppendString(s)
}

func init() {
	for level, color := range _levelToColor {
		_levelToLowercaseColorString[level] = color.Add(level.String())
		_levelToCapitalColorString[level] = color.Add(level.CapitalString())
	}
}

func NewLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = CapitalColorLevelEncoder
	config.EncoderConfig.StacktraceKey = ""
	config.EncoderConfig.FunctionKey = "F"
	return config.Build()
}
