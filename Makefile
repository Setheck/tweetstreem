VERSION:=$(shell git describe --tags)
COMMIT:=$(shell git rev-parse HEAD)
BUILT:=$(shell date +%FT%T%z)
BASE_PKG:=github.com/Setheck/tweetstream
IMAGE:=setheck/tweetstream

LDFLAGS=-ldflags "-w -s \
	-X ${BASE_PKG}/app.Version=${VERSION} \
	-X ${BASE_PKG}/app.Built=${BUILT} \
	-X ${BASE_PKG}/app.Commit=${COMMIT} \
	-X ${BASE_PKG}/app.AppToken=${APP_TOKEN} \
	-X ${BASE_PKG}/app.AppSecret=${APP_SECRET}"

test:
	go test ./... -cover

build: test
	if [[ -z "${APP_KEY}" ]]; then @echo "App Key Not Set"; fi
	go build ${LDFLAGS} .

dbuild:
	# *Note, docker file calls `make build`
	docker build . -t ${IMAGE}:latest
	docker run --rm ${IMAGE}:latest -version

tag: MAJOR=0
tag: MINOR=1
tag: PATCH=0
tag:
	git tag "${MAJOR}.${MINOR}.${PATCH}"
	git push origin --tags

deploy: clean dbuild
	docker tag ${IMAGE}:latest ${IMAGE}:${VERSION}
	docker push ${IMAGE}:latest
	docker push ${IMAGE}:${VERSION}

clean:
	rm -rf tweetstream

.PHONY: test build dbuild clean tag