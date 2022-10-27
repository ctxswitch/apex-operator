package scraper

import (
	"context"
	"reflect"
	"sync"

	apexv1 "ctx.sh/apex-operator/pkg/apis/apex.ctx.sh/v1"
	"ctx.sh/apex-operator/pkg/output"
	"ctx.sh/apex-operator/pkg/output/datadog"
	"ctx.sh/apex-operator/pkg/output/logger"
	"ctx.sh/apex-operator/pkg/output/statsd"

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
	key       types.NamespacedName
	client    client.Client
	cancel    context.CancelFunc
	log       logr.Logger
	config    apexv1.ScraperSpec
	startChan chan error
	stopChan  chan struct{}
	stopOnce  sync.Once
}

func NewScraper(opts ScraperOpts) *Scraper {
	return &Scraper{
		key:       opts.Key,
		config:    opts.Config,
		client:    opts.Client,
		log:       opts.Log,
		startChan: make(chan error),
		stopChan:  make(chan struct{}),
	}
}

func (s *Scraper) Start(ctx context.Context) <-chan error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		s.up(ctx)
	}()

	return s.startChan
}

func (s *Scraper) Stop() {
	s.stopOnce.Do(func() {
		s.cancel()
	})
}

func (s *Scraper) up(ctx context.Context) {
	workers := *s.config.Workers

	workChan := make(chan Resource, workers*10)
	defer close(workChan)

	d := NewDiscovery(DiscoveryOpts{
		Client:   s.client,
		Config:   s.config,
		Log:      s.log.WithValues("name", "discovery"),
		WorkChan: workChan,
	})
	if err := <-d.Start(ctx); err != nil {
		s.startChan <- err
		return
	}
	defer d.Stop()

	outputs, err := s.initOutputs()
	if err != nil {
		s.startChan <- err
		return
	}

	var wg sync.WaitGroup
	for i := 0; i < int(workers); i++ {
		s.log.Info("starting up worker", "id", i)
		worker := NewWorker(
			workChan,
			s.config,
			s.log.WithValues("worker", i),
			outputs,
		)
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker.Start(ctx)
		}()
	}

	s.startChan <- nil

	<-ctx.Done()
	wg.Wait()
}

func (s *Scraper) initOutputs() ([]output.Output, error) {
	v := reflect.ValueOf(*s.config.Outputs)

	outputs := make([]output.Output, 0)

	for i := 0; i < v.NumField(); i++ {
		switch oo := v.Field(i).Interface().(type) {
		case *apexv1.StatsdOutput:
			if oo == nil {
				continue
			}
			out, err := statsd.New(*oo.DeepCopy())
			if err == nil {
				if *oo.Enabled {
					outputs = append(outputs, out)
				} else {
					s.log.Info("statsd output is disabled")
				}
			} else {
				s.log.Error(err, "unable to initialize statsd output")
				return nil, err
			}
		case *apexv1.LoggerOutput:
			if oo == nil {
				continue
			}
			out, err := logger.New(s.log)
			if err == nil {
				if *oo.Enabled {
					outputs = append(outputs, out)
				} else {
					s.log.Info("logger output is disabled")
				}
			} else {
				s.log.Error(err, "unable to initialize logging output")
				return nil, err
			}
		case *apexv1.DatadogOutput:
			if oo == nil {
				continue
			}
			out, err := datadog.New(*oo.DeepCopy())
			if err == nil {
				if *oo.Enabled {
					outputs = append(outputs, out)
				} else {
					s.log.Info("datadog output is disabled")
				}
			} else {
				s.log.Error(err, "unable to initialize logging output")
				return nil, err
			}
		default:
			s.log.Info("FOOOOOOOOOOOOOOOOOOOOO", "oo", oo)
		}
	}

	s.log.Info("SSSSSSSSSSSSSSSSSSSSSS", "outputs", outputs)

	return outputs, nil
}
