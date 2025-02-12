package logger

import "go.uber.org/zap"

// MPMLogger Интерфейс для внедрения логгеров во всех микросервисах проекта MPMS
type MPMLogger interface {
	// Debug Записывает сообщение уровня Debug.
	Debug(msg string, fields ...zap.Field)
	// Info Записывает сообщение уровня Info.
	Info(msg string, fields ...zap.Field)
	// Warn Записывает сообщение уровня Warn.
	Warn(msg string, fields ...zap.Field)
	// Error Записывает сообщение уровня Error.
	Error(msg string, fields ...zap.Field)
	// Dpanic Записывает сообщение уровня DPanic.
	DPanic(msg string, fields ...zap.Field)
	// Panic Записывает сообщение уровня Panic.
	Panic(msg string, fields ...zap.Field)
	// Fatal Записывает сообщение уровня Fatal.
	Fatal(msg string, fields ...zap.Field)
	// With Строит новый логер, который будет иметь все поля текущего логера.
	With(fields ...zap.Field) *zap.Logger
	// Sync Ожидает, что все сообщения будут записаны в выходной поток (обычно используется при завершении программы).
	Sync() error
}
