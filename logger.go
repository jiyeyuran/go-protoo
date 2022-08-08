package protoo

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/rs/zerolog"
)

// default zerolog
var zl = zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
	w.TimeFormat = "2006-01-02 15:04:05.000000"
})).With().Caller().Timestamp().Logger().Level(zerolog.InfoLevel)

var NewLogger = func(name string) logr.Logger {
	return zerologr.New(&zl).WithName(name)
}
