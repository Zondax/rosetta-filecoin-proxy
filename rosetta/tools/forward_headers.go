package tools

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

// MaxForwardedHeaderValueLen caps the size of a single forwarded header value.
// Longer values are dropped (not truncated) so the node never sees partial data.
const MaxForwardedHeaderValueLen = 1024

// deniedForwardHeaders can never be forwarded, even if allowlisted. They either
// carry credentials (an incoming value must not replace LOTUS_RPC_TOKEN), are
// owned by the HTTP transport / JSON-RPC client, or are hop-by-hop headers.
var deniedForwardHeaders = map[string]struct{}{
	"Authorization":       {},
	"Proxy-Authorization": {},
	"Proxy-Authenticate":  {},
	"Cookie":              {},
	"Host":                {},
	"Accept-Encoding":     {},
	"Transfer-Encoding":   {},
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Connection":    {},
	"Te":                  {},
	"Trailer":             {},
	"Upgrade":             {},
	"Expect":              {},
	"Http2-Settings":      {},
	// Their presence lets Go's transport automatically replay the outbound POST.
	"Idempotency-Key":   {},
	"X-Idempotency-Key": {},
}

type forwardedHeadersKey struct{}

// isValidHeaderName reports whether name is an RFC 7230 token.
func isValidHeaderName(name string) bool {
	for _, c := range name {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", c):
		default:
			return false
		}
	}
	return name != ""
}

// ParseForwardHeaders parses a comma-separated allowlist of header names
// (e.g. the FORWARD_HEADERS env var). An empty input returns an empty list,
// which disables forwarding. Names are canonicalized and de-duplicated.
func ParseForwardHeaders(raw string) ([]string, error) {
	var names []string
	seen := map[string]struct{}{}

	for _, part := range strings.Split(raw, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		if !isValidHeaderName(name) {
			return nil, fmt.Errorf("invalid header name %q", name)
		}
		canonical := http.CanonicalHeaderKey(name)
		if _, denied := deniedForwardHeaders[canonical]; denied || strings.HasPrefix(canonical, "Content-") || strings.HasPrefix(canonical, "Sec-Websocket-") {
			return nil, fmt.Errorf("header %q cannot be forwarded", canonical)
		}
		if _, dup := seen[canonical]; dup {
			continue
		}
		seen[canonical] = struct{}{}
		names = append(names, canonical)
	}

	return names, nil
}

// ForwardHeadersMiddleware copies the allowlisted headers of each incoming
// request into that request's context, where ForwardHeadersTransport picks them
// up for every outbound node call made with that context. With an empty
// allowlist the handler is returned unchanged.
func ForwardHeadersMiddleware(allowlist []string, next http.Handler) http.Handler {
	if len(allowlist) == 0 {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Headers nominated by the incoming Connection header are hop-by-hop for
		// this request and must not be turned into end-to-end metadata.
		hopByHop := map[string]bool{}
		for _, value := range r.Header.Values("Connection") {
			for _, name := range strings.Split(value, ",") {
				hopByHop[http.CanonicalHeaderKey(strings.TrimSpace(name))] = true
			}
		}

		forwarded := http.Header{}
		for _, name := range allowlist {
			if hopByHop[name] {
				continue
			}
			for _, v := range r.Header.Values(name) {
				if len(v) > MaxForwardedHeaderValueLen {
					continue
				}
				forwarded.Add(name, v)
			}
		}

		if len(forwarded) > 0 {
			r = r.WithContext(WithForwardedHeaders(r.Context(), forwarded))
		}
		next.ServeHTTP(w, r)
	})
}

// WithForwardedHeaders returns a copy of ctx carrying headers to be attached to
// outbound node calls.
func WithForwardedHeaders(ctx context.Context, h http.Header) context.Context {
	return context.WithValue(ctx, forwardedHeadersKey{}, h)
}

// ForwardedHeadersFromContext returns the headers stored by WithForwardedHeaders, if any.
func ForwardedHeadersFromContext(ctx context.Context) http.Header {
	h, _ := ctx.Value(forwardedHeadersKey{}).(http.Header)
	return h
}

// ForwardHeadersTransport is an http.RoundTripper that adds the headers found in
// the outgoing request's context. Values live only in the per-request context,
// never on the shared client, so they cannot leak between incoming requests.
type ForwardHeadersTransport struct {
	Base http.RoundTripper
}

func (t *ForwardHeadersTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	h := ForwardedHeadersFromContext(req.Context())
	if len(h) == 0 {
		return base.RoundTrip(req)
	}

	// RoundTrippers must not modify the request they are given.
	out := req.Clone(req.Context())
	for name, values := range h {
		out.Header[name] = append([]string(nil), values...)
	}
	return base.RoundTrip(out)
}

// NewForwardHeadersHTTPClient returns an http.Client equivalent to go-jsonrpc's
// default client, wrapped with ForwardHeadersTransport.
func NewForwardHeadersHTTPClient() *http.Client {
	return &http.Client{
		Transport: &ForwardHeadersTransport{
			Base: &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				ForceAttemptHTTP2:     true,
				MaxIdleConns:          100,
				MaxIdleConnsPerHost:   100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}
