package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/cp0x-org/xpaywall/xgateway/internal/logger"
	"github.com/cp0x-org/xpaywall/xgateway/internal/rules"
)

// state holds the shared, per-Server resources used across the request
// lifecycle: the rule provider, optional fallback handler, async logger
// client, the registered payment protocols, and the route cache.
type state struct {
	provider  rules.Provider
	fallback  http.Handler
	logger    *logger.Client
	startedAt time.Time

	// protocols is the ordered list of payment protocols this Server knows
	// about. Order is significant — it defines selection precedence.
	protocols []protocolEntry

	mu         sync.RWMutex
	routeCache map[string]*entry

	pendingLogs sync.Map // logFingerprint → pendingLogEntry
}

func newState(provider rules.Provider, fallback http.Handler, lg *logger.Client) *state {
	return &state{
		provider:   provider,
		fallback:   fallback,
		logger:     lg,
		startedAt:  time.Now(),
		protocols:  defaultProtocols(newFacilitatorCache()),
		routeCache: make(map[string]*entry),
	}
}

func (s *state) getRoute(path string) (*entry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.routeCache[path]
	return e, ok
}

func (s *state) putRoute(path string, e *entry) {
	s.mu.Lock()
	s.routeCache[path] = e
	s.mu.Unlock()
}

// pendingLogEntry tracks a 402-response log entry waiting for its payment retry.
type pendingLogEntry struct {
	logID     uuid.UUID
	expiresAt time.Time
}

// logFingerprint returns a key that correlates the initial 402 request with its
// payment retry. Uses method+path+clientIP — reliable for single-client flows.
func logFingerprint(method, path, clientIP string) string {
	return method + "|" + path + "|" + clientIP
}
