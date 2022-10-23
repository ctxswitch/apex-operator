package datadog

import (
	"ctx.sh/apex-operator/pkg/metric"
	"github.com/DataDog/datadog-go/v5/statsd"
)

type Datadog struct {
	Client *statsd.Client
}

func (d *Datadog) Send(m []metric.Metric) {
	for _, mm := range m {
		tags := makeTags(mm.Tags())
		for _, v := range mm.Values() {
			switch mm.Type() {
			case metric.Counter:
				_ = d.Client.Count(mm.Name(), int64(v.(float64)), tags, 1)
			case metric.Gauge:
				_ = d.Client.Gauge(mm.Name(), v.(float64), tags, 1)
			case metric.Histogram:
				_ = d.Client.Histogram(mm.Name(), v.(float64), tags, 1)
			case metric.Summary:
				_ = d.Client.Distribution(mm.Name(), v.(float64), tags, 1)
			case metric.Unknown:
				_ = d.Client.Gauge(mm.Name(), v.(float64), tags, 1)
			case metric.Untyped:
				_ = d.Client.Gauge(mm.Name(), v.(float64), tags, 1)
			}
		}
	}
}

func makeTags(t map[string]string) []string {
	tags := make([]string, len(t))
	for n, v := range t {
		tags = append(tags, n+":"+v)
	}
	return tags
}
