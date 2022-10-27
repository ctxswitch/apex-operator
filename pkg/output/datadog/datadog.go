package datadog

import (
	"time"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/metric"
	"ctx.sh/apex-operator/pkg/output"
)

type Datadog struct {
	apiKey       string
	timeout      time.Duration
	url          string
	httpUrlProxy string
	compression  string
}

func New(dd apexv1.DatadogOutput) (output.Output, error) {
	return &Datadog{
		apiKey:       "",
		timeout:      time.Minute,
		url:          "",
		httpUrlProxy: "",
		compression:  "",
	}, nil
}

func (d *Datadog) Send(m []metric.Metric) {}

func (d *Datadog) Close() {}

func (d *Datadog) Name() string { return "datadog" }
