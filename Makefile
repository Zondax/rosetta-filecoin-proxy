.PHONY: build_docker rebuild_docker login run stop test

DOCKER_IMAGE=lotus:latest
CONTAINER_NAME=lotusnode

INTERACTIVE:=$(shell [ -t 0 ] && echo 1)
ROSETTA_PORT=8080
LOTUS_API_PORT = 1234

ifdef INTERACTIVE
INTERACTIVE_SETTING:="-i"
TTY_SETTING:="-t"
else
INTERACTIVE_SETTING:=
TTY_SETTING:=
endif

define run_docker
	docker run $(TTY_SETTING) $(INTERACTIVE_SETTING) --rm \
	-v $(shell pwd)/data:/data \
	--name $(CONTAINER_NAME) \
	-p $(ROSETTA_PORT):$(ROSETTA_PORT) \
	-p $(LOTUS_API_PORT):$(LOTUS_API_PORT) \
	$(DOCKER_IMAGE)
endef

define kill_docker
	docker kill $(CONTAINER_NAME)
endef

define login_docker
	docker exec -ti $(CONTAINER_NAME) /bin/bash
endef

build:
	go build

test:
	go test
.PHONY: test

build_docker:
	docker build -t $(DOCKER_IMAGE) .
.PHONY: build_docker

rebuild_docker:
	docker build --no-cache -t $(DOCKER_IMAGE) .
.PHONY: rebuild_docker

run: build_docker
	$(call run_docker)
.PHONY: run

login:
	$(call login_docker)
.PHONY: login

stop:
	$(call kill_docker)
.PHONY: stop