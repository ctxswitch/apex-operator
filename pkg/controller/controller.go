package controller

import (
	"context"

	apexv1 "ctx.sh/apex/pkg/apis/apex.ctx.sh/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ApexReconciler struct {
	Client client.Client
	Log    logr.Logger
	Mgr    ctrl.Manager
	// Scrapers
}

func (r *ApexReconciler) Reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	handler := ApexHandler{
		client:   r.Mgr.GetClient(),
		request:  request,
		ctx:      ctx,
		log:      r.Log.WithValues("name", request.Name, "namespace", request.Namespace),
		recorder: r.Mgr.GetEventRecorderFor("ApexClusterOperator"),
		// observed: NewObservedApexState(),
	}
	return handler.reconcile(request)
}

func (r *ApexReconciler) SetupWithManager(mgr ctrl.Manager) error {
	predicates := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			if e.ObjectOld == nil {
				r.Log.Error(nil, "Update event has no old object to update", "event", e)
				return false
			}
			if e.ObjectNew == nil {
				r.Log.Error(nil, "Update event has no new object for update", "event", e)
				return false
			}

			// Ignore metadata and status updates
			if e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration() {
				r.Log.Info("update", "old", e.ObjectOld, "new", e.ObjectNew)
				return true
			}
			return false
		},
		CreateFunc: func(e event.CreateEvent) bool {
			r.Log.Info("create")
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			r.Log.Info("delete")
			return true
		},
	}

	r.Mgr = mgr
	return ctrl.NewControllerManagedBy(mgr).
		For(&apexv1.Scraper{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		WithEventFilter(predicates).
		Complete(r)
}

type ApexHandler struct {
	client   client.Client
	request  ctrl.Request
	ctx      context.Context
	log      logr.Logger
	recorder record.EventRecorder
	// observed ObservedApexState
	// desired  DesiredApexState
}

func (h *ApexHandler) reconcile(request ctrl.Request) (ctrl.Result, error) {
	h.log.Info("request received", "request", request)

	// Set up observer
	// Get desired state

	scraperReconciler := &ScraperReconciler{
		client:   h.client,
		context:  h.ctx,
		log:      h.log,
		recorder: h.recorder,
		// observed: h.observed,
		// desired:  h.desired,
	}

	return scraperReconciler.reconcile(h.ctx, request)
}
