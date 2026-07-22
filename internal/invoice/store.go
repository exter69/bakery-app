package invoice

import "sync"

// Store provides in-memory storage for generated invoices.
// Can be replaced with S3/disk-backed storage in production.
type Store struct {
	mu       sync.RWMutex
	invoices map[string]string // orderID → HTML content
}

// NewStore creates a new empty invoice store.
func NewStore() *Store {
	return &Store{invoices: make(map[string]string)}
}

// Save stores an invoice HTML for the given order ID.
func (s *Store) Save(orderID string, html string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invoices[orderID] = html
}

// Get retrieves a stored invoice by order ID.
// Returns the HTML and true if found, or empty string and false if not.
func (s *Store) Get(orderID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	html, ok := s.invoices[orderID]
	return html, ok
}
