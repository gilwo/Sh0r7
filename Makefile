all: build-web build

web: build-web

build-web: webapp/common/common.go webapp/front/front_main.go webapp/frontend/front.go webapp/frontend/sh0r7.go 
	GOOS=js GOARCH=wasm go build -o web/app.wasm webapp/front/front_main.go

build:
	go build -tags webapp main.go

run: run-local

run-local: build-web
	go run -tags webapp main.go -webapp -local

deploy: build-web
	go build -tags netgo,webapp -ldflags '-s -w' -o app
