build-docker:
	docker build -t matthieudolci/hatcher .
.PHONY: build-docker

build:
	go build -installsuffix cgo -o hatcher
.PHONY: build

push:
	docker push matthieu.dolci/hatcher:latest
.PHONY: push
