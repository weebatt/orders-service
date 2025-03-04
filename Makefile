BUILD_DIR := bin

default: all

all: clean build run

.PHONY: $(BUILD_DIR)
bin/main: main/main.go
	GOMEMLIMIT=100 GOGC=100 go build -ldflags="-s -w" -o ./bin/main -v main/main.go

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	@go mod tidy

.PHONY: build
build: bin/main

.PHONY: run
run: build
	bin/main

.PHONY: productProto
productProtoGateway:
	protoc --go_out=./pkg/api/test --go_opt=paths=source_relative --go-grpc_out=./pkg/api/test  --go-grpc_opt=paths=source_relative --grpc-gateway_out=./pkg/api/test --grpc-gateway_opt=paths=source_relative ./proto/api/*proto

.PHONY: proto
proto: productProto