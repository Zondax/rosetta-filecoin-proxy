package tools

import (
	"gotest.tools/assert"
	"log"
	"testing"
	"time"
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

	assert.Error(t, err, "Lotus RPC call Timed out!")

	log.Printf("Elapsed %d", elapsed.Milliseconds())
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

	log.Printf("Elapsed %d", elapsed.Milliseconds())
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
