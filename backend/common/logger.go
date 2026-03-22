package common

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerConfig struct {
	Level      string
	File       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	TimeZone   string
}

var Logger *zap.Logger

// инициализирует zap с ротацией через lumberjack.
func InitLogger(cfg LoggerConfig) {
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.File == "" {
		cfg.File = "./storage/logs/app.log"
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 50
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 5
	}
	if cfg.MaxAge <= 0 {
		cfg.MaxAge = 30
	}
	if cfg.TimeZone == "" {
		cfg.TimeZone = "Europe/Samara"
	}

	loc, err := time.LoadLocation(cfg.TimeZone)
	if err != nil {
		loc = time.FixedZone("UTC+4", 4*60*60)
	}

	//уровень логирования
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	//ротация логов через lumberjack
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.File,
		MaxSize:    cfg.MaxSize,    // мегабайт до ротации
		MaxBackups: cfg.MaxBackups, // сколько старых файлов хранить
		MaxAge:     cfg.MaxAge,     // дней
		Compress:   true,           // gzip старые логи
	})

	consoleWriter := zapcore.AddSync(os.Stdout)

	//формат: JSON для файла, текст для консоли
	fileCfg := zap.NewProductionEncoderConfig()
	fileCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(loc).Format(time.RFC3339))
	}
	fileEncoder := zapcore.NewJSONEncoder(fileCfg)

	consoleCfg := zap.NewDevelopmentEncoderConfig()
	consoleCfg.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(loc).Format("2006-01-02 15:04:05 MST"))
	}
	consoleCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder //цветной уровень в консоли
	consoleEncoder := zapcore.NewConsoleEncoder(consoleCfg)

	//пишем и в файл и в консоль
	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, fileWriter, level),
		zapcore.NewCore(consoleEncoder, consoleWriter, level),
	)

	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// Sync сбрасывает буферы логгера — вызывать defer в main
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
