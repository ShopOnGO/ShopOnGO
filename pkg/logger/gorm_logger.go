package logger

import (
	"context"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// GormLogger — это наш кастомный логгер, реализующий интерфейс gorm.Logger.
type GormLogger struct {
	LogLevel      gormlogger.LogLevel
	SlowThreshold time.Duration
}

// NewGormLogger создает новый экземпляр GormLogger.
// Он будет использоваться в пакете db для инициализации.
func NewGormLogger() *GormLogger {
	return &GormLogger{
		LogLevel:      gormlogger.Silent,      // По умолчанию GORM не будет выводить ничего.
		SlowThreshold: 200 * time.Millisecond, // Порог для медленных запросов.
	}
}

// LogMode позволяет GORM менять уровень логов.
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info выводит информационные сообщения GORM через нашу функцию logger.Info.
func (l *GormLogger) Info(ctx context.Context, s string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		Infof(s, args...)
	}
}

// Warn выводит предупреждения GORM через нашу функцию logger.Warn.
// Мы отключаем Slow SQL, установив LogLevel в Silent в NewDB.
func (l *GormLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		Warnf(s, args...)
	}
}

// Error выводит ошибки GORM через нашу функцию logger.Error.
func (l *GormLogger) Error(ctx context.Context, s string, args ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		Errorf(s, args...)
	}
}

// Trace выводит SQL-запросы.
// Этот метод отвечает за логирование медленных запросов.
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormlogger.Error:
		sql, rows := fc()
		Errorf("%s\n[rows:%v] %s", err, rows, sql)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		Warnf("SLOW SQL >= %v\n[rows:%v] %s", l.SlowThreshold, rows, sql)
	case l.LogLevel >= gormlogger.Info:
		sql, rows := fc()
		Infof("[rows:%v] %s", rows, sql)
	}
}
