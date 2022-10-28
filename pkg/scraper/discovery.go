package scraper

import (
	"context"
	"sync"
	"time"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	typesv1 "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	DefaultMaxRetryAttempts int           = 10
	DefaultMaxMultiplier    float64       = 1.5
	DefaultBackoff          time.Duration = 500 * time.Millisecond
)

type DiscoveryOpts struct {
	WorkChan chan Resource
	Config   apexv1.ScraperSpec
	Client   client.Client
	Log      logr.Logger
}

type Discovery struct {
	client    client.Client
	log       logr.Logger
	config    apexv1.ScraperSpec
	workChan  chan<- Resource
	startChan chan error
	stopChan  chan struct{}
	stopOnce  sync.Once
}

func NewDiscovery(opts DiscoveryOpts) *Discovery {
	return &Discovery{
		client:    opts.Client,
		log:       opts.Log,
		config:    opts.Config,
		startChan: make(chan error),
		stopChan:  make(chan struct{}),
		workChan:  opts.WorkChan,
	}
}

func (d *Discovery) Start(ctx context.Context) <-chan error {
	go func() {
		d.poll(ctx)
	}()

	return d.startChan
}

func (d *Discovery) Stop() {
	d.stopOnce.Do(func() {
		close(d.stopChan)
	})
}

func (d *Discovery) poll(ctx context.Context) {
	d.startChan <- d.intervalRun(ctx)

	interval := time.Duration(*d.config.ScrapeIntervalSeconds) * time.Second

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-d.stopChan:
			return
		case <-ctx.Done():
			// If we get an interrupt/kill, block until stop is called.
			<-d.stopChan
		case <-ticker.C:
			_ = d.intervalRun(ctx)
		}
	}
}

func (d *Discovery) intervalRun(ctx context.Context) error {
	d.log.Info("starting discovery run")
	// These could be parallel
	err := d.discoverPods(ctx)
	if err != nil {
		return err
	}

	err = d.discoverServices(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (d *Discovery) discoverPods(ctx context.Context) error {
	selector := labels.SelectorFromSet(d.config.Selector.MatchLabels)
	var list corev1.PodList
	err := d.client.List(ctx, &list, &client.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return err
	}

	for _, pod := range list.Items {
		r := FromPod(pod, d.config)
		if r.enabled {
			d.workChan <- r
		}
	}
	return nil
}

func (d *Discovery) discoverServices(ctx context.Context) error {
	selector := labels.SelectorFromSet(d.config.Selector.MatchLabels)
	var list corev1.ServiceList
	err := d.client.List(ctx, &list, &client.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return err
	}

	for _, svc := range list.Items {
		r := FromService(svc, d.config)
		if r.enabled {
			// If we are a headless service or the discovery annotation
			// is set, use the endpoints.
			if r.ip == "None" || r.discovery == "endpoints" {
				return d.discoverEndpoints(ctx, typesv1.NamespacedName{
					Namespace: svc.GetNamespace(),
					Name:      svc.GetName(),
				}, svc.ObjectMeta, svc.GetAnnotations())
			} else {
				d.workChan <- r
			}
		}
	}
	return nil
}

func (d *Discovery) discoverEndpoints(
	ctx context.Context,
	nn typesv1.NamespacedName,
	obj metav1.ObjectMeta,
	annotations map[string]string,
) error {
	var endpoints corev1.Endpoints
	err := d.client.Get(ctx, nn, &endpoints, &client.GetOptions{})
	if err != nil {
		return err
	}

	for _, sset := range endpoints.Subsets {
		for _, addr := range sset.Addresses {
			r := FromEndpointAddress(addr, obj, annotations, d.config)
			// Redundant check since we only call this from the service
			// right now.
			if r.enabled {
				d.workChan <- r
			}
		}
	}
	return nil
}
