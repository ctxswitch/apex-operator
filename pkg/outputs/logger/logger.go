package logger

import (
	"ctx.sh/apex-operator/pkg/metric"
	"github.com/go-logr/logr"
)

type Logger struct {
	Log logr.Logger
}

func (l *Logger) Send(m []metric.Metric) {
	for _, x := range m {
		l.Log.Info("metric", "values", x.Values())
	}
}
