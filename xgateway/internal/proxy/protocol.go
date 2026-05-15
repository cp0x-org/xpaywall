package proxy

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/cp0x-org/xpaywall/xgateway/internal/rules"
)

// PaymentProtocol abstracts a single payment scheme (x402, MPP, …) for one
// route. Implementations verify any payment proof carried on the incoming
// request and either let the chain continue to the upstream or abort with a
// 402-style challenge response.
//
// Selection rule (applied by entry.runPayment when more than one protocol is
// configured for a route): the first protocol whose HasClientAuth returns true
// is chosen; otherwise the first protocol in the entry (the default) is used.
type PaymentProtocol interface {
	// Name identifies the protocol; the value must match the corresponding
	// rules.PaymentChannel.Protocol string.
	Name() string

	// HasClientAuth reports whether the request signals that the client is
	// paying via this protocol (typically by inspecting a protocol-specific
	// header). The proxy uses this signal to pick a protocol when more than
	// one is configured for the same route.
	HasClientAuth(c *gin.Context) bool

	// Handle runs the protocol's middleware. It must either:
	//   - call c.Next() to let the request continue to the upstream, or
	//   - call c.Abort() after writing a 402-style challenge response.
	//
	// decorators are invoked against the response headers when the protocol
	// emits a challenge response — that is the standard way sibling protocols
	// are advertised. Implementations may ignore decorators if their underlying
	// library does not expose a hook for header mutation.
	Handle(c *gin.Context, decorators ...func(http.Header))

	// Challenge writes this protocol's payment-required metadata into the
	// supplied header set. Used by other protocols as a Handle decorator so
	// that a 402 response advertises this protocol as an alternative. Protocols
	// that cannot express their challenge purely through response headers may
	// leave headers unchanged.
	Challenge(ctx context.Context, headers http.Header)
}

// ProtocolFactory builds a PaymentProtocol from the subset of a rule's payment
// channels that belong to one protocol family.
//
// Returning (nil, nil) means the protocol opts out for this rule (e.g., no
// usable channel config); the rule still works without it. Returning a non-nil
// error fails entry construction and the request gets a 500.
type ProtocolFactory func(
	ctx context.Context,
	rule *rules.Rule,
	reqPath string,
	channels []*rules.PaymentChannel,
) (PaymentProtocol, error)

// protocolEntry pairs a protocol name with its factory. A Server stores a list
// of these in selection order — first entry is the default fallback.
type protocolEntry struct {
	name    string
	factory ProtocolFactory
}

// defaultProtocols returns the built-in protocol set for a new Server, in the
// order they should be considered. The first entry is the default protocol
// (used when no protocol's HasClientAuth matches the request).
//
// To add a new built-in protocol: create protocol_<name>.go with a struct
// implementing PaymentProtocol and a factory constructor, then append the
// entry here in the desired selection position.
func defaultProtocols(cache *facilitatorCache) []protocolEntry {
	return []protocolEntry{
		{name: protoX402, factory: newX402Factory(cache)},
		{name: protoMPP, factory: newMPPFactory()},
	}
}
