package scraper

import (
	"context"
	"sync"
	"time"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"github.com/go-logr/logr"
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
	sync.RWMutex
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
	s.RLock()
	defer s.RUnlock()

	s.log.Info("scraping targets")
	// input
	// output
}
