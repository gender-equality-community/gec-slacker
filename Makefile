IMG ?= ghcr.io/gender-equality-community/gec-slacker
TAG ?= latest
LOGLVL ?= INFO

default: app

app: *.go go.mod go.sum
	go build -o $@ -ldflags="-s -w -linkmode=external -X main.LogLevel=$(LOGLVL)" -buildmode=pie -trimpath
	upx $@

.PHONY: docker-build docker-push
docker-build:
	docker build --build-arg logLevel=$(LOGLVL) -t $(IMG):$(TAG) .

docker-push:
	docker push $(IMG):$(TAG)

.image:
	echo $(IMG):$(TAG) > $@
