package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var logger *zap.SugaredLogger

type Config struct {
	Encoding string
	Level    string
	DevMode  bool
	Caller   bool
}
type appLogger struct {
	level       string
	devMode     bool
	encoding    string
	sugarLogger *zap.SugaredLogger
	logger      *zap.Logger
}

var loggerLevelMap = map[string]zapcore.Level{
	"DEBUG":  zapcore.DebugLevel,
	"INFO":   zapcore.InfoLevel,
	"WARN":   zapcore.WarnLevel,
	"ERROR":  zapcore.ErrorLevel,
	"DPANIC": zapcore.DPanicLevel,
	"PANIC":  zapcore.PanicLevel,
	"FATAL":  zapcore.FatalLevel,
}

func (l *appLogger) getLoggerLevel() zapcore.Level {
	level, exist := loggerLevelMap[l.level]
	if !exist {
		return zapcore.DebugLevel
	}

	return level
}
func (config *Config) SetConfiguration(appName string) {
	appLogger := &appLogger{level: config.Level, devMode: config.DevMode, encoding: config.Encoding}
	logLevel := appLogger.getLoggerLevel()

	logWriter := zapcore.AddSync(os.Stdout)

	var encoderCfg zapcore.EncoderConfig
	if appLogger.devMode {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	var encoder zapcore.Encoder
	encoderCfg.NameKey = "service"
	encoderCfg.TimeKey = "time"
	encoderCfg.LevelKey = "level"
	encoderCfg.CallerKey = "line"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

	if appLogger.encoding == "console" {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderCfg.EncodeCaller = zapcore.FullCallerEncoder
		encoderCfg.ConsoleSeparator = " | "
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoderCfg.FunctionKey = "caller"
		encoderCfg.EncodeName = zapcore.FullNameEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))
	lg := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	logger = lg.Sugar()
}

func Debug(args ...interface{}) {
	logger.Debug(args)
}

func Info(args ...interface{}) {
	logger.Info(args)
}

func Warn(args ...interface{}) {
	logger.Warn(args)
}

func Error(args ...interface{}) {
	logger.Error(args)
}

func DPanic(args ...interface{}) {
	logger.DPanic(args)
}

func Panic(args ...interface{}) {
	logger.Panic(args)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args)
}

func DebugF(template string, args ...interface{}) {
	logger.Debugf(template, args)
}

func InfoF(template string, args ...interface{}) {
	logger.Infof(template, args)
}

func WarnF(template string, args ...interface{}) {
	logger.Warnf(template, args)
}

func ErrorF(template string, args ...interface{}) {
	logger.Errorf(template, args)
}

func DPanicF(template string, args ...interface{}) {
	logger.DPanicf(template, args)
}

func PanicF(template string, args ...interface{}) {
	logger.Panicf(template, args)
}

func FatalF(template string, args ...interface{}) {
	logger.Fatalf(template, args)
}

func DebugW(msg string, keysAndValues ...interface{}) {
	logger.Debugw(msg, keysAndValues)
}

func InfoW(msg string, keysAndValues ...interface{}) {
	logger.Info(msg, keysAndValues)
}

func WarnW(msg string, keysAndValues ...interface{}) {
	logger.Warnw(msg, keysAndValues)
}

func ErrorW(msg string, keysAndValues ...interface{}) {
	logger.Errorw(msg, keysAndValues)
}

func DPanicW(msg string, keysAndValues ...interface{}) {
	logger.DPanicw(msg, keysAndValues)
}

func PanicW(msg string, keysAndValues ...interface{}) {
	logger.Info(msg, keysAndValues)
}

func FatalW(msg string, keysAndValues ...interface{}) {
	logger.Fatal(msg, keysAndValues)
}
