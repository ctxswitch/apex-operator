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
			AllowLabels:           &[]bool{false}[0],
			AnnotationPrefix:      &[]string{"prometheus.io"}[0],
			ScrapeIntervalSeconds: &[]int32{10}[0],
			Resources:             []string{"pods", "services"},
			Workers:               &[]int32{10}[0],
		},
	}

	scraper := &Scraper{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
	}
	Defaulted(scraper)

	assert.Equal(t, expected, scraper)
}
