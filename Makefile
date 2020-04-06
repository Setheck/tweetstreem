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
	-X ${BASE_PKG}/app.AppToken=${APP_TOKEN} \
	-X ${BASE_PKG}/app.AppSecret=${APP_SECRET}"

test:
	go test ./... -cover

build: test
	@if [ -z "${APP_TOKEN}" ]; then echo "APP_TOKEN Not Set"; exit 1; fi
	@if [ -z "${APP_SECRET}" ]; then echo "APP_SECRET Not Set"; exit 1; fi
	go build ${LDFLAGS} -o tweetstreem

buildmac: test
	@if [ -z "${APP_TOKEN}" ]; then echo "APP_TOKEN Not Set"; exit 1; fi
	@if [ -z "${APP_SECRET}" ]; then echo "APP_SECRET Not Set"; exit 1; fi
	GOOS=darwin go build ${LDFLAGS} -o tweetstreem_mac

buildarm: test
	@if [ -z "${APP_TOKEN}" ]; then echo "APP_TOKEN Not Set"; exit 1; fi
	@if [ -z "${APP_SECRET}" ]; then echo "APP_SECRET Not Set"; exit 1; fi
	GOOS=linux GOARCH=arm go build ${LDFLAGS} -o tweetstreem_arm

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
tag: MINOR=1
tag: PATCH=0
tag:
	git tag "${MAJOR}.${MINOR}.${PATCH}"
	git push origin --tags

ddeploy: clean dbuild
	docker tag ${IMAGE}:latest ${IMAGE}:${VERSION}
	docker push ${IMAGE}:latest
	docker push ${IMAGE}:${VERSION}

clean:
	rm -rf tweetstreem*

.PHONY: test build dbuild clean tag
