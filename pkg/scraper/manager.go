package scraper

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/types"
)

type Manager struct {
	scrapers map[types.NamespacedName]*Scraper
	sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		scrapers: make(map[types.NamespacedName]*Scraper),
	}
}

func (m *Manager) Has(key types.NamespacedName) (*Scraper, bool) {
	sc, has := m.scrapers[key]
	return sc, has
}

func (m *Manager) Add(ctx context.Context, opts ScraperOpts) error {
	key := opts.Key
	// If it exists, it's most likely been changed, so replace.
	if _, has := m.Has(key); has {
		m.Remove(key)
	}

	var scraper = NewScraper(opts)
	if err := <-scraper.Start(ctx); err != nil {
		return err
	}

	m.scrapers[key] = scraper
	return nil
}

func (m *Manager) Remove(key types.NamespacedName) {
	scraper, has := m.scrapers[key]
	if !has {
		return
	}

	scraper.Stop()
	delete(m.scrapers, key)
}

func (m *Manager) Stop() {
	m.Lock()
	defer m.Unlock()

	for i, g := range m.scrapers {
		g.Stop()
		delete(m.scrapers, i)
	}
}
