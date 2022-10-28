package scraper

import (
	"context"
	"net/http"
	"sync"
	"time"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/metric"
	"ctx.sh/apex-operator/pkg/output"
	"github.com/go-logr/logr"
)

const (
	DefaultTimeout = 5 * time.Second
)

type Worker struct {
	httpClient http.Client
	config     apexv1.ScraperSpec
	log        logr.Logger
	outputs    []output.Output
	workChan   <-chan Resource
	stopChan   chan struct{}
	stopOnce   sync.Once
}

func NewWorker(
	workChan <-chan Resource,
	config apexv1.ScraperSpec,
	log logr.Logger,
	outputs []output.Output,
) *Worker {
	return &Worker{
		httpClient: http.Client{
			Timeout: DefaultTimeout,
		},
		config:   config,
		workChan: workChan,
		log:      log,
		outputs:  outputs,
	}
}

func (w *Worker) Start(ctx context.Context) {
	w.poll(ctx)
}

func (w *Worker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopChan)
	})
}

func (w *Worker) poll(ctx context.Context) {
	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			<-w.stopChan
		case r := <-w.workChan:
			w.process(r)
		}
	}
}

func (w *Worker) process(r Resource) {
	if !r.enabled {
		return
	}

	m, err := w.scrape(r)
	if err != nil {
		w.log.Error(err, "unable to scrape resource", "resource", r)
		return
	}

	for _, o := range w.outputs {
		o.Send(m)
	}

	// update status?
}

func (w *Worker) scrape(r Resource) ([]metric.Metric, error) {
	input := Prometheus{
		Url:    r.URL(),
		Client: w.httpClient,
	}

	m, err := input.Get(r.tags)
	if err != nil {
		return nil, err
	}

	return m, nil
}
