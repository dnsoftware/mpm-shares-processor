package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/dnsoftware/mpm-shares-processor/internal/constants"

	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"
)

const LogLevelProduction = "production"
const LogLevelDebug = "debug"

type Logger struct {
	*zap.Logger
}

var (
	instance *Logger
	once     sync.Once
)

func getLogLevel(env string) zapcore.Level {
	if env == LogLevelProduction {
		return zapcore.InfoLevel
	}
	return zapcore.DebugLevel
}

func getFileWriter(filePath string) zapcore.WriteSyncer {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return zapcore.AddSync(file)
}

func InitLogger(env string, filePath string) {
	once.Do(func() {
		logLevel := getLogLevel(env)
		encoderConfig := zap.NewProductionEncoderConfig()

		if env == LogLevelProduction {
			encoderConfig = zap.NewProductionEncoderConfig()
			encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		}
		encoder := zapcore.NewJSONEncoder(encoderConfig)
		core := zapcore.NewTee(
			zapcore.NewCore(encoder, getFileWriter(filePath), logLevel),
		)

		if env != LogLevelProduction {
			// Добавить вывод в консоль в режиме отладки
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // добавим подсветку при выводе в консоль в режиме отладки
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
			consoleCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), logLevel)
			core = zapcore.NewTee(core, consoleCore)
		}

		logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		instance = &Logger{logger}
	})
}

func Log() *Logger {
	if instance == nil {
		panic("Logger is not initialized. Call InitLogger() before using GetLogger()")
	}
	return instance
}

func GetLoggerMainLogPath() (string, error) {
	dir, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		return "", err
	}
	filePath := dir + "/" + constants.AppLogFile

	return filePath, nil
}

func GetLoggerTestLogPath() (string, error) {
	dir, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		return "", err
	}
	filePath := dir + "/" + constants.TestLogFile

	return filePath, nil
}
