package statsd

import (
	"fmt"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/metric"
	"ctx.sh/apex-operator/pkg/output"
	"github.com/DataDog/datadog-go/v5/statsd"
)

type Statsd struct {
	client *statsd.Client
}

func New(cfg apexv1.StatsdOutput) (output.Output, error) {
	s, err := statsd.New(fmt.Sprintf("%s:%d", *cfg.Host, *cfg.Port))
	if err != nil {
		return nil, err
	}

	return &Statsd{
		client: s,
	}, nil
}

func (s *Statsd) Send(m []metric.Metric) {
	for _, mm := range m {
		tags := makeTags(mm.Tags())
		for _, v := range mm.Values() {
			switch mm.Type() {
			case metric.Counter:
				_ = s.client.Count(mm.Name(), int64(v.(float64)), tags, 1)
			case metric.Gauge:
				_ = s.client.Gauge(mm.Name(), v.(float64), tags, 1)
			case metric.Histogram:
				_ = s.client.Histogram(mm.Name(), v.(float64), tags, 1)
			case metric.Summary:
				_ = s.client.Distribution(mm.Name(), v.(float64), tags, 1)
			case metric.Unknown:
				_ = s.client.Gauge(mm.Name(), v.(float64), tags, 1)
			case metric.Untyped:
				_ = s.client.Gauge(mm.Name(), v.(float64), tags, 1)
			}
		}
	}
}

func (s *Statsd) Close() {
	s.client.Close()
}

func (s *Statsd) Name() string { return "statsd" }

func makeTags(t map[string]string) []string {
	tags := make([]string, len(t))
	for n, v := range t {
		tags = append(tags, n+":"+v)
	}
	return tags
}