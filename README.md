# Rosetta integration for Filecoin (Proxy)

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CircleCI](https://circleci.com/gh/Zondax/rosetta-filecoin-proxy/tree/master.svg?style=shield)](https://circleci.com/gh/Zondax/rosetta-filecoin-proxy/tree/master)
[![Github-Actions](https://github.com/Zondax/rosetta-filecoin-proxy/workflows/rosetta-cli/badge.svg)](https://github.com/Zondax/rosetta-filecoin-proxy/actions)

To build the proxy run:

```bash
make
```

If you have upgraded and you find FFI issues, try:

```bash
make gitclean
make
```

If you want to install the linter we use try:

```bash
make install_lint
make lint
```

## Forwarding request headers to Lotus

Set `FORWARD_HEADERS` to a comma-separated allowlist of header names to copy
from each incoming Rosetta request onto every Lotus RPC call made while serving
it (e.g. for caller attribution and metrics):

```bash
FORWARD_HEADERS=x-client-id,x-request-id LOTUS_RPC_URL=https://<node>/rpc ./rosetta-filecoin-proxy
```

- Empty or unset (default): nothing is forwarded and behaviour is unchanged.
- Values are attached per request; they are never set on the shared Lotus client.
- Requires an `http(s)://` `LOTUS_RPC_URL`. Websocket connections send headers
  only when the connection is opened, so startup fails if `FORWARD_HEADERS` is set
  with a `ws(s)://` URL.
- `Authorization`, `Cookie`, `Host`, `Content-*`, hop-by-hop, `Sec-WebSocket-*` and
  `(X-)Idempotency-Key` (it makes Go's HTTP client auto-replay POSTs) headers can
  never be forwarded; startup fails if one is listed.
- Allowlisted headers named in the incoming request's `Connection` header are dropped.
- Values longer than `FORWARD_HEADERS_MAX_VALUE_BYTES` (default 1024) are dropped
  and a warning naming the header (never its value) is logged. Startup fails if it
  is not a positive integer.
- Forwarded values are whatever the caller sent. Don't rely on them for access
  control or billing unless something in front of Rosetta sets or strips them.
- Actor lookups served from the in-memory cache make no Lotus call, so they
  carry no headers.
