package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDefaulted(t *testing.T) {
	expected := &Scraper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: ScraperSpec{
			AnnotationPrefix:      &[]string{"prometheus.io"}[0],
			ScrapeIntervalSeconds: &[]int32{10}[0],
			Selector: &metav1.LabelSelector{
				MatchLabels:      make(map[string]string),
				MatchExpressions: make([]metav1.LabelSelectorRequirement, 0),
			},
		},
	}

	scraper := &Scraper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	defaulted(scraper)

	assert.Equal(t, expected, scraper)
}
