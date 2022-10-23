package scraper

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/outputs/datadog"

	"github.com/DataDog/datadog-go/v5/statsd"
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
	Input   any
	Output  any
}

type Scraper struct {
	key      types.NamespacedName
	context  context.Context
	client   client.Client
	log      logr.Logger
	config   apexv1.ScraperSpec
	input    any
	output   any
	stopChan chan struct{}
	stopOnce sync.Once
}

func NewScraper(opts ScraperOpts) *Scraper {
	return &Scraper{
		key:      opts.Key,
		config:   opts.Config,
		context:  opts.Context,
		client:   opts.Client,
		input:    opts.Input,
		output:   opts.Output,
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
			s.log.Info("shutting down scraper")
			return
		case <-s.context.Done():
			s.log.Info("shutting down scraper")
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

	statsdClient, err := statsd.New("ddagent.example.svc:8125")
	if err != nil {
		log.Fatal(err)
	}
	defer statsdClient.Close()

	pods, err := s.discoverPods()
	if err != nil {
		s.log.Error(err, "pod discovery failed")
	}

	for _, pod := range pods.Items {
		log := s.log.WithValues("pod", pod.GetName()+"/"+pod.GetNamespace())

		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		annotations := pod.GetAnnotations()
		scrape := *s.config.AnnotationPrefix + "/" + "scrape"
		if a, ok := annotations[scrape]; ok && a == "true" {
			// hardcode for testing
			input := Prometheus{
				Url:    fmt.Sprintf("http://%s:%d/metrics", pod.Status.PodIP, 9000),
				Client: httpClient,
			}

			// hardcode for testing
			// output := logger.Logger{
			// 	Log: log,
			// }
			output := datadog.Datadog{
				Client: statsdClient,
			}

			m, err := input.Get()
			if err != nil {
				log.Error(err, "unable to scrape metrics")
				continue
			}

			output.Send(m)
		}
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
