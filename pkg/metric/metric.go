package metric

import (
	"time"

	"ctx.sh/apex-operator/pkg/clock"
)

// ValueType represents the type of metric that has been
// collected.
type ValueType int

const (
	_ ValueType = iota
	Counter
	Gauge
	Untyped
	Summary
	Histogram
	Unknown
)

// Metric is the interface that all metrics must be adhear to
// in order to be processed through and output.  This is a bit
// overkill for now, but I may add in additional inputs later.
type Metric interface {
	// WithClock sets the clock interface for the metric
	WithClock(clock.Clock) Metric
	// WithPrefix adds a prefix to the metric.  The prefix will be
	// prepended to the name before output.
	WithPrefix(string) Metric
	// Name returns the name of the metric.  If there is a prefix,
	// the prefix is prepended to it.
	Name() string
	// Tags returns the value of the metric
	Tags() map[string]string
	// AddLabel adds a label to the map
	Time() time.Time
	// Since returns the delta between the time the metric was
	// collected and now.
	Since() time.Duration
	// Type returns a general type for the metric.
	Type() ValueType
	// SetType sets the type of the metric values
	SetType(ValueType)
	// Values returns all of the available measurements for
	// the metric.
	Values() map[string]interface{}
	// Add measurement adds a new measurement to the metric.
	AddValue(string, interface{})
	// Ack marks the metric processing as succeeded
	Ack()
	// Nack marks the metric processing as failed.
	Nack()
	// Drop marks the metric as dropped without being output.
	Drop()
}
