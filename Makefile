ifndef VERSION
    VERSION = `[ -e .version ] && cat .version || git describe --always --long --dirty`
endif

# https://wiki.debian.org/ReproducibleBuilds/TimestampsProposal
ifndef SOURCE_DATE_EPOCH
    SOURCE_DATE_EPOCH = `git log -1 --format=%ct`
    SOURCE_DATE_EPOCH2 = `git diff --quiet && git log -1 --format=%ct || date +%s*`
endif


BUILD_TIME_EPOCH = $(shell date +%s)

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


web: build-web

webprod: build-web-prod

build-web: echo-version
	GOOS=js GOARCH=wasm go build ${LDFLAGS} \
	-o web/app.wasm webapp/front/front_main.go

build-web-prod: echo-version
	GOOS=js GOARCH=wasm go build ${LDFLAGS} \
	-tags prod -o web/app.wasm webapp/front/front_main.go

build:
	go build -tags webapp main.go

run: run-local

run-local: build-web
	go run ${LDFLAGS}  \
	-tags webapp main.go -webapp -local

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

run-prod-local: deploy
	./app -webapp -local
