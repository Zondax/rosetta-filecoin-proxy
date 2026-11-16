package tools

import (
	"testing"
	"testing/synctest"
	"time"

	"gotest.tools/assert"
)

func TestTimeout(t *testing.T) {
	// wrap the implementation and use variable capture
	impl := func() {
		time.Sleep(2 * time.Second)
	}

	// Measure time
	start := time.Now()

	// Now run the implementation wrapped with a timeout
	err := WrapWithTimeout(impl, 1*time.Second)

	// Calculate Elapsed time
	elapsed := time.Since(start)

	assert.Error(t, err, "call to Lotus RPC timed out!")

	assert.Assert(t, elapsed.Milliseconds() < 1100)
	assert.Assert(t, elapsed.Milliseconds() > 900)
}

func TestDoNotTimeout(t *testing.T) {
	// wrap the implementation and use variable capture
	impl := func() {
		time.Sleep(2 * time.Second)
	}

	// Measure time
	start := time.Now()

	// Now run the implementation wrapped with a timeout
	err := WrapWithTimeout(impl, 3*time.Second)

	// Calculate Elapsed time
	elapsed := time.Since(start)

	assert.NilError(t, err)

	assert.Assert(t, elapsed.Milliseconds() < 2100)
	assert.Assert(t, elapsed.Milliseconds() > 1900)
}

func TestCopyVariable(t *testing.T) {
	// wrap the implementation and use variable capture
	var someVar string
	impl := func() {
		time.Sleep(1 * time.Second)
		someVar = "Copied"
	}

	err := WrapWithTimeout(impl, 3*time.Second)
	assert.NilError(t, err)
	assert.Equal(t, someVar, "Copied")
}

// Regression: lotusFunc finishing after the timeout fired used to race on the
// shared channel variable and could panic with "send on closed channel".
// Run with -race: fails deterministically on the old implementation. synctest
// also fails if the late worker is left blocked.
func TestLateCompletionAfterTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		err := WrapWithTimeout(func() {
			time.Sleep(2 * time.Millisecond)
		}, time.Millisecond)
		assert.Error(t, err, "call to Lotus RPC timed out!")

		// Advance fake time past the late completion before leaving the bubble.
		time.Sleep(2 * time.Millisecond)
	})
}
