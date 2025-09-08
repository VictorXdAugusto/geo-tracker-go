package logger

import (
	"go.uber.org/zap"
)

// Logger interface para logging estruturado
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
	Sync() error
}

// zapLogger implementação com Zap
type zapLogger struct {
	logger *zap.SugaredLogger
}

// NewLogger cria uma nova instância do logger
func NewLogger() Logger {
	// Configuração para desenvolvimento (mais legível)
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)

	logger, _ := config.Build()

	return &zapLogger{
		logger: logger.Sugar(),
	}
}

// Info registra uma mensagem informativa
func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Infow(msg, fields...)
}

// Error registra uma mensagem de erro
func (l *zapLogger) Error(msg string, fields ...interface{}) {
	l.logger.Errorw(msg, fields...)
}

// Fatal registra uma mensagem fatal e encerra o programa
func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatalw(msg, fields...)
}

// Debug registra uma mensagem de debug
func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debugw(msg, fields...)
}

// Sync força a escrita de logs pendentes
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}
