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
	"sync"
	"time"

	"ctx.sh/apex"
	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	typesv1 "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultMaxRetryAttempts int           = 10
	DefaultMaxMultiplier    float64       = 1.5
	DefaultBackoff          time.Duration = 500 * time.Millisecond
)

type DiscoveryOpts struct {
	WorkChan chan Resource
	Scraper  apexv1.Scraper
	Client   client.Client
	Log      logr.Logger
	Metrics  *apex.Metrics
}

type Discovery struct {
	client    client.Client
	log       logr.Logger
	metrics   *apex.Metrics
	scraper   apexv1.Scraper
	workChan  chan<- Resource
	startChan chan error
	stopChan  chan struct{}
	stopOnce  sync.Once
}

func NewDiscovery(opts DiscoveryOpts) *Discovery {
	return &Discovery{
		client:    opts.Client,
		log:       opts.Log,
		metrics:   opts.Metrics,
		scraper:   opts.Scraper,
		startChan: make(chan error),
		stopChan:  make(chan struct{}),
		workChan:  opts.WorkChan,
	}
}

func (d *Discovery) Start(ctx context.Context) <-chan error {
	go func() {
		d.poll(ctx)
	}()

	return d.startChan
}

func (d *Discovery) Stop() {
	d.stopOnce.Do(func() {
		close(d.stopChan)
	})
}

func (d *Discovery) poll(ctx context.Context) {
	d.startChan <- d.intervalRun(ctx)

	interval := time.Duration(*d.scraper.Spec.ScrapeIntervalSeconds) * time.Second

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-d.stopChan:
			return
		case <-ctx.Done():
			// If we get an interrupt/kill, block until stop is called.
			<-d.stopChan
			return
		case <-ticker.C:
			_ = d.intervalRun(ctx)
		}
	}
}

func (d *Discovery) intervalRun(ctx context.Context) error {
	var discovered int = 0
	var enabled int = 0
	d.log.Info("starting discovery run")
	d.metrics.CounterInc("run_total", d.scraper.Name)
	// These could be parallel
	err := d.discoverPods(ctx, &discovered, &enabled)
	if err != nil {
		return err
	}

	err = d.discoverServices(ctx, &discovered, &enabled)
	if err != nil {
		return err
	}

	err = d.update(ctx, apexv1.ScraperStatus{
		Discovered:  int64(discovered),
		Enabled:     int64(enabled),
		LastScraped: metav1.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (d *Discovery) discoverPods(ctx context.Context, discovered *int, enabled *int) error {
	selector := labels.SelectorFromSet(d.scraper.Spec.Selector.MatchLabels)
	var list corev1.PodList
	err := d.client.List(ctx, &list, &client.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return err
	}

	for _, pod := range list.Items {
		r := FromPod(pod, d.scraper.Spec)
		if r.enabled {
			*enabled++
			d.workChan <- r
		}
	}

	*discovered += len(list.Items)
	d.metrics.GaugeSet("pods_total", float64(*discovered), d.scraper.Name)
	return nil
}

func (d *Discovery) discoverServices(ctx context.Context, discovered *int, enabled *int) error {
	selector := labels.SelectorFromSet(d.scraper.Spec.Selector.MatchLabels)
	var list corev1.ServiceList
	err := d.client.List(ctx, &list, &client.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return err
	}

	for _, svc := range list.Items {
		r := FromService(svc, d.scraper.Spec)
		if r.enabled {
			// If we are a headless service or the discovery annotation
			// is set, use the endpoints.
			if r.ip == "None" || r.discovery == "endpoints" {
				return d.discoverEndpoints(ctx, typesv1.NamespacedName{
					Namespace: svc.GetNamespace(),
					Name:      svc.GetName(),
				}, svc.ObjectMeta, discovered, enabled, svc.GetAnnotations())
			} else {
				*enabled++
				d.workChan <- r
			}
		}
	}

	*discovered += len(list.Items)
	d.metrics.GaugeSet("services_total", float64(*discovered), d.scraper.Name)
	return nil
}

func (d *Discovery) discoverEndpoints(
	ctx context.Context,
	nn typesv1.NamespacedName,
	obj metav1.ObjectMeta,
	discovered *int,
	enabled *int,
	annotations map[string]string,
) error {
	var endpoints corev1.Endpoints
	err := d.client.Get(ctx, nn, &endpoints, &client.GetOptions{})
	if err != nil {
		return err
	}

	for _, sset := range endpoints.Subsets {
		for _, addr := range sset.Addresses {
			r := FromEndpointAddress(addr, obj, annotations, d.scraper.Spec)
			// Redundant check since we only call this from the service
			// right now.
			if r.enabled {
				*enabled++
				d.workChan <- r
			}
		}
	}

	*discovered += len(endpoints.Subsets)
	d.metrics.GaugeSet("pods_total", float64(*discovered), d.scraper.Name)
	return nil
}

func (d *Discovery) update(ctx context.Context, status apexv1.ScraperStatus) error {
	var scraper = apexv1.Scraper{}
	d.scraper.DeepCopyInto(&scraper)
	scraper.Status = status
	err := d.client.Status().Update(ctx, &scraper, &client.UpdateOptions{})

	return err
}
