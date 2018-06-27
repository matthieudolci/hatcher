TAG ?= latest

build-docker:
	docker build -t matthieudolci/hatcher:$(TAG) .
.PHONY: build-docker

build:
	go build -installsuffix cgo -o hatcher
.PHONY: build

push:
	docker push matthieudolci/hatcher:$(TAG)
.PHONY: push
