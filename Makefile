build-docker:
	docker build -t matthieudolci/hatcher .
.PHONY: build-docker

build:
	go build -installsuffix cgo -o hatcher
.PHONY: build

push:
	docker push matthieudolci/hatcher:latest
.PHONY: push
