package controller

import (
	"context"

	"ctx.sh/apex-operator/pkg/scraper"
	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScraperReconciler struct {
	client   client.Client
	log      logr.Logger
	observed ObservedState
	recorder record.EventRecorder
	scrapers *scraper.Manager
}

func (r *ScraperReconciler) reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	if r.observed.scraper == nil {
		r.log.Info("the cluster has been deleted, ensuring cleanup")
		r.scrapers.Remove(request.NamespacedName)
		return ctrl.Result{}, nil
	}

	r.log.Info("reconciling scraper", "request", request)
	r.scrapers.Update(scraper.ScraperOpts{
		Key:     request.NamespacedName,
		Config:  r.observed.scraper.Spec,
		Client:  r.client,
		Context: ctx,
		Log:     r.log.WithValues("scraper", request.NamespacedName),
	})

	return ctrl.Result{}, nil
}
