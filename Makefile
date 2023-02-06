ifndef VERSION
    VERSION = $(shell git describe --always --long --dirty)
endif

# https://wiki.debian.org/ReproducibleBuilds/TimestampsProposal
ifndef SOURCE_DATE_EPOCH
    SOURCE_DATE_EPOCH = `git log -1 --format=%ct`
    SOURCE_DATE_EPOCH2 = `git diff --quiet && git log -1 --format=%ct || date +%s*`
endif


GIT_COMMIT := $(shell git log --pretty=format:"%h" -n 1 webapp/common webapp/front webapp/frontend/)
BUILD_TIME := $(shell date)

all: build-web build

test:
	@echo SOURCE_DATE_EPOCH: ${SOURCE_DATE_EPOCH}
	@echo SOURCE_DATE_EPOCH2: ${SOURCE_DATE_EPOCH2}
	@echo VERSION: ${VERSION}


web: build-web

webprod: build-web-prod

build-web:
	GOOS=js GOARCH=wasm go build -ldflags \
	"-X 'github.com/gilwo/Sh0r7/webapp/frontend.BuildVer=${VERSION}' \
	-X 'github.com/gilwo/Sh0r7/webapp/frontend.BuildTime=${SOURCE_DATE_EPOCH2}' \
	-X 'github.com/gilwo/Sh0r7/webapp/frontend.ExternalTimeBuild=${SOURCE_DATE_EPOCH2}'" \
	-o web/app.wasm webapp/front/front_main.go
	@export GIT_COMMIT=$(git log --pretty=format:"%h" -n 1 webapp/common webapp/front webapp/frontend/)
	GOOS=js GOARCH=wasm go build -ldflags "-X 'github.com/gilwo/Sh0r7/webapp/frontend.BuildVer=${GIT_COMMIT}'" -o web/app.wasm webapp/front/front_main.go

build-web-prod:
	GOOS=js GOARCH=wasm go build -tags prod -o web/app.wasm webapp/front/front_main.go

build:
	go build -tags webapp main.go

run: run-local

run-local: build-web
	go run -tags webapp main.go -webapp -local

deploy: build-web-prod
	go build -tags netgo,webapp,prod,redis -ldflags '-s -w' -o app

run-prod-local: deploy
	./app -webapp -local
