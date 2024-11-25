package logging

import (
	"log"
	"os"
	"path/filepath"
)

type Logger struct {
	AccessLog      *log.Logger
	TransactionLog *log.Logger
	ErrorLog       *log.Logger
}

func NewLogger() (*Logger, error) {
	logDir := "internal/logging"
	accessLogPath := filepath.Join(logDir, "access.log")
	transactionLogPath := filepath.Join(logDir, "transaction.log")
	errorLogPath := filepath.Join(logDir, "error.log")

	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		return nil, err
	}

	accessFile, err := os.OpenFile(accessLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	transactionFile, err := os.OpenFile(transactionLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	errorFile, err := os.OpenFile(errorLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	accessLogger := log.New(accessFile, "ACCESS ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	transactionLogger := log.New(transactionFile, "TRANSACTION ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	errorLogger := log.New(errorFile, "ERROR ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	return &Logger{
		AccessLog:      accessLogger,
		TransactionLog: transactionLogger,
		ErrorLog:       errorLogger,
	}, nil
}

func (l *Logger) LogAccess(message string) {
	l.AccessLog.Println(message)
}

// LogTransaction writes a message to the transaction log
func (l *Logger) LogTransaction(message string) {
	l.TransactionLog.Println(message)
}

func (l *Logger) LogError(message string) {
	l.ErrorLog.Println(message)
}
