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

// +kubebuilder:docs-gen:collapse=Apache License
package v1

import (
	"crypto/tls"
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:docs-gen:collapse=Go imports

// SetupWebhookWithManager adds webhook for FlinkCluster.
func (s *Scraper) SetupWebhookWithManager(mgr ctrl.Manager, certDir string) error {
	if certDir != "" {
		whs := mgr.GetWebhookServer()
		whs.CertDir = certDir
		whs.TLSOpts = append(whs.TLSOpts, func(t *tls.Config) {
			t.InsecureSkipVerify = true
		})
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(s).
		Complete()
}

// +kubebuilder:webhook:admissionReviewVersions=v1,sideEffects=none,path=/mutate-apex-ctx-sh-v1-scraper,mutating=true,failurePolicy=fail,groups=apex.ctx.sh,resources=scrapers,verbs=create;update,versions=v1,name=mscraper.apex.ctx.sh

var _ webhook.Defaulter = &Scraper{}

// Default implements webhook. Defaulter so a webhook will be registered for the
// type.
func (s *Scraper) Default() {
	Defaulted(s)
	data, _ := json.Marshal(s)
	fmt.Printf("%s", string(data))
}

// +kubebuilder:webhook:admissionReviewVersions=v1,sideEffects=none,path=/validate-apex-ctx-sh-v1-scraper,mutating=false,failurePolicy=fail,groups=apex.ctx.sh,resources=scraper,verbs=create;update,versions=v1,name=vscraper.apex.ctx.sh

var _ webhook.Validator = &Scraper{}
var validator = Validator{}

// ValidateCreate implements webhook. Validator so a webhook will be registered
// for the type.
func (s *Scraper) ValidateCreate() error {
	return validator.ValidateCreate(s)
}

// ValidateUpdate implements webhook. Validator so a webhook will be registered
// for the type.
func (s *Scraper) ValidateUpdate(old runtime.Object) error {
	var oldCluster = old.(*Scraper)
	return validator.ValidateUpdate(oldCluster, s)
}

// ValidateDelete implements webhook. Validator so a webhook will be registered
// for the type.
func (s *Scraper) ValidateDelete() error {
	return validator.ValidateDelete(s)
}

// +kubebuilder:docs-gen:collapse=Validate object name
