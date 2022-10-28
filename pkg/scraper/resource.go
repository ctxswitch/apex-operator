package scraper

import (
	"fmt"
	"strings"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Resource struct {
	enabled   bool
	ip        string
	port      string
	path      string
	scheme    string
	discovery string
	labels    []string
	tags      map[string]string
}

func (r *Resource) Enabled() bool {
	return r.enabled
}

func (r *Resource) URL() string {
	return fmt.Sprintf("%s://%s:%s%s", r.scheme, r.ip, r.port, r.path)
}

func (r *Resource) Tags() map[string]string {
	return r.tags
}

func (r *Resource) parseTags(obj metav1.ObjectMeta) {
	r.tags = make(map[string]string)

	labels := obj.GetLabels()
	for _, name := range r.labels {
		if v, ok := labels[name]; ok {
			r.tags[name] = v
		}
	}
}

func FromService(svc corev1.Service, config apexv1.ScraperSpec) Resource {
	// two options, hit service or hit endpoints... how to do that
	resource := parseAnnotations(svc.GetAnnotations(), config)
	resource.parseTags(svc.ObjectMeta)
	resource.ip = svc.Spec.ClusterIP
	return resource
}

func FromPod(pod corev1.Pod, config apexv1.ScraperSpec) Resource {
	resource := parseAnnotations(pod.GetAnnotations(), config)
	resource.parseTags(pod.ObjectMeta)
	resource.ip = pod.Status.PodIP
	return resource
}

func FromEndpointAddress(
	address corev1.EndpointAddress,
	obj metav1.ObjectMeta,
	annotations map[string]string,
	config apexv1.ScraperSpec,
) Resource {
	resource := parseAnnotations(annotations, config)
	resource.parseTags(obj)
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
	var labels []string = []string{}

	scrapeAnnotation := fmt.Sprintf("%s/scrape", prefix)
	schemeAnnotation := fmt.Sprintf("%s/scheme", prefix)
	portAnnotation := fmt.Sprintf("%s/port", prefix)
	pathAnnotation := fmt.Sprintf("%s/path", prefix)
	discoveryAnnotation := fmt.Sprintf("%s/discovery", prefix)
	labelsAnnotations := fmt.Sprintf("%s/labels", prefix)

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

	if a, ok := annotations[labelsAnnotations]; ok {
		labels = strings.Split(a, ",")
	}

	return Resource{
		enabled:   enabled,
		scheme:    scheme,
		port:      port,
		path:      path,
		discovery: discovery,
		labels:    labels,
	}
}
