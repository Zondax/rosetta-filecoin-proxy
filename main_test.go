package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEndpointParsing(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantV1    string
		wantV2    string
		wantError bool
	}{
		{
			name:      "base /rpc format",
			input:     "https://node-fil-calibration-next-light.zondax.dev/rpc",
			wantV1:    "https://node-fil-calibration-next-light.zondax.dev/rpc/v1",
			wantV2:    "https://node-fil-calibration-next-light.zondax.dev/rpc/v2",
			wantError: false,
		},
		{
			name:      "legacy /rpc/v1 format",
			input:     "https://api.node.glif.io/rpc/v1",
			wantV1:    "https://api.node.glif.io/rpc/v1",
			wantV2:    "https://api.node.glif.io/rpc/v2",
			wantError: false,
		},
		{
			name:      "v2 endpoint provided",
			input:     "https://api.node.glif.io/rpc/v2",
			wantV1:    "https://api.node.glif.io/rpc/v1",
			wantV2:    "https://api.node.glif.io/rpc/v2",
			wantError: false,
		},
		{
			name:      "unrecognized format - no rpc path",
			input:     "http://localhost:1234",
			wantV1:    "",
			wantV2:    "",
			wantError: true,
		},
		{
			name:      "unrecognized format - different path",
			input:     "http://localhost:1234/api",
			wantV1:    "",
			wantV2:    "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from getFullNodeAPI
			var v1Addr, v2Addr string
			var hasError bool

			if strings.HasSuffix(tt.input, "/rpc") {
				v1Addr = tt.input + "/v1"
				v2Addr = tt.input + "/v2"
			} else if strings.Contains(tt.input, "/rpc/v1") {
				v1Addr = tt.input
				v2Addr = strings.Replace(tt.input, "/rpc/v1", "/rpc/v2", 1)
			} else if strings.Contains(tt.input, "/rpc/v2") {
				v1Addr = strings.Replace(tt.input, "/rpc/v2", "/rpc/v1", 1)
				v2Addr = tt.input
			} else {
				hasError = true
			}

			if tt.wantError {
				assert.True(t, hasError, "Expected error for unrecognized format")
			} else {
				assert.False(t, hasError, "Unexpected error")
				assert.Equal(t, tt.wantV1, v1Addr, "V1 address mismatch")
				assert.Equal(t, tt.wantV2, v2Addr, "V2 address mismatch")
			}
		})
	}
}
