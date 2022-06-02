package services

import (
	"context"
	"github.com/filecoin-project/go-state-types/abi"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/filecoin-project/specs-actors/v8/actors/builtin"
	"github.com/ipfs/go-cid"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"

	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	mocks "github.com/zondax/rosetta-filecoin-proxy/rosetta/services/mocks"
)

func TestAccountAPIService_AccountBalance(t *testing.T) {

	nodeMock := mocks.FullNodeMock{}

	type fields struct {
		network *types.NetworkIdentifier
		node    api.FullNode
	}
	type args struct {
		ctx     context.Context
		request *types.AccountBalanceRequest
	}

	// Mock needed input arguments
	var mockHeight int64 = 100
	var mockVestingStartEpoch = abi.ChainEpoch(139337)
	var mockVestingUnlockDur = abi.ChainEpoch(373248)
	var mockVestingInitialBalance = abi.NewTokenAmount(1000000)
	var mockAvailableBalance = abi.NewTokenAmount(100)
	mockTipSet := buildMockTargetTipSet(mockHeight)
	mockHeadTipSet := buildMockTargetTipSet(mockHeight + 10)
	mockTipSetHash, _ := BuildTipSetKeyHash(mockTipSet.Key())
	mockAddress := "t0128015"
	mockMsigActor := buildActorMock(builtin.MultisigActorCodeID, "100")
	///

	// Output
	mdVestingSchedule := make(map[string]interface{})
	vestingMap := map[string]string{}
	vestingMap[VestingStartEpochKey] = mockVestingStartEpoch.String()
	vestingMap[VestingUnlockDurationKey] = mockVestingUnlockDur.String()
	vestingMap[VestingInitialBalanceKey] = mockVestingInitialBalance.String()
	mdVestingSchedule[VestingScheduleStr] = vestingMap
	mdVestingSchedule[NonceKey] = "0"

	mdLockedBalanceOfMultiSig := make(map[string]interface{})
	mdLockedBalanceOfMultiSig[NonceKey] = "0"

	mdAvailableBalanceOfMultiSig := make(map[string]interface{})
	mdAvailableBalanceOfMultiSig[NonceKey] = "0"
	///

	// Mock functions
	nodeMock.On("StateNetworkName", mock.Anything).
		Return(dtypes.NetworkName(NetworkID.Network), nil)
	nodeMock.On("SyncState", mock.Anything).
		Return(&api.SyncState{
			ActiveSyncs: []api.ActiveSync{
				{
					Stage:  api.StageSyncComplete,
					Target: &filTypes.TipSet{},
				},
			},
		},
			nil)
	nodeMock.On("ChainGetTipSetByHeight", mock.Anything, mock.Anything, mock.Anything).
		Return(mockTipSet, nil)
	nodeMock.On("StateGetActor", mock.Anything, mock.Anything, mock.Anything).
		Return(mockMsigActor, nil)
	nodeMock.On("MsigGetAvailableBalance", mock.Anything, mock.Anything, mock.Anything).
		Return(mockAvailableBalance, nil)
	nodeMock.On("MsigGetVestingSchedule", mock.Anything, mock.Anything, mock.Anything).
		Return(api.MsigVesting{
			InitialBalance: mockVestingInitialBalance,
			StartEpoch:     mockVestingStartEpoch,
			UnlockDuration: mockVestingUnlockDur,
		},
			nil)
	nodeMock.On("ChainHead", mock.Anything).
		Return(mockHeadTipSet, nil)
	///

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *types.AccountBalanceResponse
		want1  *types.Error
	}{
		{
			name: "AvailableBalanceOfMultiSig",
			fields: fields{
				network: NetworkID,
				node:    &nodeMock,
			},
			args: args{
				ctx: context.Background(),
				request: &types.AccountBalanceRequest{
					NetworkIdentifier: NetworkID,
					BlockIdentifier: &types.PartialBlockIdentifier{
						Index: &mockHeight,
					},
					AccountIdentifier: &types.AccountIdentifier{
						Address:    mockAddress,
						SubAccount: nil,
						Metadata:   nil,
					},
				},
			},
			want: &types.AccountBalanceResponse{
				BlockIdentifier: &types.BlockIdentifier{
					Index: mockHeight,
					Hash:  *mockTipSetHash,
				},
				Balances: []*types.Amount{{
					Value:    mockMsigActor.Balance.String(),
					Currency: GetCurrencyData(),
					Metadata: nil,
				},
				},
				Metadata: mdAvailableBalanceOfMultiSig,
			},
			want1: nil,
		},
		{
			name: "LockedBalanceOfMultiSig",
			fields: fields{
				network: NetworkID,
				node:    &nodeMock,
			},
			args: args{
				ctx: context.Background(),
				request: &types.AccountBalanceRequest{
					NetworkIdentifier: NetworkID,
					BlockIdentifier: &types.PartialBlockIdentifier{
						Index: &mockHeight,
					},
					AccountIdentifier: &types.AccountIdentifier{
						Address: mockAddress,
						SubAccount: &types.SubAccountIdentifier{
							Address:  "LockedBalance",
							Metadata: nil,
						},
						Metadata: nil,
					},
				},
			},
			want: &types.AccountBalanceResponse{
				BlockIdentifier: &types.BlockIdentifier{
					Index: mockHeight,
					Hash:  *mockTipSetHash,
				},
				Balances: []*types.Amount{
					{
						Value:    "0",
						Currency: GetCurrencyData(),
						Metadata: nil,
					},
				},
				Metadata: mdLockedBalanceOfMultiSig,
			},
			want1: nil,
		},
		{
			name: "VestingSchedule",
			fields: fields{
				network: NetworkID,
				node:    &nodeMock,
			},
			args: args{
				ctx: context.Background(),
				request: &types.AccountBalanceRequest{
					NetworkIdentifier: NetworkID,
					BlockIdentifier: &types.PartialBlockIdentifier{
						Index: &mockHeight,
					},
					AccountIdentifier: &types.AccountIdentifier{
						Address: mockAddress,
						SubAccount: &types.SubAccountIdentifier{
							Address:  "VestingSchedule",
							Metadata: nil,
						},
						Metadata: nil,
					},
				},
			},
			want: &types.AccountBalanceResponse{
				BlockIdentifier: &types.BlockIdentifier{
					Index: mockHeight,
					Hash:  *mockTipSetHash,
				},
				Balances: []*types.Amount{{
					Value:    "0",
					Currency: GetCurrencyData(),
					Metadata: nil,
				},
				},
				Metadata: mdVestingSchedule,
			},
			want1: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AccountAPIService{
				network: tt.fields.network,
				node:    tt.fields.node,
			}
			got, got1 := a.AccountBalance(tt.args.ctx, tt.args.request)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccountBalance() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("AccountBalance() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestNewAccountAPIService(t *testing.T) {
	type args struct {
		network *types.NetworkIdentifier
		node    *api.FullNode
	}
	tests := []struct {
		name string
		args args
		want server.AccountAPIServicer
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAccountAPIService(tt.args.network, tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAccountAPIService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func buildActorMock(actorCode cid.Cid, balanceStr string) *filTypes.Actor {
	balance, _ := filTypes.BigFromString(balanceStr)
	return &filTypes.Actor{
		Code:    actorCode,
		Head:    cid.Cid{},
		Nonce:   0,
		Balance: balance,
	}
}
