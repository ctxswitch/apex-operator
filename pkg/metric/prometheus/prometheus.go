package prometheus

import (
	"time"

	"ctx.sh/apex-operator/pkg/clock"
	"ctx.sh/apex-operator/pkg/metric"
)

type CommitType int

const (
	// Success implies that output has been successful.
	Success CommitType = iota
	// Failed implies that output has failed and is retryable.
	Failed
	// Dropped implies that the output has failed or has not
	// occurred and no retry should occur.
	Dropped
)

// Metric is used to store a scraped metric from prometheus
type Metric struct {
	name   string
	values map[string]interface{}
	tags   map[string]string
	time   time.Time
	clock  clock.Clock
	prefix string
	vtype  metric.ValueType
	commit CommitType
}

func New(t time.Time, name string, tags map[string]string) metric.Metric {
	metric := &Metric{
		name:  name,
		tags:  tags,
		time:  t,
		vtype: metric.Unknown,
		clock: clock.RealClock{},
	}

	return metric
}

func (m *Metric) WithClock(c clock.Clock) metric.Metric {
	m.clock = c
	return m
}

func (m *Metric) WithPrefix(prefix string) metric.Metric {
	m.prefix = prefix
	return m
}

func (m *Metric) Name() string {
	return m.prefix + m.name
}

func (m *Metric) Tags() map[string]string {
	return m.tags
}

func (m *Metric) Time() time.Time {
	return m.time
}

func (m *Metric) Since() time.Duration {
	return m.clock.Since(m.time)
}

func (m *Metric) Type() metric.ValueType {
	return m.vtype
}

func (m *Metric) SetType(vtype metric.ValueType) {
	m.vtype = vtype
}

func (m *Metric) Values() map[string]interface{} {
	return m.values
}

func (m *Metric) AddValue(name string, value interface{}) {
	m.values[name] = value
}

func (m *Metric) Ack() {
	m.commit = Success
}

func (m *Metric) Nack() {
	m.commit = Failed
}

func (m *Metric) Drop() {
	m.commit = Dropped
}

var _ Metric = Metric{}
