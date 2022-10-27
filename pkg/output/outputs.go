package output

import "ctx.sh/apex-operator/pkg/metric"

type Output interface {
	Send([]metric.Metric)
	Close()
	Name() string
}
