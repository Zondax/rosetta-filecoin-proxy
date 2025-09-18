# Rosetta Filecoin Proxy

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CircleCI](https://circleci.com/gh/Zondax/rosetta-filecoin-proxy/tree/master.svg?style=shield)](https://circleci.com/gh/Zondax/rosetta-filecoin-proxy/tree/master)
[![Github-Actions](https://github.com/Zondax/rosetta-filecoin-proxy/workflows/rosetta-cli/badge.svg)](https://github.com/Zondax/rosetta-filecoin-proxy/actions)

A Rosetta API implementation for Filecoin that acts as a proxy between the standardized Rosetta API and Filecoin Lotus nodes.

## Quick Start

### Build & Run

```bash
# Build
make build

# Run with required environment variables
LOTUS_RPC_URL="YOUR_RPC_URL" \
LOTUS_RPC_TOKEN="YOUR_RPC_TOKEN" \
./rosetta-filecoin-proxy
```

Server starts on port `8080`.

### Configuration

| Variable                 | Description                    | Required              |
| ------------------------ | ------------------------------ | --------------------- |
| `LOTUS_RPC_URL`          | Lotus node RPC endpoint        | Yes                   |
| `LOTUS_RPC_TOKEN`        | Lotus node auth token          | Yes if node auth      |
| `ENABLE_LOTUS_V2_APIS`   | Enable F3 finality support     | No (default: `false`) |
| `FORCE_SAFE_F3_FINALITY` | Force safe finality as default | No (default: `false`) |

## F3 Finality Support

The proxy supports Filecoin's F3 finality with three tags: `latest`, `safe`, `finalized`.

Use finality by adding `sub_network_identifier` to requests:

```json
{
  "network_identifier": {
    "blockchain": "Filecoin",
    "network": "mainnet",
    "sub_network_identifier": {
      "network": "f3",
      "metadata": {
        "finality_tag": "safe"
      }
    }
  }
}
```

**Key Behavior**: With finality tags, the proxy returns `max(requested_height, finality_height)`.

## For Exchanges

### F3 Finality Levels

When integrating with exchanges, understanding F3 finality levels is crucial for balance security:

- **`finalized`**: Cryptographically final - irreversible under normal conditions (best case: ~few minutes behind head, worst case: ~900 epochs / ~7.5 hours behind head)
- **`safe`**: Conservative safety buffer - protects against minor reorganizations (~100 minutes behind head, but never older than finalized)
- **`latest`**: Chain head - no finality guarantees, may change due to forks

### Safe vs Finalized

The `safe` tag calculates: `max(finalized_height, latest_height - 200_epochs)`

This means:

- `safe` is **never older** than `SafeHeightDistance` (200 epochs by [default](https://github.com/filecoin-project/lotus/blob/08b62be9c1c0f8c8be40278b89ec44547b1592c3/build/buildconstants/params_shared_vals.go#L95-L108))
- `safe` provides 200-epoch safety buffer (~100 minutes) from chain head
- Both `safe` and `finalized` offer strong reorganization protection

### Using F3 with Block Identifiers

When querying specific block heights with finality:

```bash
# Request block 2000000 with safe finality
# Returns: max(2000000, safe_height)
curl -X POST http://localhost:8080/block \
  -d '{
    "network_identifier": {
      "blockchain": "Filecoin", "network": "mainnet",
      "sub_network_identifier": {
        "network": "f3", "metadata": {"finality_tag": "safe"}
      }
    },
    "block_identifier": {"index": 2000000}
  }'
```

**Critical**: If requested height < finality height, you get the finality height instead. This ensures you never receive non-final data when finality is requested.

### Exchange Recommendations

- **Balance Queries**: Use `safe` for recent but stable data
- **Block Scanning**: Use `safe` to avoid rescanning due to minor reorgs
- **Latest State**: Only use `latest` when immediate data is required and reorgs are acceptable

## API Endpoints

### Network

- `POST /network/list` - Available networks
- `POST /network/status` - Network status and sync state
- `POST /network/options` - Supported operations

### Account

- `POST /account/balance` - Account balance with optional finality

```bash
# Basic balance
curl -X POST http://localhost:8080/account/balance \
  -H "Content-Type: application/json" \
  -d '{
    "network_identifier": {"blockchain": "Filecoin", "network": "mainnet"},
    "account_identifier": {"address": "f1abc123..."}
  }'
```

### Block

- `POST /block` - Block data with optional finality

### Mempool

- `POST /mempool` - Pending transactions
- `POST /mempool/transaction` - Specific pending transaction

### Construction

- `POST /construction/metadata` - Transaction metadata and gas estimates
- `POST /construction/submit` - Submit signed transaction

## Multisig Accounts

Query specific balance types using `sub_account`:

```json
{
  "account_identifier": {
    "address": "f2multisig...",
    "sub_account": { "address": "LockedBalance" }
  }
}
```

Available sub-account types: `LockedBalance`, `SpendableBalance`, `VestingSchedule`

## Supported Operations

Core: `Send`, `Fee`, `AddBalance`, `Exec`
Multisig: `SwapSigner`, `Propose`, `Approve`, `Cancel`
Mining: `PreCommitSector`, `ProveCommitSector`, `SubmitWindowedPoSt`, `AwardBlockReward`
EVM: `InvokeContract`, `EVM_CALL`

See `rosetta/services/constants.go` for the complete list.

## Development

```bash
# Test
make test

# Lint
make lint

# Generate mocks
make generate_mocks LOTUS_DIR=/path/to/lotus LOTUS_VERSION=v1.32.0
```

## Troubleshooting

**FFI Issues**: `make gitclean && make build`
**Connection Issues**: Verify `LOTUS_RPC_URL` accessibility and `LOTUS_RPC_TOKEN` permissions
**Sync Issues**: Proxy requires synced Lotus node - check `/network/status`
**F3 Issues**: Ensure Lotus node supports V2 APIs when `ENABLE_LOTUS_V2_APIS=true`

## Transaction Construction Example

```go
// 1. Get metadata
metadataReq := &types.ConstructionMetadataRequest{
    NetworkIdentifier: network,
    Options: map[string]interface{}{
        "idSender": "f1abc...", "idReceiver": "f1def...", "value": "1000000",
    },
}
metadata, _, _ := client.ConstructionAPI.ConstructionMetadata(ctx, metadataReq)

// 2. Build transaction with rosetta-filecoin-lib
r := rosettaFilecoinLib.NewRosettaConstructionFilecoin(nil)
txJSON, _ := r.ConstructPayment(&rosettaFilecoinLib.PaymentRequest{
    From: "f1abc...", To: "f1def...", Quantity: "1000000",
    Metadata: rosettaFilecoinLib.TxMetadata{/* from metadata response */},
})

// 3. Sign and submit
signedTxJSON, _ := r.SignTxJSON(txJSON, privateKey)
submitReq := &types.ConstructionSubmitRequest{
    NetworkIdentifier: network, SignedTransaction: signedTxJSON,
}
response, _, _ := client.ConstructionAPI.ConstructionSubmit(ctx, submitReq)
```

Complete example in `rosetta/examples/create_transaction.go`.
