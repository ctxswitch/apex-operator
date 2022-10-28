package scraper

import (
	"fmt"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	corev1 "k8s.io/api/core/v1"
)

type Resource struct {
	enabled   bool
	ip        string
	port      string
	path      string
	scheme    string
	discovery string
}

func (r *Resource) Enabled() bool {
	return r.enabled
}

func (r *Resource) URL() string {
	return fmt.Sprintf("%s://%s:%s%s", r.scheme, r.ip, r.port, r.path)
}

func FromService(svc corev1.Service, config apexv1.ScraperSpec) Resource {
	// two options, hit service or hit endpoints... how to do that
	resource := parseAnnotations(svc.GetAnnotations(), config)
	resource.ip = svc.Spec.ClusterIP
	return resource
}

func FromPod(pod corev1.Pod, config apexv1.ScraperSpec) Resource {
	resource := parseAnnotations(pod.GetAnnotations(), config)
	resource.ip = pod.Status.PodIP
	return resource
}

func FromEndpointAddress(
	address corev1.EndpointAddress,
	annotations map[string]string,
	config apexv1.ScraperSpec,
) Resource {
	resource := parseAnnotations(annotations, config)
	resource.ip = address.IP
	return resource
}

func parseAnnotations(annotations map[string]string, config apexv1.ScraperSpec) Resource {
	prefix := *config.AnnotationPrefix

	var enabled bool = false
	var scheme string = "http"
	var port string = "9090"
	var path string = "/metrics"
	var discovery string = "self"

	scrapeAnnotation := fmt.Sprintf("%s/scrape", prefix)
	schemeAnnotation := fmt.Sprintf("%s/scheme", prefix)
	portAnnotation := fmt.Sprintf("%s/port", prefix)
	pathAnnotation := fmt.Sprintf("%s/path", prefix)
	discoveryAnnotation := fmt.Sprintf("%s/discovery", prefix)

	if a, ok := annotations[scrapeAnnotation]; ok {
		enabled = a == "true"
	}

	if a, ok := annotations[schemeAnnotation]; ok {
		scheme = a
	}

	if a, ok := annotations[portAnnotation]; ok {
		port = a
	}

	if a, ok := annotations[pathAnnotation]; ok {
		path = a
	}

	if a, ok := annotations[discoveryAnnotation]; ok {
		discovery = a
	}

	return Resource{
		enabled:   enabled,
		scheme:    scheme,
		port:      port,
		path:      path,
		discovery: discovery,
	}
}
