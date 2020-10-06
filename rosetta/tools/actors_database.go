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
	//Address-ActorCID Map
	GetActorCode(address address.Address) (cid.Cid, error)
	storeActorCode(address address.Address, actorCode cid.Cid)
	//Address-ActorPubkey Map
	GetActorPubKey(address address.Address) (string, error)
	storeActorPubKey(address address.Address, pubKey string)
}

/// In-memory database ///
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

func (m *Cache) GetActorCode(address address.Address) (cid.Cid, error) {
	code, ok := m.cidMap.Get(address.String())
	if !ok {
		var err error
		code, err = m.retrieveActorFromLotus(address)
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

func (m *Cache) retrieveActorFromLotus(add address.Address) (cid.Cid, error) {
	actor, err := (*m.Node).StateGetActor(context.Background(), add, filTypes.EmptyTSK)
	if err != nil {
		return cid.Cid{}, err
	}

	return actor.Code, nil
}

func (m *Cache) GetActorPubKey(address address.Address) (string, error) {
	pubKey, ok := m.pubKeyMap.Get(address.String())
	if !ok {
		var err error
		pubKey, err = m.retrieveActorPubKeyFromLotus(address)
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

func (m *Cache) retrieveActorPubKeyFromLotus(add address.Address) (string, error) {
	key, err := (*m.Node).StateAccountKey(context.Background(), add, filTypes.EmptyTSK)
	if err != nil {
		return add.String(), nil
	}

	return key.String(), nil
}

/////
