/*
 * Copyright 2022 Rob Lyon <rob@ctxswitch.com>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
