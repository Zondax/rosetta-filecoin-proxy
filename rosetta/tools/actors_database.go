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
	GetActorCode(address address.Address) (cid.Cid, error)
	storeActorCode(address address.Address, actorCode cid.Cid)
}

/// In-memory database ///
type Cache struct {
	data cmap.ConcurrentMap
	Node *api.FullNode
}

func (m *Cache) NewImpl(node *api.FullNode) {
	m.data = cmap.New()
	m.Node = node
}

func (m *Cache) GetActorCode(address address.Address) (cid.Cid, error) {
	code, ok := m.data.Get(address.String())
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
	m.data.Set(key.String(), value)
}

func (m *Cache) retrieveActorFromLotus(add address.Address) (cid.Cid, error) {
	actor, err := (*m.Node).StateGetActor(context.Background(), add, filTypes.EmptyTSK)
	if err != nil {
		return cid.Cid{}, err
	}

	return actor.Code, nil
}

/////
