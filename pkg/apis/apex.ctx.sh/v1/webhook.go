// +kubebuilder:docs-gen:collapse=Apache License

package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:docs-gen:collapse=Go imports

// SetupWebhookWithManager adds webhook for FlinkCluster.
func (s *Scraper) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(s).
		Complete()
}

/*
Kubebuilder markers to generate webhook manifests.
This marker is responsible for generating a mutating webhook manifest.
The meaning of each marker can be found [here](/reference/markers/webhook.md).
*/

// +kubebuilder:webhook:admissionReviewVersions=v1,sideEffects=none,path=/mutate-apex-ctx-sh-v1-scraper,mutating=true,failurePolicy=fail,groups=apex.ctx.sh,resources=scrapers,verbs=create;update,versions=v1,name=scraper.apex.ctx.sh

var _ webhook.Defaulter = &Scraper{}

// Default implements webhook. Defaulter so a webhook will be registered for the
// type.
func (s *Scraper) Default() {
	defaulted(s)
}

// +kubebuilder:webhook:admissionReviewVersions=v1,sideEffects=none,path=/validate-apex-ctx-sh-v1-scraper,mutating=false,failurePolicy=fail,groups=apex.ctx.sh,resources=scraper,verbs=create;update,versions=v1,name=scraper.apex.ctx.sh

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
