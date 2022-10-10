package controller

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// var requeueResult = ctrl.Result{RequeueAfter: 10 * time.Second, Requeue: true}

type ScraperReconciler struct {
	client  client.Client
	context context.Context
	log     logr.Logger
	// observed ObservedScraperState
	// desired  DesiredScraperState
	recorder record.EventRecorder
}

func (r *ScraperReconciler) reconcile(ctx context.Context, request ctrl.Request) (ctrl.Result, error) {
	// var err error

	// if r.observed.scraper == nil {
	// 	r.log.Info("the cluster has been deleted, ensuring cleanup")
	// 	// Remove the scraper
	// 	return ctrl.Result{}, nil
	// }

	// update scraper
	// update status

	return ctrl.Result{}, nil
}
