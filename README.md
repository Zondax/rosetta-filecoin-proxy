# Rosetta integration for Filecoin (Proxy)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![CircleCI](https://circleci.com/gh/Zondax/rosetta-filecoin-proxy/tree/master.svg?style=shield)](https://circleci.com/gh/Zondax/rosetta-filecoin-proxy/tree/master)

This command will run a Lotus node, and a rosetta-proxy instance listening en port 8080 (this will also build the container if you haven't yet):
```
make run
```

You can rebuild the container by using 
```
make rebuild_docker
```

Once the docker container is up, you can log in into it:
```
make login
```

For only building rosetta-filecoin-proxy, run:
```
make build
```