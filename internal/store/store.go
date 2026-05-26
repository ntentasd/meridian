package store

import (
	"sort"
	"sync"
)

type Store struct {
	mu            sync.RWMutex
	entries       map[string]RouteEntry // key: Hostname
	resourceHosts map[string][]string   // key: "<kind>/<namespace>/<name>" -> []Hostname
	subs          map[chan Event]struct{}
}

// New returns an initialized Store ready for concurrent use.
func New() *Store {
	return &Store{
		entries:       make(map[string]RouteEntry),
		resourceHosts: make(map[string][]string),
		subs:          make(map[chan Event]struct{}),
	}
}

// Sync updates the store with the latest entries for a given resource.
func (s *Store) Sync(resourceKey string, entries []RouteEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldHosts := s.resourceHosts[resourceKey]
	newHosts := make(map[string]bool)

	for _, e := range entries {
		newHosts[e.Hostname] = true
		s.entries[e.Hostname] = e
	}

	for _, h := range oldHosts {
		if !newHosts[h] {
			delete(s.entries, h)
		}
	}

	hosts := make([]string, 0, len(newHosts))
	for h := range newHosts {
		hosts = append(hosts, h)
	}
	s.resourceHosts[resourceKey] = hosts

	for ch := range s.subs {
		select {
		case ch <- Event{Kind: EventUpsert}:
		default:
		}
	}
}

// Delete removes all entries associated with the given resource key.
func (s *Store) Delete(resourceKey string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	hosts, ok := s.resourceHosts[resourceKey]
	if !ok {
		return
	}

	for _, h := range hosts {
		delete(s.entries, h)
	}
	delete(s.resourceHosts, resourceKey)

	for ch := range s.subs {
		select {
		case ch <- Event{Kind: EventDelete}:
		default:
		}
	}
}

// List returns a consistent snapshot of all current entries.
func (s *Store) List() []RouteEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]RouteEntry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Name != entries[j].Name {
			return entries[i].Name < entries[j].Name
		}
		return entries[i].Hostname < entries[j].Hostname
	})

	return entries
}

// Subscribe registers a new subscriber and returns a buffered channel that receives future events.
func (s *Store) Subscribe() chan Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	ch := make(chan Event, 16)
	s.subs[ch] = struct{}{}
	return ch
}

// Unsubscribe removes the channel from the subscriber set and closes it.
func (s *Store) Unsubscribe(ch chan Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subs, ch)
	close(ch)
}
