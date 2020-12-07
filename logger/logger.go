package logger

import "github.com/Gumkle/consoler/consoler"

var logger *consoler.Logger

func InitLogger() {
	logger = consoler.NewLogger()
}

func Get() *consoler.Logger {
	return logger
}
