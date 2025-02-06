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

func TestFindMethodInType_NonStruct(t *testing.T) {
	type TestStruct struct {
		Method1 uint64
		Method2 uint64
		Method3 uint64
	}

	tests := []struct {
		name      string
		methodNum uint64
		actorType interface{}
		want      string
	}{
		// Happy path cases
		{
			name:      "valid struct with existing method",
			methodNum: 1,
			actorType: TestStruct{Method1: 1, Method2: 2, Method3: 3},
			want:      "Method1",
		},
		{
			name:      "valid struct with non-existing method",
			methodNum: 99,
			actorType: TestStruct{Method1: 1, Method2: 2, Method3: 3},
			want:      "<unknown>",
		},
		// Error cases
		{
			name:      "string input should return UnknownStr",
			methodNum: 1,
			actorType: "not a struct",
			want:      "<unknown>",
		},
		{
			name:      "nil input should return UnknownStr",
			methodNum: 1,
			actorType: nil,
			want:      "<unknown>",
		},
		{
			name:      "integer input should return UnknownStr",
			methodNum: 1,
			actorType: 42,
			want:      "<unknown>",
		},
		{
			name:      "slice input should return UnknownStr",
			methodNum: 1,
			actorType: []string{"not", "a", "struct"},
			want:      "<unknown>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := services.FindMethodInType(tt.methodNum, tt.actorType)
			require.Equal(t, tt.want, got, "FindMethodInType result mismatch for test case: %s", tt.name)
		})
	}
}
