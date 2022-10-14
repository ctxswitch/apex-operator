package scraper

import (
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
	monitor, has := m.scrapers[key]
	return monitor, has
}

func (m *Manager) Update(opts ScraperOpts) {
	key := opts.Key
	// If it exists, it's most likely been changed, so replace.
	if _, has := m.Has(key); has {
		m.Remove(key)
	}

	var scraper = NewScraper(opts)
	m.scrapers[key] = scraper
	scraper.Start()
}

func (m *Manager) Remove(key types.NamespacedName) {
	monitor, has := m.scrapers[key]
	if !has {
		return
	}

	monitor.Stop()
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
