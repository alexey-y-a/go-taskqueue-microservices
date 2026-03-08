package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Init() {
	log = logrus.New()

	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,

		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	log.SetOutput(os.Stdout)

	log.SetLevel(logrus.InfoLevel)

	log.SetReportCaller(true)
}

func L() *logrus.Logger {
	if log == nil {
		Init()
	}
	return log
}

func WithComponent(component string) *logrus.Entry {
	return L().WithField("component", component)
}

func WithRequest(component, method, path, requestID string) *logrus.Entry {
	return L().WithFields(logrus.Fields{
		"component":   component,
		"http_method": method,
		"http_path":   path,
		"request_id":  requestID,
	})
}

func WithError(component string, err error) *logrus.Entry {
	return L().WithFields(logrus.Fields{
		"component": component,
		"error":     err.Error(),
	})
}

func WithFields(component string, fields logrus.Fields) *logrus.Entry {
	fields["component"] = component
	return L().WithFields(fields)
}
