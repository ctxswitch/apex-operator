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
	"context"
	"net/http"
	"sync"
	"time"

	"ctx.sh/apex"
	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/metric"
	"ctx.sh/apex-operator/pkg/output"
	"github.com/go-logr/logr"
)

const (
	DefaultTimeout = 5 * time.Second
)

type Worker struct {
	name       string
	httpClient http.Client
	config     apexv1.ScraperSpec
	log        logr.Logger
	metrics    *apex.Metrics
	outputs    []output.Output
	workChan   <-chan Resource
	stopChan   chan struct{}
	stopOnce   sync.Once
}

func NewWorker(
	name string,
	workChan <-chan Resource,
	config apexv1.ScraperSpec,
	log logr.Logger,
	metrics *apex.Metrics,
	outputs []output.Output,
) *Worker {
	return &Worker{
		name: name,
		httpClient: http.Client{
			Timeout: DefaultTimeout,
		},
		config:   config,
		workChan: workChan,
		log:      log,
		metrics:  metrics,
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
	w.metrics.CounterInc("processed_total", w.name)
	if !r.enabled {
		return
	}

	m, err := w.scrape(r)
	if err != nil {
		w.log.Error(err, "unable to scrape resource", "resource", r)
		return
	}

	for _, o := range w.outputs {
		// Come back through here and implement metrics in outputs.
		o.Send(m)
	}
}

func (w *Worker) scrape(r Resource) ([]metric.Metric, error) {
	input := Prometheus{
		Url:    r.URL(),
		Client: w.httpClient,
	}

	w.metrics.CounterInc("scraped_total", w.name)
	m, err := input.Get(r.tags)
	if err != nil {
		w.metrics.CounterInc("scraped_error_total", w.name)
		return nil, err
	}

	w.metrics.CounterInc("scraped_success_total", w.name)
	return m, nil
}
