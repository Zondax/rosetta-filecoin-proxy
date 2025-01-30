package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/zondax/rosetta-filecoin-proxy/rosetta/services"
)

func TestFindMethodNameInAllActors(t *testing.T) {
	tests := []struct {
		name   string
		method uint64
		want   string
	}{
		{
			name:   "Test with method 0",
			method: 0,
			want:   "<unknown>",
		},
		{
			name:   "Test with EVM method",
			method: 3844450837,
			want:   "InvokeContract",
		},
		{
			name:   "Test with invalid method",
			method: 999999,
			want:   "<unknown>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := services.FindMethodNameInAllActors(tt.method)
			require.Equal(t, tt.want, got, "Method name mismatch for method %d", tt.method)
		})
	}
}
