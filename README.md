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
