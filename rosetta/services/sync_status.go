package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/filecoin-project/lotus/api"
	"math"
)

const FactorSecondToMillisecond int64 = 1e3

type SyncStatus struct {
	targetIndex     []int64
	currentHeight   []int64
	globalSyncState api.SyncStateStage
}

func (status SyncStatus) IsSynced() bool {
	return status.globalSyncState == api.StageSyncComplete
}

func (status SyncStatus) GetMaxHeight() int64 {

	if status.globalSyncState == api.StageSyncErrored {
		return -1
	}

	if status.globalSyncState < api.StageMessages {
		return 0
	}

	var max int64
	for _, height := range status.currentHeight {
		if height > max {
			max = height
		}
	}

	return max
}

func (status SyncStatus) GetMinHeight() int64 {
	if status.globalSyncState == api.StageSyncErrored {
		return -1
	}

	if status.globalSyncState < api.StageMessages {
		return 0
	}

	if len(status.currentHeight) == 0 {
		return 0
	}

	var min int64 = math.MaxInt64
	for _, height := range status.currentHeight {
		if height < min {
			min = height
		}
	}

	return min
}

func (status SyncStatus) GetTargetIndex() *int64 {
	var target int64
	for _, height := range status.targetIndex {
		if height > target {
			target = height
		}
	}

	return &target
}

func CheckSyncStatus(ctx context.Context, node *api.FullNode) (*SyncStatus, *types.Error) {

	fullAPI := *node
	syncState, err := fullAPI.SyncState(ctx)

	if err != nil || len(syncState.ActiveSyncs) == 0 {
		return nil, BuildError(ErrUnableToGetSyncStatus, err, true)
	}

	var (
		status = SyncStatus{
			globalSyncState: api.StageIdle,
		}
		syncComplete = true
		earliestStat = api.StageIdle
	)

	for _, w := range syncState.ActiveSyncs {
		if w.Target == nil {
			continue
		}

		switch w.Stage {
		case api.StageSyncErrored:
			return nil, ErrSyncErrored
		case api.StageSyncComplete:
			break
		default:
			syncComplete = false
			if w.Stage > earliestStat {
				earliestStat = w.Stage
			}
		}

		status.currentHeight = append(status.currentHeight, int64(w.Height))
		status.targetIndex = append(status.targetIndex, int64(w.Target.Height()))
	}

	if syncComplete {
		status.globalSyncState = api.StageSyncComplete
	} else {
		status.globalSyncState = earliestStat
	}

	return &status, nil
}
