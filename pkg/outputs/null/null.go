package null

import (
	"ctx.sh/apex-operator/pkg/metric"
)

type Null struct{}

func (n *Null) Send(m metric.Metric) {}
