package tools

import (
	"fmt"
	"time"

	logging "github.com/ipfs/go-log"
)

var log = logging.Logger("RPC_Lotus")

type LotusRPCWrapper func()

// WrapWithTimeout executes lotusFunc but returns an error after the `timeout`.
//
// On timeout lotusFunc keeps running in its goroutine until it returns; callers
// must not read variables written by lotusFunc when an error is returned.
func WrapWithTimeout(lotusFunc LotusRPCWrapper, timeout time.Duration) error {
	// Buffered so the worker can always deliver its result and exit, even if
	// nobody is listening anymore because the timeout already fired. The channel
	// is never closed, so a late completion can't panic with "send on closed channel".
	done := make(chan struct{}, 1)

	go func() {
		lotusFunc()
		done <- struct{}{}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-done:
		log.Debug("received answer from Lotus")
		return nil
	case <-timer.C:
		log.Error("call to Lotus RPC timed out!")
		return fmt.Errorf("call to Lotus RPC timed out!")
	}
}
