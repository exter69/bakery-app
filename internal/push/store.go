package push

import "sync"

// Store provides thread-safe in-memory storage for push subscriptions.
// Subscriptions are grouped by user ID (one user can have multiple devices).
type Store struct {
	mu   sync.RWMutex
	subs map[string][]Subscription // userID → subscriptions
}

// NewStore creates a new empty push subscription store.
func NewStore() *Store {
	return &Store{
		subs: make(map[string][]Subscription),
	}
}

// Save stores a push subscription for the user. If a subscription with
// the same endpoint already exists for the user, it is replaced.
func (s *Store) Save(sub Subscription) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.subs[sub.UserID]
	for i, e := range existing {
		if e.Endpoint == sub.Endpoint {
			// Replace existing subscription for same endpoint (keys may have rotated)
			existing[i] = sub
			return
		}
	}
	s.subs[sub.UserID] = append(existing, sub)
}

// GetByUser returns all subscriptions for a given user ID.
// Returns nil if the user has no subscriptions.
func (s *Store) GetByUser(userID string) []Subscription {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.subs[userID]
}

// Delete removes a subscription by its ID, scanning all users.
func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for userID, subs := range s.subs {
		for i, sub := range subs {
			if sub.ID == id {
				s.subs[userID] = append(subs[:i], subs[i+1:]...)
				if len(s.subs[userID]) == 0 {
					delete(s.subs, userID)
				}
				return
			}
		}
	}
}

// DeleteByEndpoint removes a subscription by endpoint for a specific user.
func (s *Store) DeleteByEndpoint(userID, endpoint string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	subs := s.subs[userID]
	for i, sub := range subs {
		if sub.Endpoint == endpoint {
			s.subs[userID] = append(subs[:i], subs[i+1:]...)
			if len(s.subs[userID]) == 0 {
				delete(s.subs, userID)
			}
			return
		}
	}
}

// DeleteByUser removes all push subscriptions for a given user ID.
func (s *Store) DeleteByUser(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subs, userID)
}
