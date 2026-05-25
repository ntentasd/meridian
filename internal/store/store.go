package store

import "sync"

type Store struct {
	mu      sync.RWMutex
	entries map[string]RouteEntry // key: "<kind>/<namespace>/<name>"
	subs    map[chan Event]struct{}
}

// New returns an initialized Store ready for concurrent use.
func New() *Store {
	return &Store{
		entries: make(map[string]RouteEntry),
		subs:    make(map[chan Event]struct{}),
	}
}

// Upsert writes or overwrites a RouteEntry and broadcasts an upsert event to all subscribers.
func (s *Store) Upsert(e RouteEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[GetKey(e.Kind, e.Namespace, e.Name)] = e

	for ch := range s.subs {
		select {
		case ch <- Event{Kind: EventUpsert, Entry: e}:
		default:
		}
	}
}

// Delete removes the entry identified by key and broadcasts a delete event to all subscribers.
// It is a no-op if the key does not exist.
func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[key]
	if !ok {
		return
	}

	delete(s.entries, key)

	for ch := range s.subs {
		select {
		case ch <- Event{Kind: EventDelete, Entry: e}:
		default:
		}
	}
}

// List returns a consistent snapshot of all current entries, safe for use as an SSE initial push.
func (s *Store) List() []RouteEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries := make([]RouteEntry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}

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
