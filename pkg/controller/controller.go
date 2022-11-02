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

	"ctx.sh/apex"
	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/scraper"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Mgr = mgr
	return ctrl.NewControllerManagedBy(mgr).
		For(&apexv1.Scraper{}).
		WithEventFilter(r.predicates()).
		Complete(r)
}

type Reconciler struct {
	Client   client.Client
	Log      logr.Logger
	Mgr      ctrl.Manager
	Scrapers *scraper.Manager
	Metrics  *apex.Metrics
}

// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=pods/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get
// +kubebuilder:rbac:groups=core,resources=endpoints,verbs=get;list;watch
// +kubebuilder:rbac:groups=core,resources=endpoints/status,verbs=get
// +kubebuilder:rbac:groups=apex.ctx.sh,resources=scrapers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apex.ctx.sh,resources=scrapers/status,verbs=get;update;patch

func (r *Reconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	handler := Handler{
		client:   r.Mgr.GetClient(),
		request:  request,
		ctx:      ctx,
		log:      r.Log.WithValues("name", request.Name, "namespace", request.Namespace),
		recorder: r.Mgr.GetEventRecorderFor("ApexOperator"),
		observed: NewObservedState(),
		scrapers: r.Scrapers,
	}
	r.Metrics.CounterInc("reconcile_total")
	return handler.reconcile(request)
}

func (r *Reconciler) predicates() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			r.Metrics.CounterInc("reconcile_update_total")
			if e.ObjectOld == nil {
				return false
			}
			if e.ObjectNew == nil {
				return false
			}

			return e.ObjectNew.GetResourceVersion() != e.ObjectOld.GetResourceVersion()
		},
		CreateFunc: func(e event.CreateEvent) bool {
			r.Metrics.CounterInc("reconcile_create_total")
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			r.Metrics.CounterInc("reconcile_delete_total")
			return true
		},
	}
}

type Handler struct {
	client   client.Client
	request  ctrl.Request
	ctx      context.Context
	log      logr.Logger
	recorder record.EventRecorder
	observed ObservedState
	scrapers *scraper.Manager
	metrics  *apex.Metrics
}

func (h *Handler) reconcile(request ctrl.Request) (ctrl.Result, error) {
	h.log.Info("request received", "request", request)

	observer := &StateObserver{
		Client:  h.client,
		Request: request,
		Context: h.ctx,
		Log:     h.log,
	}

	err := observer.observe(&h.observed)
	if err != nil {
		return ctrl.Result{}, err
	}

	scraperReconciler := &ScraperReconciler{
		client:   h.client,
		log:      h.log,
		recorder: h.recorder,
		observed: h.observed,
		scrapers: h.scrapers,
		metrics:  h.metrics,
	}

	return scraperReconciler.reconcile(h.ctx, request)
}
