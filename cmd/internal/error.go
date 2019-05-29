package internal

import (
	"fmt"
	"os"
)

type (
	// Logger is based on the standard logger proposals discussed in detail, linked below
	// https://docs.google.com/document/d/1oTjtY49y8iSxmM9YBaz2NrZIlaXtQsq3nQMd-E0HtwM/edit#
	Logger interface {
		// Log is a flexible log function described in the standard logger proposals.
		Log(...interface{}) error
	}

	noOpLogger struct{}
)

var _logger Logger = &noOpLogger{}

func (*noOpLogger) Log(_ ...interface{}) error {
	return nil
}

// SetLogger sets a Logger interface for sigctx to use.
func SetLogger(logger Logger) {
	if logger != nil {
		_logger = logger
	}
}

// Fatal will log using the configured logger and exit the application.
// Note: if err is nil this will result in a no-op (instead of killing
// the application).
func Fatal(err error) {
	if err == nil {
		return
	}

	log(err.Error())
	os.Exit(1)
}

func log(v ...interface{}) {
	if err := _logger.Log(v...); err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
}
