package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// defaulted sets the scraper resource defaults
func defaulted(scraper *Scraper) {
	defaultedSpec(&scraper.Spec)
}

func defaultedSpec(spec *ScraperSpec) {
	if spec.AnnotationPrefix == nil {
		spec.AnnotationPrefix = new(string)
		*spec.AnnotationPrefix = "prometheus.io"
	}

	if spec.ScrapeIntervalSeconds == nil {
		spec.ScrapeIntervalSeconds = new(int32)
		*spec.ScrapeIntervalSeconds = 10
	}

	if spec.Selector == nil {
		spec.Selector = &metav1.LabelSelector{
			MatchLabels:      make(map[string]string),
			MatchExpressions: make([]metav1.LabelSelectorRequirement, 0),
		}
	}
}
