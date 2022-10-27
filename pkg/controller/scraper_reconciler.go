package controller

import (
	"context"
	"time"

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
}

var requeueResult reconcile.Result = ctrl.Result{
	Requeue:      true,
	RequeueAfter: 30 * time.Second,
}

func (r *ScraperReconciler) reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	if r.observed.scraper == nil {
		r.log.Info("the scraper has been deleted, ensuring cleanup")
		r.scrapers.Remove(request.NamespacedName)
		return ctrl.Result{}, nil
	}

	r.log.Info("reconciling scraper", "request", request)

	err := r.scrapers.Add(ctx, scraper.ScraperOpts{
		Key:    request.NamespacedName,
		Config: *r.observed.scraper.Spec.DeepCopy(),
		Client: r.client,
		Log:    r.log.WithValues("scraper", request.NamespacedName),
	})
	if err != nil {
		// Later we can get more explicit about what is a retryable
		// error in the scraper start.
		return requeueResult, err
	}

	return ctrl.Result{}, nil
}
