/*
 * Copyright 2022 Rob Lyon <rob@ctxswitch.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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

func (p *Prometheus) Get(tags map[string]string) ([]metric.Metric, error) {
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
	m, err := p.parse(time.Now(), buf, tags)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (p *Prometheus) parse(now time.Time, buf []byte, tags map[string]string) ([]metric.Metric, error) {
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
			tags := parseLabels(m, tags)
			tags = parseLabelPairs(tags, m.GetLabel())
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

func parseLabels(m *dto.Metric, tags map[string]string) map[string]string {
	result := map[string]string{}

	for key, value := range tags {
		result[key] = value
	}

	for _, pair := range m.Label {
		result[pair.GetName()] = pair.GetValue()
	}

	return result
}

func parseLabelPairs(tags map[string]string, pairs []*dto.LabelPair) map[string]string {
	for _, pair := range pairs {
		name := pair.GetName()
		value := pair.GetValue()
		tags[name] = value
	}
	return tags
}
