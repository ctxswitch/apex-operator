package controller

import (
	"context"

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
}

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
	return handler.reconcile(request)
}

func (r *Reconciler) predicates() predicate.Funcs {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == nil {
				return false
			}
			if e.ObjectNew == nil {
				return false
			}

			// Ignore metadata and status updates
			if e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() {
				return true
			}
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
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
		context:  h.ctx,
		log:      h.log,
		recorder: h.recorder,
		observed: h.observed,
		scrapers: h.scrapers,
	}

	return scraperReconciler.reconcile(h.ctx, request)
}
