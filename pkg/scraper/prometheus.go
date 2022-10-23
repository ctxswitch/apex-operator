package scraper

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"ctx.sh/apex-operator/pkg/metric"
	pmx "ctx.sh/apex-operator/pkg/metric/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

type Prometheus struct {
	Url    string
	Client http.Client
}

func (p *Prometheus) Get() ([]metric.Metric, error) {
	req, _ := http.NewRequest("GET", p.Url, nil)
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// use clock interface later.  just not sure what the interface for inputs
	// is going to fully look like yet.
	m, err := p.parse(time.Now(), buf)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (p *Prometheus) parse(now time.Time, buf []byte) ([]metric.Metric, error) {
	var parser expfmt.TextParser
	var err error

	var metrics []metric.Metric

	buf = bytes.TrimPrefix(buf, []byte("\n"))
	buffer := bytes.NewBuffer(buf)
	reader := bufio.NewReader(buffer)

	// support protobuf here as well
	metricFamilies, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	for name, mf := range metricFamilies {
		for _, m := range mf.Metric {
			// Add k8s labels and default tags if needed
			tags := parseLabels(m, nil)
			p := pmx.New(now, name, tags)

			switch mf.GetType() {
			// Parse summary metrics
			case dto.MetricType_SUMMARY:
				for _, q := range m.GetSummary().Quantile {
					if v := q.GetValue(); !math.IsNaN(v) {
						p.AddValue(fmt.Sprint(q.GetQuantile()), v)
					}
				}
			// Parse histogram metrics
			case dto.MetricType_HISTOGRAM:
				p.SetType(metric.Histogram)
				for _, b := range m.GetHistogram().Bucket {
					p.AddValue(fmt.Sprint(b.GetUpperBound()), float64(b.GetCumulativeCount()))
				}
			// Parse counter metrics
			case dto.MetricType_COUNTER:
				if v := m.GetCounter().GetValue(); !math.IsNaN(v) {
					p.SetType(metric.Counter)
					p.AddValue("counter", v)
				}
			// Parse gauge metrics
			case dto.MetricType_GAUGE:
				if v := m.GetGauge().GetValue(); !math.IsNaN(v) {
					p.SetType(metric.Gauge)
					p.AddValue("gauge", v)
				}
			// Parse untyped metrics
			case dto.MetricType_UNTYPED:
				if v := m.GetUntyped().GetValue(); !math.IsNaN(v) {
					p.SetType(metric.Untyped)
					p.AddValue("value", v)
				}
			default:
				continue
			}

			metrics = append(metrics, p)
		}
	}

	return metrics, nil
}

func parseLabels(m *dto.Metric, other map[string]string) map[string]string {
	result := map[string]string{}

	for key, value := range other {
		result[key] = value
	}

	for _, pair := range m.Label {
		result[pair.GetName()] = pair.GetValue()
	}

	return result
}
