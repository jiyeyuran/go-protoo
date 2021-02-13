package protoo

import (
	"github.com/bombsimon/logrusr"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
)

var NewLogger = func(name string) logr.Logger {
	return logrusr.NewLogger(logrus.New(), name)
}
