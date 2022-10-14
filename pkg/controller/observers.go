package controller

import (
	"context"
	"time"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObservedState struct {
	scraper     *apexv1.Scraper
	observeTime time.Time
}

func NewObservedState() ObservedState {
	return ObservedState{
		scraper:     nil,
		observeTime: time.Now(),
	}
}

type StateObserver struct {
	Client  client.Client
	Request ctrl.Request
	Context context.Context
	Log     logr.Logger
}

func (o *StateObserver) observe(observed *ObservedState) error {
	var err error
	var log = o.Log.WithValues("func", "observe")

	var observedScraper = new(apexv1.Scraper)
	err = o.observeScraper(o.Request.NamespacedName, observedScraper)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to get scraper")
			return err
		}
		log.V(6).Info("scraper", "state", "nil")
		return nil
	}
	observed.scraper = observedScraper

	return nil
}

func (o *StateObserver) observeScraper(key types.NamespacedName, scraper *apexv1.Scraper) error {
	return o.Client.Get(o.Context, key, scraper)
}
