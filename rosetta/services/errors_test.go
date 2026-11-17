package services

import (
	"errors"
	"sync"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/asserter"
)

// Regression: BuildError used to write details into the shared Err* values.
// After one failed node call /network/options listed errors with details,
// which rosetta-cli (and any client using the SDK asserter) rejects.
func TestBuildErrorDoesNotMutateSharedErrors(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got := BuildError(ErrUnableToGetTrace, errors.New("connection timed out"), true)
			if got == ErrUnableToGetTrace {
				t.Error("BuildError returned the shared error value")
			}
			if got.Details[LotusErrKey] != "connection timed out" {
				t.Errorf("details not set on returned error: %v", got.Details)
			}
			if got.Code != ErrUnableToGetTrace.Code || got.Message != ErrUnableToGetTrace.Message {
				t.Errorf("returned error lost its code/message: %+v", got)
			}
		}()
	}
	wg.Wait()

	for _, e := range ErrorList {
		if e.Details != nil {
			t.Errorf("shared error %d (%s) was modified: details %v", e.Code, e.Message, e.Details)
		}
	}
	if err := asserter.Errors(ErrorList); err != nil {
		t.Fatalf("/network/options error list is no longer valid: %v", err)
	}
}
