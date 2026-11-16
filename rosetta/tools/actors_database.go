package tools

import (
	"context"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	filTypes "github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
	"github.com/orcaman/concurrent-map"
)

var ActorsDB Database

type Database interface {
	NewImpl(*api.FullNode)
	// Address-ActorCID Map
	GetActorCode(ctx context.Context, address address.Address) (cid.Cid, error)
	storeActorCode(address address.Address, actorCode cid.Cid)
	// Address-ActorPubkey Map
	GetActorPubKey(ctx context.Context, address address.Address, reverse bool) (string, error)
	storeActorPubKey(address address.Address, pubKey string)
}

// In-memory database
type Cache struct {
	cidMap    cmap.ConcurrentMap
	pubKeyMap cmap.ConcurrentMap
	Node      *api.FullNode
}

func (m *Cache) NewImpl(node *api.FullNode) {
	m.cidMap = cmap.New()
	m.pubKeyMap = cmap.New()
	m.Node = node
}

func (m *Cache) GetActorCode(ctx context.Context, address address.Address) (cid.Cid, error) {
	code, ok := m.cidMap.Get(address.String())
	if !ok {
		var err error
		code, err = m.retrieveActorFromLotus(ctx, address)
		if err != nil {
			return cid.Cid{}, err
		}
		m.storeActorCode(address, code.(cid.Cid))
	}

	return code.(cid.Cid), nil
}

func (m *Cache) storeActorCode(key address.Address, value cid.Cid) {
	m.cidMap.Set(key.String(), value)
}

func (m *Cache) retrieveActorFromLotus(ctx context.Context, add address.Address) (cid.Cid, error) {
	// Keep the request values (forwarded headers) but not its cancellation, so a
	// shared cache fill isn't aborted by one client disconnecting (same lifetime
	// as the previous context.Background()).
	ctx = context.WithoutCancel(ctx)
	actor, err := (*m.Node).StateGetActor(ctx, add, filTypes.EmptyTSK)
	if err != nil {
		return cid.Cid{}, err
	}

	return actor.Code, nil
}

func (m *Cache) GetActorPubKey(ctx context.Context, address address.Address, reverse bool) (string, error) {
	pubKey, ok := m.pubKeyMap.Get(address.String())
	if !ok {
		var err error
		pubKey, err = m.retrieveActorPubKeyFromLotus(ctx, address, reverse)
		if err != nil {
			return address.String(), err
		}
		m.storeActorPubKey(address, pubKey.(string))
	}

	return pubKey.(string), nil
}

func (m *Cache) storeActorPubKey(address address.Address, pubKey string) {
	m.pubKeyMap.Set(address.String(), pubKey)
}

func (m *Cache) retrieveActorPubKeyFromLotus(ctx context.Context, add address.Address, reverse bool) (string, error) {
	// Lookup errors are cached below as the fallback value, so a client disconnect
	// must not be able to cause one (same lifetime as the previous context.Background()).
	ctx = context.WithoutCancel(ctx)
	var key address.Address
	var err error
	if reverse {
		key, err = (*m.Node).StateLookupID(ctx, add, filTypes.EmptyTSK)
	} else {
		key, err = (*m.Node).StateAccountKey(ctx, add, filTypes.EmptyTSK)
	}

	if err != nil {
		return add.String(), nil
	}
	return key.String(), nil
}
