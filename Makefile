# version-file := .version
# DEPLOY_VERSIOM = $(shell cat ${version-file})

ifndef VERSION
    VERSION = `[ -e .version ] && cat .version || git describe --always --long --dirty`
endif

# https://wiki.debian.org/ReproducibleBuilds/TimestampsProposal
ifndef SOURCE_DATE_EPOCH
    SOURCE_DATE_EPOCH = `git log -1 --format=%ct`
    SOURCE_DATE_EPOCH2 = `git diff --quiet && git log -1 --format=%ct || date +%s*`
endif


BUILD_TIME_EPOCH = $(shell date +%s)
# GIT_COMMIT := $(shell git log --pretty=format:"%h" -n 1 webapp/common webapp/front webapp/frontend/)
# BUILD_TIME := $(shell date)
# BUILD_TIME2 := $(shell ${PWD}/tsec2)


LDFLAGS=-ldflags \
	"-X 'github.com/gilwo/Sh0r7/common.BuildVersion=${VERSION}' \
	-X 'github.com/gilwo/Sh0r7/common.SourceTime=${SOURCE_DATE_EPOCH2}' \
	-X 'github.com/gilwo/Sh0r7/common.BuildTime=${BUILD_TIME_EPOCH}'"

all: build-web build

echo-version:
	@echo SOURCE_DATE_EPOCH: ${SOURCE_DATE_EPOCH}
	@echo SOURCE_DATE_EPOCH2: ${SOURCE_DATE_EPOCH2}
	@echo VERSION: ${VERSION}
	@echo BUILD_TIME_EPOCH: ${BUILD_TIME_EPOCH}
	@echo DEPLOY_VERSIOM: ${DEPLOY_VERSIOM}


web: build-web

all2: build-web2
	go run -tags webapp main.go -webapp -local

webprod: build-web-prod

build-web: echo-version
	GOOS=js GOARCH=wasm go build ${LDFLAGS} \
	-o web/app.wasm webapp/front/front_main.go

build-web2:
	@export GIT_COMMIT=$(git log --pretty=format:"%h" -n 1 webapp/common webapp/front webapp/frontend/)
	GOOS=js GOARCH=wasm tinygo build -ldflags "-X 'main.BB=${GIT_COMMIT}' -X 'github.com/gilwo/Sh0r7/webapp/frontend.BuildVer=${GIT_COMMIT}'" -o web/app.wasm


build-web-prod: echo-version
	GOOS=js GOARCH=wasm go build ${LDFLAGS} \
	-tags prod -o web/app.wasm webapp/front/front_main.go

build:
	go build -tags webapp main.go

run: run-local

run-local: build-web
	go run ${LDFLAGS}  \
	-tags webapp main.go -webapp -local

run-redis: build-web
	go run ${LDFLAGS}  \
	-tags webapp,redis main.go -webapp 

get-version:
ifeq ($(origin RENDER_GIT_COMMIT), environment)
	mkdir _build
	cd _build
	git init
	git remote add origin https://github.com/gilwo/Sh0r7
	git fetch --depth 1 origin ${RENDER_GIT_COMMIT}
	git fetch --prune --unshallow
	git describe --always --long --dirty > ../.version
	rm -rf _build
	cat ../.version
else
	@echo get-version VERSION: ${VERSION}
endif

deploy: get-version build-web-prod
	go build -tags netgo,webapp,prod,redis -ldflags '-s -w' ${LDFLAGS} -o app

build-web-test:
	GOOS=js GOARCH=wasm go build -tags prod -ldflags '-s -w' -o web/app.wasm webapp/front/front_main.go

run-prod-local: deploy
	./app -webapp -local

rundev:
	SH0R7_DEPLOY="localdev" \
	_SH0R7__DEV_ENV="true" \
	SH0R7_DEV_HOST=$(shell hostname) \
	_SH0R7_OTEL_UPTRACE=https://6vG7FMQMnHKEilyqybbAeg@uptrace.dev/1374 \
	SH0R7_ADMIN_KEY=pass \
	make run #-redis


deploy-dev-live:
	curl https://api.render.com/deploy/srv-ce8h2varrk03sibof1v0?key=52S8DAn6biY