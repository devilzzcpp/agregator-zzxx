package common

import (
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger

// инициализирует zap с ротацией через lumberjack.
func InitLogger() {
	//todo сделать единный забор конфига для логгер
	logLevel := os.Getenv("LOG_LEVEL")
	logFile := os.Getenv("LOG_FILE")
	sizeStr := os.Getenv("LOG_MAX_SIZE")
	backupsStr := os.Getenv("LOG_MAX_BACKUPS")
	ageStr := os.Getenv("LOG_MAX_AGE")
	tzName := os.Getenv("APP_TIMEZONE")

	if sizeStr == "" {
		sizeStr = os.Getenv("LOG_SIZE")
	}
	if backupsStr == "" {
		backupsStr = os.Getenv("LOG_BACKUP")
	}
	if ageStr == "" {
		ageStr = os.Getenv("LOG_AGE")
	}

	if logLevel == "" {
		logLevel = "info"
	}
	if logFile == "" {
		logFile = "./storage/logs/app.log"
	}
	if sizeStr == "" {
		sizeStr = "50"
	}
	if backupsStr == "" {
		backupsStr = "5"
	}
	if ageStr == "" {
		ageStr = "30"
	}
	if tzName == "" {
		tzName = "Europe/Samara"
	}

	loc, err := time.LoadLocation(tzName)
	if err != nil {
		loc = time.FixedZone("UTC+4", 4*60*60)
	}

	Size, err := strconv.Atoi(sizeStr)
	if err != nil {
		Size = 50
	}
	Backups, err := strconv.Atoi(backupsStr)
	if err != nil {
		Backups = 5
	}
	Age, err := strconv.Atoi(ageStr)
	if err != nil {
		Age = 30
	}

	//уровень логирования
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = zapcore.InfoLevel
	}

	//ротация логов через lumberjack
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    Size,    // мегабайт до ротации
		MaxBackups: Backups, // сколько старых файлов хранить
		MaxAge:     Age,     // дней
		Compress:   true,    // gzip старые логи
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
