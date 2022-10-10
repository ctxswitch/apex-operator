package metric

import (
	"time"

	"ctx.sh/apex-operator/pkg/clock"
)

type CommitType int

const (
	// Success implies that output has been successful.
	Success CommitType = iota
	// Failed implies that output has failed and is retryable.
	Failed
	// Dropped implies that the output has failed or has not
	// occured and no retry should occur.
	Dropped
)

// PrometheusMetric is used to store a scraped metric from prometheus
type PrometheusMetric struct {
	name         string
	measurements map[string]interface{}
	labels       map[string]string
	time         time.Time
	clock        clock.Clock
	prefix       string
	vtype        ValueType
	commit       CommitType
}

func (m *PrometheusMetric) New(t time.Time, name string, labels map[string]string) Metric {
	metric := &PrometheusMetric{
		name:   name,
		labels: labels,
		time:   t,
		clock:  clock.RealClock{},
	}

	return metric
}

func (m *PrometheusMetric) WithClock(c clock.Clock) Metric {
	m.clock = c
	return m
}

func (m *PrometheusMetric) WithPrefix(prefix string) Metric {
	m.prefix = prefix
	return m
}

func (m *PrometheusMetric) Name() string {
	return m.prefix + m.name
}

func (m *PrometheusMetric) Labels() map[string]string {
	return m.labels
}

func (m *PrometheusMetric) Time() time.Time {
	return m.time
}

func (m *PrometheusMetric) Since() time.Duration {
	return m.clock.Since(m.time)
}

func (m *PrometheusMetric) Type() ValueType {
	return m.vtype
}

func (m *PrometheusMetric) Measurements() map[string]interface{} {
	return m.measurements
}

func (m *PrometheusMetric) AddMeasurement(name string, value interface{}) {
	m.measurements[name] = value
}

func (m *PrometheusMetric) Ack() {
	m.commit = Success
}

func (m *PrometheusMetric) Nack() {
	m.commit = Failed
}

func (m *PrometheusMetric) Drop() {
	m.commit = Dropped
}

var _ Metric = &PrometheusMetric{}
