#!make
-include env.conf

VERSION:=$(shell git describe --tags)
COMMIT:=$(shell git rev-parse HEAD)
BUILT:=$(shell date +%FT%T%z)
BASE_PKG:=github.com/Setheck/tweetstreem
IMAGE:=setheck/tweetstreem

LDFLAGS=-ldflags "-w -s \
	-X ${BASE_PKG}/app.Version=${VERSION} \
	-X ${BASE_PKG}/app.Built=${BUILT} \
	-X ${BASE_PKG}/app.Commit=${COMMIT} \
	-X ${BASE_PKG}/twitter.AppToken=${APP_TOKEN} \
	-X ${BASE_PKG}/twitter.AppSecret=${APP_SECRET}"

test:
	go test ./... -cover -v -race

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

build: GOOS=linux
build:
	GOOS=$(GOOS) go build ${LDFLAGS} -o tweetstreem

buildwin: GOOS=windows
buildwin:
	GOOS=$(GOOS) go build ${LDFLAGS} -o tweetstreem_win.exe

buildmac: GOOS=darwin
buildmac:
	GOOS=$(GOOS) go build ${LDFLAGS} -o tweetstreem_mac

buildarm: GOOS=linux
buildarm: GOARCH=arm
buildarm:
	GOOS=$(GOOS) go build ${LDFLAGS} -o tweetstreem_arm

tokencheck:
	@if [ -z "${APP_TOKEN}" ]; then echo "APP_TOKEN Not Set"; exit 1; fi
	@if [ -z "${APP_SECRET}" ]; then echo "APP_SECRET Not Set"; exit 1; fi

package:
	mkdir -p deploy/
	mv tweetstreem* deploy/
	tar -czf tweetstreem.tar.gz deploy

dbuild:
	exit 1  #no docker for now
	# *Note, docker file calls `make build`
	docker build . -t ${IMAGE}:latest
	docker run --rm ${IMAGE}:latest -version

tag: MAJOR=0
tag: MINOR=3
tag: PATCH=0
tag:
	git tag "v${MAJOR}.${MINOR}.${PATCH}"
	git push origin --tags

ddeploy: clean dbuild
	docker tag ${IMAGE}:latest ${IMAGE}:${VERSION}
	docker push ${IMAGE}:latest
	docker push ${IMAGE}:${VERSION}

clean:
	rm -rf tweetstreem* deploy

.PHONY: test build dbuild clean tag tokencheck
