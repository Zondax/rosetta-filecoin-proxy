module github.com/zondax/rosetta-filecoin-proxy

go 1.15

require (
	github.com/coinbase/rosetta-sdk-go v0.5.10
	github.com/filecoin-project/go-address v0.0.5-0.20201103152444-f2023ef3f5bb
	github.com/filecoin-project/go-bitfield v0.2.3-0.20201110211213-fe2c1862e816
	github.com/filecoin-project/go-data-transfer v1.2.3
	github.com/filecoin-project/go-fil-markets v1.0.10
	github.com/filecoin-project/go-jsonrpc v0.1.2
	github.com/filecoin-project/go-multistore v0.0.3
	github.com/filecoin-project/go-state-types v0.0.0-20201102161440-c8033295a1fc
	github.com/filecoin-project/lotus v1.4.0
	github.com/filecoin-project/specs-actors v0.9.13
	github.com/filecoin-project/specs-actors/v2 v2.3.3
	github.com/google/uuid v1.1.2
	github.com/ipfs/go-cid v0.0.7
	github.com/ipfs/go-log v1.0.4
	github.com/libp2p/go-libp2p-core v0.7.0
	github.com/multiformats/go-multihash v0.0.14
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/stretchr/testify v1.6.1
	github.com/zondax/rosetta-filecoin-lib v1.201.0
	gotest.tools v2.2.0+incompatible
)

replace github.com/filecoin-project/filecoin-ffi => ./extern/filecoin-ffi
