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

	apexv1.Defaulted(observedScraper)
	observed.scraper = observedScraper

	return nil
}

func (o *StateObserver) observeScraper(key types.NamespacedName, scraper *apexv1.Scraper) error {
	return o.Client.Get(o.Context, key, scraper)
}
