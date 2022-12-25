all: build-web build

web: build-web

webprod: build-web-prod

build-web:
	GOOS=js GOARCH=wasm go build -o web/app.wasm webapp/front/front_main.go

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