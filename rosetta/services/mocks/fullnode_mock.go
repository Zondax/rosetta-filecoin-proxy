package services

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-fil-markets/storagemarket"
	"github.com/filecoin-project/go-jsonrpc/auth"
	"github.com/filecoin-project/go-multistore"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	marketevents "github.com/filecoin-project/lotus/markets/loggers"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/filecoin-project/specs-actors/actors/builtin/miner"
	"github.com/filecoin-project/specs-actors/actors/builtin/paych"
	"github.com/filecoin-project/specs-actors/actors/builtin/verifreg"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/mock"
)

type FullNodeMock struct {
	mock.Mock
}

func (f *FullNodeMock) ClientGetDealUpdates(ctx context.Context) (<-chan api.DealInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychVoucherSubmit(ctx context.Context, a address.Address, voucher *paych.SignedVoucher, bytes []byte, bytes2 []byte) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolPushMessage(ctx context.Context, msg *filTypes.Message, spec *api.MessageSendSpec) (*filTypes.SignedMessage, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientImport(ctx context.Context, ref api.FileRef) (*api.ImportRes, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientFindData(ctx context.Context, root cid.Cid, piece *cid.Cid) ([]api.QueryOffer, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientMinerQueryOffer(ctx context.Context, miner address.Address, root cid.Cid, piece *cid.Cid) (api.QueryOffer, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerDeadlines(ctx context.Context, a address.Address, key filTypes.TipSetKey) ([]*miner.Deadline, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerFaults(ctx context.Context, a address.Address, key filTypes.TipSetKey) (abi.BitField, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerRecoveries(ctx context.Context, a address.Address, key filTypes.TipSetKey) (abi.BitField, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerInitialPledgeCollateral(ctx context.Context, a address.Address, info miner.SectorPreCommitInfo, key filTypes.TipSetKey) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) MsigCreate(ctx context.Context, u uint64, addresses []address.Address, epoch abi.ChainEpoch, bigInt filTypes.BigInt, a address.Address, bigInt2 filTypes.BigInt) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) MsigCancel(ctx context.Context, a address.Address, u uint64, a2 address.Address, bigInt filTypes.BigInt, a3 address.Address, u2 uint64, bytes []byte) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) NetAutoNatStatus(ctx context.Context) (api.NatInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetRandomnessFromTickets(ctx context.Context, tsk filTypes.TipSetKey, personalization crypto.DomainSeparationTag, randEpoch abi.ChainEpoch, entropy []byte) (abi.Randomness, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetRandomnessFromBeacon(ctx context.Context, tsk filTypes.TipSetKey, personalization crypto.DomainSeparationTag, randEpoch abi.ChainEpoch, entropy []byte) (abi.Randomness, error) {
	panic("implement me")
}

func (f *FullNodeMock) BeaconGetEntry(ctx context.Context, epoch abi.ChainEpoch) (*filTypes.BeaconEntry, error) {
	panic("implement me")
}

func (f *FullNodeMock) GasEstimateFeeCap(ctx context.Context, message *filTypes.Message, i int64, key filTypes.TipSetKey) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) GasEstimateGasLimit(ctx context.Context, message *filTypes.Message, key filTypes.TipSetKey) (int64, error) {
	panic("implement me")
}

func (f *FullNodeMock) GasEstimateGasPremium(_ context.Context, nblocksincl uint64, sender address.Address, gaslimit int64, tsk filTypes.TipSetKey) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) GasEstimateMessageGas(ctx context.Context, message *filTypes.Message, spec *api.MessageSendSpec, key filTypes.TipSetKey) (*filTypes.Message, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolSelect(ctx context.Context, key filTypes.TipSetKey, f2 float64) ([]*filTypes.SignedMessage, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolClear(ctx context.Context, b bool) error {
	panic("implement me")
}

func (f *FullNodeMock) MpoolGetConfig(ctx context.Context) (*filTypes.MpoolConfig, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolSetConfig(ctx context.Context, config *filTypes.MpoolConfig) error {
	panic("implement me")
}

func (f *FullNodeMock) ClientRemoveImport(ctx context.Context, importID multistore.StoreID) error {
	panic("implement me")
}

func (f *FullNodeMock) ClientRetrieveWithEvents(ctx context.Context, order api.RetrievalOrder, ref *api.FileRef) (<-chan marketevents.RetrievalEvent, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientDealSize(ctx context.Context, root cid.Cid) (api.DataSize, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientListDataTransfers(ctx context.Context) ([]api.DataTransferChannel, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientDataTransferUpdates(ctx context.Context) (<-chan api.DataTransferChannel, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerActiveSectors(ctx context.Context, a address.Address, key filTypes.TipSetKey) ([]*api.ChainSectorInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerPartitions(ctx context.Context, a address.Address, u uint64, key filTypes.TipSetKey) ([]*miner.Partition, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerPreCommitDepositForPower(ctx context.Context, a address.Address, info miner.SectorPreCommitInfo, key filTypes.TipSetKey) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateSectorExpiration(ctx context.Context, a address.Address, number abi.SectorNumber, key filTypes.TipSetKey) (*api.SectorExpiration, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateSectorPartition(ctx context.Context, maddr address.Address, sectorNumber abi.SectorNumber, tok filTypes.TipSetKey) (*api.SectorLocation, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateVerifiedClientStatus(ctx context.Context, addr address.Address, tsk filTypes.TipSetKey) (*verifreg.DataCap, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateDealProviderCollateralBounds(ctx context.Context, size abi.PaddedPieceSize, b bool, key filTypes.TipSetKey) (api.DealCollateralBounds, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateCirculatingSupply(ctx context.Context, key filTypes.TipSetKey) (api.CirculatingSupply, error) {
	panic("implement me")
}

func (f *FullNodeMock) MsigSwapPropose(ctx context.Context, a address.Address, a2 address.Address, a3 address.Address, a4 address.Address) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) MsigSwapApprove(ctx context.Context, a address.Address, a2 address.Address, u uint64, a3 address.Address, a4 address.Address, a5 address.Address) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) MsigSwapCancel(ctx context.Context, a address.Address, a2 address.Address, u uint64, a3 address.Address, a4 address.Address) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychGetWaitReady(ctx context.Context, c cid.Cid) (address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychSettle(ctx context.Context, a address.Address) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychCollect(ctx context.Context, a address.Address) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) AuthVerify(ctx context.Context, token string) ([]auth.Permission, error) {
	panic("implement me")
}

func (f *FullNodeMock) AuthNew(ctx context.Context, perms []auth.Permission) ([]byte, error) {
	panic("implement me")
}

func (f *FullNodeMock) NetConnectedness(ctx context.Context, id peer.ID) (network.Connectedness, error) {
	panic("implement me")
}

func (f *FullNodeMock) NetPeers(ctx context.Context) ([]peer.AddrInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) NetConnect(ctx context.Context, info peer.AddrInfo) error {
	panic("implement me")
}

func (f *FullNodeMock) NetAddrsListen(ctx context.Context) (peer.AddrInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) NetDisconnect(ctx context.Context, id peer.ID) error {
	panic("implement me")
}

func (f *FullNodeMock) NetFindPeer(ctx context.Context, id peer.ID) (peer.AddrInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) NetPubsubScores(ctx context.Context) ([]api.PubsubScore, error) {
	panic("implement me")
}

func (f *FullNodeMock) ID(ctx context.Context) (peer.ID, error) {
	panic("implement me")
}

func (f *FullNodeMock) Version(ctx context.Context) (api.Version, error) {
	panic("implement me")
}

func (f *FullNodeMock) LogList(ctx context.Context) ([]string, error) {
	panic("implement me")
}

func (f *FullNodeMock) LogSetLevel(ctx context.Context, s string, s2 string) error {
	panic("implement me")
}

func (f *FullNodeMock) Shutdown(ctx context.Context) error {
	panic("implement me")
}

func (f *FullNodeMock) Closing(ctx context.Context) (<-chan struct{}, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainNotify(ctx context.Context) (<-chan []*api.HeadChange, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainHead(ctx context.Context) (*filTypes.TipSet, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetRandomness(ctx context.Context, tsk filTypes.TipSetKey, personalization crypto.DomainSeparationTag, randEpoch abi.ChainEpoch, entropy []byte) (abi.Randomness, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetBlock(ctx context.Context, c cid.Cid) (*filTypes.BlockHeader, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetTipSet(ctx context.Context, key filTypes.TipSetKey) (*filTypes.TipSet, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetBlockMessages(ctx context.Context, blockCid cid.Cid) (*api.BlockMessages, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetParentReceipts(ctx context.Context, blockCid cid.Cid) ([]*filTypes.MessageReceipt, error) {
	args := f.Called(ctx, blockCid)
	return args.Get(0).([]*filTypes.MessageReceipt), args.Error(1)
}

func (f *FullNodeMock) ChainGetParentMessages(ctx context.Context, blockCid cid.Cid) ([]api.Message, error) {
	args := f.Called(ctx, blockCid)
	return args.Get(0).([]api.Message), args.Error(1)
}

func (f *FullNodeMock) ChainGetTipSetByHeight(ctx context.Context, epoch abi.ChainEpoch, key filTypes.TipSetKey) (*filTypes.TipSet, error) {
	args := f.Called(ctx, epoch, key)
	return args.Get(0).(*filTypes.TipSet), args.Error(1)
}

func (f *FullNodeMock) ChainReadObj(ctx context.Context, c cid.Cid) ([]byte, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainHasObj(ctx context.Context, c cid.Cid) (bool, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainStatObj(ctx context.Context, c cid.Cid, c2 cid.Cid) (api.ObjStat, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainSetHead(ctx context.Context, key filTypes.TipSetKey) error {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetGenesis(ctx context.Context) (*filTypes.TipSet, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainTipSetWeight(ctx context.Context, key filTypes.TipSetKey) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetNode(ctx context.Context, p string) (*api.IpldObject, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetMessage(ctx context.Context, c cid.Cid) (*filTypes.Message, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainGetPath(ctx context.Context, from filTypes.TipSetKey, to filTypes.TipSetKey) ([]*api.HeadChange, error) {
	panic("implement me")
}

func (f *FullNodeMock) ChainExport(ctx context.Context, key filTypes.TipSetKey) (<-chan []byte, error) {
	panic("implement me")
}

func (f *FullNodeMock) SyncState(ctx context.Context) (*api.SyncState, error) {
	args := f.Called(ctx)
	return args.Get(0).(*api.SyncState), args.Error(1)
}

func (f *FullNodeMock) SyncSubmitBlock(ctx context.Context, blk *filTypes.BlockMsg) error {
	panic("implement me")
}

func (f *FullNodeMock) SyncIncomingBlocks(ctx context.Context) (<-chan *filTypes.BlockHeader, error) {
	panic("implement me")
}

func (f *FullNodeMock) SyncMarkBad(ctx context.Context, bcid cid.Cid) error {
	panic("implement me")
}

func (f *FullNodeMock) SyncCheckBad(ctx context.Context, bcid cid.Cid) (string, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolPending(ctx context.Context, key filTypes.TipSetKey) ([]*filTypes.SignedMessage, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolPush(ctx context.Context, message *filTypes.SignedMessage) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolGetNonce(ctx context.Context, a address.Address) (uint64, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolSub(ctx context.Context) (<-chan api.MpoolUpdate, error) {
	panic("implement me")
}

func (f *FullNodeMock) MpoolEstimateGasPrice(ctx context.Context, nblocksincl uint64, sender address.Address, gaslimit int64, tsk filTypes.TipSetKey) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) MinerGetBaseInfo(ctx context.Context, a address.Address, epoch abi.ChainEpoch, key filTypes.TipSetKey) (*api.MiningBaseInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) MinerCreateBlock(ctx context.Context, template *api.BlockTemplate) (*filTypes.BlockMsg, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletNew(ctx context.Context, sigType crypto.SigType) (address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletHas(ctx context.Context, a address.Address) (bool, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletList(ctx context.Context) ([]address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletBalance(ctx context.Context, a address.Address) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletSign(ctx context.Context, a address.Address, bytes []byte) (*crypto.Signature, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletSignMessage(ctx context.Context, a address.Address, message *filTypes.Message) (*filTypes.SignedMessage, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletVerify(ctx context.Context, a address.Address, bytes []byte, signature *crypto.Signature) bool {
	panic("implement me")
}

func (f *FullNodeMock) WalletDefaultAddress(ctx context.Context) (address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletSetDefault(ctx context.Context, a address.Address) error {
	panic("implement me")
}

func (f *FullNodeMock) WalletExport(ctx context.Context, a address.Address) (*filTypes.KeyInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletImport(ctx context.Context, info *filTypes.KeyInfo) (address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) WalletDelete(ctx context.Context, a address.Address) error {
	panic("implement me")
}

func (f *FullNodeMock) ClientStartDeal(ctx context.Context, params *api.StartDealParams) (*cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientGetDealInfo(ctx context.Context, c cid.Cid) (*api.DealInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientListDeals(ctx context.Context) ([]api.DealInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientHasLocal(ctx context.Context, root cid.Cid) (bool, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientRetrieve(ctx context.Context, order api.RetrievalOrder, ref *api.FileRef) error {
	panic("implement me")
}

func (f *FullNodeMock) ClientQueryAsk(ctx context.Context, p peer.ID, miner address.Address) (*storagemarket.SignedStorageAsk, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientCalcCommP(ctx context.Context, inpath string) (*api.CommPRet, error) {
	panic("implement me")
}

func (f *FullNodeMock) ClientGenCar(ctx context.Context, ref api.FileRef, outpath string) error {
	panic("implement me")
}

func (f *FullNodeMock) ClientListImports(ctx context.Context) ([]api.Import, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateCall(ctx context.Context, message *filTypes.Message, key filTypes.TipSetKey) (*api.InvocResult, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateReplay(ctx context.Context, key filTypes.TipSetKey, c cid.Cid) (*api.InvocResult, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateGetActor(ctx context.Context, actor address.Address, tsk filTypes.TipSetKey) (*filTypes.Actor, error) {
	args := f.Called(ctx, actor, tsk)
	return args.Get(0).(*filTypes.Actor), args.Error(1)
}

func (f *FullNodeMock) StateReadState(ctx context.Context, actor address.Address, tsk filTypes.TipSetKey) (*api.ActorState, error) {
	args := f.Called(ctx, actor, tsk)
	return args.Get(0).(*api.ActorState), args.Error(1)
}

func (f *FullNodeMock) StateListMessages(ctx context.Context, match *filTypes.Message, tsk filTypes.TipSetKey, toht abi.ChainEpoch) ([]cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateNetworkName(ctx context.Context) (dtypes.NetworkName, error) {
	args := f.Called(ctx)
	return args.Get(0).(dtypes.NetworkName), args.Error(1)
}

func (f *FullNodeMock) StateMinerSectors(ctx context.Context, a address.Address, field *abi.BitField, b bool, key filTypes.TipSetKey) ([]*api.ChainSectorInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerProvingSet(ctx context.Context, a address.Address, key filTypes.TipSetKey) ([]*api.ChainSectorInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerProvingDeadline(ctx context.Context, a address.Address, key filTypes.TipSetKey) (*miner.DeadlineInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerPower(ctx context.Context, a address.Address, key filTypes.TipSetKey) (*api.MinerPower, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerInfo(ctx context.Context, a address.Address, key filTypes.TipSetKey) (api.MinerInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateAllMinerFaults(ctx context.Context, lookback abi.ChainEpoch, ts filTypes.TipSetKey) ([]*api.Fault, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerAvailableBalance(ctx context.Context, a address.Address, key filTypes.TipSetKey) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateSectorPreCommitInfo(ctx context.Context, a address.Address, number abi.SectorNumber, key filTypes.TipSetKey) (miner.SectorPreCommitOnChainInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateSectorGetInfo(ctx context.Context, a address.Address, number abi.SectorNumber, key filTypes.TipSetKey) (*miner.SectorOnChainInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) StatePledgeCollateral(ctx context.Context, key filTypes.TipSetKey) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateWaitMsg(ctx context.Context, cid cid.Cid, confidence uint64) (*api.MsgLookup, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateSearchMsg(ctx context.Context, c cid.Cid) (*api.MsgLookup, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateListMiners(ctx context.Context, key filTypes.TipSetKey) ([]address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateListActors(ctx context.Context, key filTypes.TipSetKey) ([]address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMarketBalance(ctx context.Context, a address.Address, key filTypes.TipSetKey) (api.MarketBalance, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMarketParticipants(ctx context.Context, key filTypes.TipSetKey) (map[string]api.MarketBalance, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMarketDeals(ctx context.Context, key filTypes.TipSetKey) (map[string]api.MarketDeal, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMarketStorageDeal(ctx context.Context, id abi.DealID, key filTypes.TipSetKey) (*api.MarketDeal, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateLookupID(ctx context.Context, a address.Address, key filTypes.TipSetKey) (address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateAccountKey(ctx context.Context, a address.Address, key filTypes.TipSetKey) (address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateChangedActors(ctx context.Context, c cid.Cid, c2 cid.Cid) (map[string]filTypes.Actor, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateGetReceipt(ctx context.Context, c cid.Cid, key filTypes.TipSetKey) (*filTypes.MessageReceipt, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateMinerSectorCount(ctx context.Context, a address.Address, key filTypes.TipSetKey) (api.MinerSectors, error) {
	panic("implement me")
}

func (f *FullNodeMock) StateCompute(ctx context.Context, epoch abi.ChainEpoch, messages []*filTypes.Message, key filTypes.TipSetKey) (*api.ComputeStateOutput, error) {
	panic("implement me")
}

func (f *FullNodeMock) MsigGetAvailableBalance(ctx context.Context, a address.Address, key filTypes.TipSetKey) (filTypes.BigInt, error) {
	args := f.Called(ctx, a, key)
	return args.Get(0).(filTypes.BigInt), args.Error(1)
}

func (f *FullNodeMock) MsigPropose(ctx context.Context, a address.Address, a2 address.Address, bigInt filTypes.BigInt, a3 address.Address, u uint64, bytes []byte) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) MsigApprove(ctx context.Context, a address.Address, u uint64, a2 address.Address, a3 address.Address, bigInt filTypes.BigInt, a4 address.Address, u2 uint64, bytes []byte) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) MarketEnsureAvailable(ctx context.Context, a address.Address, a2 address.Address, bigInt filTypes.BigInt) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychGet(ctx context.Context, from, to address.Address, ensureFunds filTypes.BigInt) (*api.ChannelInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychList(ctx context.Context) ([]address.Address, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychStatus(ctx context.Context, a address.Address) (*api.PaychStatus, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychClose(ctx context.Context, a address.Address) (cid.Cid, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychAllocateLane(ctx context.Context, ch address.Address) (uint64, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychNewPayment(ctx context.Context, from, to address.Address, vouchers []api.VoucherSpec) (*api.PaymentInfo, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychVoucherCheckValid(ctx context.Context, a address.Address, voucher *paych.SignedVoucher) error {
	panic("implement me")
}

func (f *FullNodeMock) PaychVoucherCheckSpendable(ctx context.Context, a address.Address, voucher *paych.SignedVoucher, bytes []byte, bytes2 []byte) (bool, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychVoucherCreate(ctx context.Context, a address.Address, bigInt filTypes.BigInt, u uint64) (*paych.SignedVoucher, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychVoucherAdd(ctx context.Context, a address.Address, voucher *paych.SignedVoucher, bytes []byte, bigInt filTypes.BigInt) (filTypes.BigInt, error) {
	panic("implement me")
}

func (f *FullNodeMock) PaychVoucherList(ctx context.Context, a address.Address) ([]*paych.SignedVoucher, error) {
	panic("implement me")
}
