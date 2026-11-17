package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api/client"
)

const (
	hdrClientID  = "X-Client-Id"
	hdrRequestID = "X-Request-Id"
	testClient   = "alice"
)

func TestParseForwardHeaders(t *testing.T) {
	got, err := ParseForwardHeaders(" x-client-id, X-Request-ID ,,x-client-id ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []string{hdrClientID, hdrRequestID}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v, want %v", got, want)
	}

	if got, err := ParseForwardHeaders(""); err != nil || len(got) != 0 {
		t.Fatalf("empty input: got %v, %v", got, err)
	}

	for _, bad := range []string{"authorization", "Host", "content-type", "Connection", "sec-websocket-key", "bad header", "x:y",
		"content-range", "CONTENT-LANGUAGE", "Proxy-Authenticate", "HTTP2-Settings", "Idempotency-Key", "X-Idempotency-Key"} {
		if _, err := ParseForwardHeaders("x-ok," + bad); err == nil {
			t.Errorf("expected %q to be rejected", bad)
		}
	}
}

func TestForwardHeadersMiddleware(t *testing.T) {
	var seen http.Header
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = ForwardedHeadersFromContext(r.Context())
	})

	t.Run("empty allowlist returns handler unchanged", func(t *testing.T) {
		h := ForwardHeadersMiddleware(nil, next)
		if fmt.Sprintf("%p", h) != fmt.Sprintf("%p", next) {
			t.Fatal("expected the original handler")
		}
	})

	t.Run("copies only allowlisted headers", func(t *testing.T) {
		h := ForwardHeadersMiddleware([]string{hdrClientID, hdrRequestID}, next)
		req := httptest.NewRequest(http.MethodPost, "/block", nil)
		req.Header.Set("x-client-id", testClient)
		req.Header.Set("X-Other", "nope")
		req.Header.Set("Authorization", "Bearer caller")
		req.Header.Set(hdrRequestID, strings.Repeat("a", MaxForwardedHeaderValueLen+1))
		h.ServeHTTP(httptest.NewRecorder(), req)

		if len(seen) != 1 || seen.Get(hdrClientID) != testClient {
			t.Fatalf("unexpected forwarded headers: %v", seen)
		}
	})

	t.Run("drops headers nominated by Connection", func(t *testing.T) {
		h := ForwardHeadersMiddleware([]string{hdrClientID, hdrRequestID}, next)
		req := httptest.NewRequest(http.MethodPost, "/block", nil)
		req.Header.Add("Connection", "keep-alive")
		req.Header.Add("Connection", " x-CLIENT-id ")
		req.Header.Set(hdrClientID, testClient)
		req.Header.Set(hdrRequestID, "req-1")
		h.ServeHTTP(httptest.NewRecorder(), req)
		if len(seen) != 1 || seen.Get(hdrRequestID) != "req-1" {
			t.Fatalf("unexpected forwarded headers: %v", seen)
		}
	})

	t.Run("no matching headers leaves context untouched", func(t *testing.T) {
		h := ForwardHeadersMiddleware([]string{hdrClientID}, next)
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/block", nil))
		if seen != nil {
			t.Fatalf("expected no forwarded headers, got %v", seen)
		}
	})
}

func TestForwardHeadersTransportDoesNotMutateRequest(t *testing.T) {
	var got string
	tr := &ForwardHeadersTransport{Base: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		got = r.Header.Get(hdrClientID)
		return &http.Response{StatusCode: http.StatusOK, Body: http.NoBody}, nil
	})}

	ctx := WithForwardedHeaders(context.Background(), http.Header{hdrClientID: {testClient}})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "http://node/rpc/v1", nil)
	if _, err := tr.RoundTrip(req); err != nil {
		t.Fatal(err)
	}
	if got != testClient {
		t.Fatalf("outbound header = %q, want alice", got)
	}
	if req.Header.Get(hdrClientID) != "" {
		t.Fatal("original request was mutated")
	}
}

// End to end through the real Lotus JSON-RPC client: concurrent calls with
// different forwarded values must each reach the node with only their own
// value, and calls without forwarded headers must carry none.
func TestForwardHeadersThroughLotusClientConcurrent(t *testing.T) {
	type rpcReq struct {
		ID     json.RawMessage `json:"id"`
		Params []any           `json:"params"`
	}

	var mu sync.Mutex
	mismatches := 0
	node := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req rpcReq
		_ = json.NewDecoder(r.Body).Decode(&req)

		// StateNetworkName has no params; the expected client id is echoed
		// back as the result so the caller can compare.
		if r.Header.Get("Authorization") != "Bearer node-token" {
			mu.Lock()
			mismatches++
			mu.Unlock()
		}
		result := r.Header.Get(hdrClientID) + "|" + strings.Join(r.Header.Values(hdrRequestID), ",")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": result})
	}))
	defer node.Close()

	full, closer, err := client.NewFullNodeRPCV1(context.Background(), node.URL+"/rpc/v1",
		http.Header{"Authorization": {"Bearer node-token"}},
		jsonrpc.WithHTTPClient(NewForwardHeadersHTTPClient()))
	if err != nil {
		t.Fatal(err)
	}
	defer closer()

	var wg sync.WaitGroup
	errs := make(chan error, 200)
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ctx := context.Background()
			want := "|"
			if i%4 != 0 {
				id := fmt.Sprintf("client-%d", i)
				ctx = WithForwardedHeaders(ctx, http.Header{hdrClientID: {id}, hdrRequestID: {fmt.Sprintf("req-%d", i)}})
				want = fmt.Sprintf("%s|req-%d", id, i)
			}
			got, err := full.StateNetworkName(ctx)
			if err != nil {
				errs <- err
				return
			}
			if string(got) != want {
				errs <- fmt.Errorf("call %d: node saw %q, want %q", i, got, want)
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
	if mismatches != 0 {
		t.Errorf("%d calls lost the static Authorization header", mismatches)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
