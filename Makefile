.PHONY: build
build:
	go build -o ./builds/server server/*.go
	go build -o ./builds/client client/*.go