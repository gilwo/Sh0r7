GIT_COMMIT := $(shell git log --pretty=format:"%h" -n 1 webapp/common webapp/front webapp/frontend/)
BUILD_TIME := $(shell date)

all: build-web build

web: build-web

webprod: build-web-prod

build-web:
	GOOS=js GOARCH=wasm go build -ldflags \
	"-X 'github.com/gilwo/Sh0r7/webapp/frontend.BuildVer=${GIT_COMMIT}' \
	-X 'github.com/gilwo/Sh0r7/webapp/frontend.BuildTime=${BUILD_TIME}'" \
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
	go build -tags netgo,webapp,prod -ldflags '-s -w' -o app

run-prod-local: deploy
	./app -webapp -local