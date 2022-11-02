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

package controller

import (
	"context"
	"time"

	"ctx.sh/apex"
	"ctx.sh/apex-operator/pkg/scraper"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type ScraperReconciler struct {
	client   client.Client
	log      logr.Logger
	observed ObservedState
	recorder record.EventRecorder
	scrapers *scraper.Manager
	metrics  *apex.Metrics
}

var requeueResult reconcile.Result = ctrl.Result{
	Requeue:      true,
	RequeueAfter: 30 * time.Second,
}

func (r *ScraperReconciler) reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	metrics := r.metrics.WithPrefix("reconcile").WithLabels("name", "namespace")

	timer := metrics.HistogramTimer("reconcile_seconds", request.Name, request.Namespace)
	defer timer.ObserveDuration()

	metrics.CounterInc("request_total", request.Name, request.Namespace)

	if r.observed.scraper == nil {
		metrics.CounterInc("request_cleanup_total", request.Name, request.Namespace)
		r.log.Info("the scraper has been deleted, ensuring cleanup")
		r.scrapers.Remove(request.NamespacedName)
		return ctrl.Result{}, nil
	}

	r.log.Info("reconciling scraper", "request", request)

	err := r.scrapers.Add(ctx, scraper.ScraperOpts{
		Key:     request.NamespacedName,
		Scraper: *r.observed.scraper.DeepCopy(),
		Client:  r.client,
		Log:     r.log.WithValues("scraper", request.NamespacedName),
		Metrics: r.metrics,
	})
	if err != nil {
		// Later we can get more explicit about what is a retryable
		// error in the scraper start.
		metrics.CounterInc("request_error_total", request.Name, request.Namespace)
		return requeueResult, err
	}

	metrics.CounterInc("request_success_total", request.Name, request.Namespace)
	return ctrl.Result{}, nil
}
