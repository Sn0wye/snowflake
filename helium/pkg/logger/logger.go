package logger

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const LOGGER_KEY = "zapLogger"

type Logger struct {
	*zap.Logger
}

func NewLog(conf *viper.Viper) *Logger {
	return initZap(conf)
}

func initZap(conf *viper.Viper) *Logger {
	lp := conf.GetString("log.log_file_name")
	lv := conf.GetString("log.level")
	var level zapcore.Level
	switch lv {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	var encoder zapcore.Encoder
	if conf.GetString("log.encoding") == "console" {
		encoder = zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "Logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseColorLevelEncoder,
			EncodeTime:     timeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.FullCallerEncoder,
		})
	} else {
		encoder = zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		})
	}

	var syncer zapcore.WriteSyncer
	if lp == "" {
		syncer = zapcore.AddSync(os.Stdout)
	} else {
		hook := lumberjack.Logger{
			Filename:   lp,
			MaxSize:    conf.GetInt("log.max_size"),
			MaxBackups: conf.GetInt("log.max_backups"),
			MaxAge:     conf.GetInt("log.max_age"),
			Compress:   conf.GetBool("log.compress"),
		}
		syncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(&hook))
	}

	core := zapcore.NewCore(encoder, syncer, level)
	// service identifies the origin service on every log line for log aggregation.
	serviceField := zap.Fields(zap.String("service", "helium"))
	if conf.GetString("env") != "production" {
		return &Logger{zap.New(core, serviceField, zap.Development(), zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))}
	}
	return &Logger{zap.New(core, serviceField, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))}
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000000000"))
}

func (l *Logger) NewContext(ctx *fiber.Ctx, fields ...zapcore.Field) {
	ctx.Locals(LOGGER_KEY, l.With(fields...))
}

func (l *Logger) WithContext(ctx *fiber.Ctx) *Logger {
	if ctx == nil {
		return l
	}
	zl := ctx.Locals(LOGGER_KEY)
	if ctxLogger, ok := zl.(*zap.Logger); ok {
		return &Logger{ctxLogger}
	}
	return l
}
