module github.com/zondax/rosetta-filecoin-proxy

go 1.16

require (
	github.com/coinbase/rosetta-sdk-go v0.6.10
	github.com/filecoin-project/go-address v0.0.5
	github.com/filecoin-project/go-amt-ipld/v2 v2.1.1-0.20201006184820-924ee87a1349 // indirect
	github.com/filecoin-project/go-bitfield v0.2.4
	github.com/filecoin-project/go-data-transfer v1.10.1
	github.com/filecoin-project/go-fil-markets v1.12.0
	github.com/filecoin-project/go-jsonrpc v0.1.4-0.20210217175800-45ea43ac2bec
	github.com/filecoin-project/go-multistore v0.0.3
	github.com/filecoin-project/go-state-types v0.1.1-0.20210915140513-d354ccf10379
	github.com/filecoin-project/lotus v1.12.0-rc1
	github.com/filecoin-project/specs-actors v0.9.14
	github.com/filecoin-project/specs-actors/v2 v2.3.5
	github.com/filecoin-project/specs-actors/v4 v4.0.1
	github.com/google/uuid v1.2.0
	github.com/ipfs/go-cid v0.1.0
	github.com/ipfs/go-log v1.0.5
	github.com/libp2p/go-libp2p-core v0.8.6
	github.com/multiformats/go-multihash v0.0.15
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/zondax/rosetta-filecoin-lib v1.1200.0
	gotest.tools v2.2.0+incompatible
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
