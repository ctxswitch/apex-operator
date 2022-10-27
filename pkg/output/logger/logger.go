package logger

import (
	"ctx.sh/apex-operator/pkg/metric"
	"ctx.sh/apex-operator/pkg/output"
	"github.com/go-logr/logr"
)

type Logger struct {
	log logr.Logger
}

func New(logger logr.Logger) (output.Output, error) {
	return &Logger{
		log: logger,
	}, nil
}

func (l *Logger) Send(m []metric.Metric) {
	for _, x := range m {
		l.log.Info("metric", "values", x.Values())
	}
}

func (l *Logger) Close() {}

func (l *Logger) Name() string { return "logger" }
