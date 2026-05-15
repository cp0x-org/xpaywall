package proxy

import (
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	x402 "github.com/x402-foundation/x402/go"
	x402http "github.com/x402-foundation/x402/go/http"

	"github.com/cp0x-org/xpaywall/xgateway/internal/logger"
	"github.com/cp0x-org/xpaywall/xgateway/internal/rules"
)

// state holds the shared, per-Server resources used across the request lifecycle:
// the rule provider, optional fallback handler, async logger client, and two
// caches keyed by inbound path / facilitator URL.
type state struct {
	provider  rules.Provider
	fallback  http.Handler
	logger    *logger.Client
	startedAt time.Time

	mu         sync.RWMutex
	routeCache map[string]*entry
	facilCache map[string]x402.FacilitatorClient

	pendingLogs sync.Map // logFingerprint → pendingLogEntry
}

func newState(provider rules.Provider, fallback http.Handler, lg *logger.Client) *state {
	return &state{
		provider:   provider,
		fallback:   fallback,
		logger:     lg,
		startedAt:  time.Now(),
		routeCache: make(map[string]*entry),
		facilCache: make(map[string]x402.FacilitatorClient),
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

// facilitator returns a cached x402 facilitator client for url, constructing
// one on first use. The cache key is the URL itself.
func (s *state) facilitator(url string) x402.FacilitatorClient {
	s.mu.RLock()
	fac, ok := s.facilCache[url]
	s.mu.RUnlock()
	if ok {
		return fac
	}
	fac = x402http.NewHTTPFacilitatorClient(&x402http.FacilitatorConfig{
		URL:        url,
		HTTPClient: FacilitatorHTTPClient(),
	})
	s.mu.Lock()
	s.facilCache[url] = fac
	s.mu.Unlock()
	return fac
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
