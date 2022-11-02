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
	"reflect"
	"sync"

	"ctx.sh/apex"
	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/output"
	"ctx.sh/apex-operator/pkg/output/datadog"
	"ctx.sh/apex-operator/pkg/output/logger"
	"ctx.sh/apex-operator/pkg/output/statsd"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScraperOpts struct {
	Key     types.NamespacedName
	Scraper apexv1.Scraper
	Client  client.Client
	Context context.Context
	Log     logr.Logger
	Metrics *apex.Metrics
}

type Scraper struct {
	key       types.NamespacedName
	client    client.Client
	cancel    context.CancelFunc
	log       logr.Logger
	metrics   *apex.Metrics
	scraper   apexv1.Scraper
	startChan chan error
	stopChan  chan struct{}
	stopOnce  sync.Once
}

func NewScraper(opts ScraperOpts) *Scraper {
	return &Scraper{
		key:       opts.Key,
		scraper:   opts.Scraper,
		client:    opts.Client,
		log:       opts.Log,
		metrics:   opts.Metrics,
		startChan: make(chan error),
		stopChan:  make(chan struct{}),
	}
}

func (s *Scraper) Start(ctx context.Context) <-chan error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		s.up(ctx)
	}()

	return s.startChan
}

func (s *Scraper) Stop() {
	s.stopOnce.Do(func() {
		s.cancel()
	})
}

func (s *Scraper) up(ctx context.Context) {
	workers := *s.scraper.Spec.Workers
	workChan := make(chan Resource, workers)
	defer close(workChan)

	d := NewDiscovery(DiscoveryOpts{
		Client:   s.client,
		Scraper:  s.scraper,
		Log:      s.log.WithValues("name", "discovery"),
		Metrics:  s.metrics.WithPrefix("discovery").WithLabels("name"),
		WorkChan: workChan,
	})
	if err := <-d.Start(ctx); err != nil {
		s.startChan <- err
		return
	}
	defer d.Stop()

	outputs, err := s.initOutputs()
	if err != nil {
		s.startChan <- err
		return
	}

	var wg sync.WaitGroup
	for i := 0; i < int(workers); i++ {
		s.log.Info("starting up worker", "id", i)
		worker := NewWorker(
			d.scraper.Name,
			workChan,
			d.scraper.Spec,
			s.log.WithValues("worker", i),
			d.metrics.WithPrefix("output").WithLabels("name"),
			outputs,
		)
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker.Start(ctx)
		}()
	}

	s.startChan <- nil

	<-ctx.Done()
	wg.Wait()
}

func (s *Scraper) initOutputs() ([]output.Output, error) {
	v := reflect.ValueOf(*s.scraper.Spec.Outputs)

	outputs := make([]output.Output, 0)

	for i := 0; i < v.NumField(); i++ {
		switch oo := v.Field(i).Interface().(type) {
		case *apexv1.StatsdOutput:
			if oo == nil {
				continue
			}
			out, err := statsd.New(*oo.DeepCopy())
			if err == nil {
				if *oo.Enabled {
					outputs = append(outputs, out)
				} else {
					s.log.Info("statsd output is disabled")
				}
			} else {
				s.log.Error(err, "unable to initialize statsd output")
				return nil, err
			}
		case *apexv1.LoggerOutput:
			if oo == nil {
				continue
			}
			out, err := logger.New(s.log)
			if err == nil {
				if *oo.Enabled {
					outputs = append(outputs, out)
				} else {
					s.log.Info("logger output is disabled")
				}
			} else {
				s.log.Error(err, "unable to initialize logging output")
				return nil, err
			}
		case *apexv1.DatadogOutput:
			if oo == nil {
				continue
			}
			out, err := datadog.New(*oo.DeepCopy())
			if err == nil {
				if *oo.Enabled {
					outputs = append(outputs, out)
				} else {
					s.log.Info("datadog output is disabled")
				}
			} else {
				s.log.Error(err, "unable to initialize logging output")
				return nil, err
			}
		}
	}

	return outputs, nil
}
