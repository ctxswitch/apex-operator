package scraper

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/inputs/prometheus"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScraperOpts struct {
	Key     types.NamespacedName
	Config  apexv1.ScraperSpec
	Client  client.Client
	Context context.Context
	Log     logr.Logger
}

type Scraper struct {
	key      types.NamespacedName
	context  context.Context
	client   client.Client
	log      logr.Logger
	config   apexv1.ScraperSpec
	stopChan chan struct{}
	stopOnce sync.Once
}

func NewScraper(opts ScraperOpts) *Scraper {
	return &Scraper{
		key:      opts.Key,
		config:   opts.Config,
		context:  opts.Context,
		client:   opts.Client,
		log:      opts.Log,
		stopChan: make(chan struct{}),
	}
}

func (s *Scraper) Start() {
	go s.intervalRun()
}

func (s *Scraper) intervalRun() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	s.Scrape()
	for {
		select {
		case <-ticker.C:
			s.Scrape()
		case <-s.stopChan:
			s.log.Info("shutting down monitor")
			return
		case <-s.context.Done():
			s.log.Info("shutting down monitor")
			return
		}
	}
}

func (s *Scraper) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
	})
}

func (s *Scraper) Scrape() {
	// Make a client pool to get through things more quickly
	s.log.Info("scraping targets")

	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}

	pods, err := s.discoverPods()
	if err != nil {
		s.log.Error(err, "pod discovery failed")
	}

	for _, pod := range pods.Items {
		annotations := pod.GetAnnotations()
		scrape := *s.config.AnnotationPrefix + "/" + "scrape"
		if a, ok := annotations[scrape]; ok && a == "true" {
			log := s.log.WithValues("pod", pod.GetName()+"/"+pod.GetNamespace())
			log.Info("found pod")

			endpoint := prometheus.Prometheus{
				Url:    fmt.Sprintf("http://%s:%d/metrics", pod.Status.PodIP, 9000),
				Client: httpClient,
			}

			m, err := endpoint.Get()
			if err != nil {
				log.Error(err, "unable to scrape metrics")
				continue
			}

			// test
			log.Info("got metric", "metric", m)
			for _, x := range m {
				log.Info("found values", "values", x.Values())
			}
		}

		// Input (maybe on to an output channel?)
		// Output (maybe spin up dedicated output workers?)
	}

	services, err := s.discoverServices()
	if err != nil {
		s.log.Error(err, "pod discovery failed")
	}

	for _, svc := range services.Items {
		annotations := svc.GetAnnotations()
		scrape := *s.config.AnnotationPrefix + "/" + "scrape"
		if a, ok := annotations[scrape]; ok && a == "true" {
			s.log.Info("found pod", "name", svc.GetName(), "namespace", svc.GetNamespace())
		}

		// Input (maybe on to an output channel?)
		// Output (maybe spin up dedicated output workers?)
	}
}

func (s *Scraper) discoverPods() (list corev1.PodList, err error) {
	selector := labels.SelectorFromSet(s.config.Selector.MatchLabels)
	err = s.client.List(s.context, &list, &client.ListOptions{
		LabelSelector: selector,
	})
	return
}

func (s *Scraper) discoverServices() (list corev1.ServiceList, err error) {
	selector := labels.SelectorFromSet(s.config.Selector.MatchLabels)
	err = s.client.List(s.context, &list, &client.ListOptions{
		LabelSelector: selector,
	})
	return
}
