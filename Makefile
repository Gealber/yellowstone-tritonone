GO_BIN := $(shell go env GOPATH)/bin
PATH := $(GO_BIN):$(PATH)

# Include dir that ships vtproto/ext.proto (the (vtproto.mempool) extension).
VTPROTO_INCLUDE := $(shell go list -m -f '{{.Dir}}' github.com/planetscale/vtprotobuf)/include

.PHONY: protoc protoc-tools

protoc:
	$(MAKE) protoc-tools
	protoc \
		--go_out=./proto \
		--go_opt=paths=source_relative \
		--go-grpc_out=./proto \
		--go-grpc_opt=paths=source_relative \
		-I./proto \
		-I$(VTPROTO_INCLUDE) \
		./proto/*.proto
	protoc \
		--go-vtproto_out=./proto \
		--go-vtproto_opt=paths=source_relative \
		--go-vtproto_opt=features=marshal+unmarshal+size+pool \
		-I./proto \
		-I/usr/include \
		-I$(VTPROTO_INCLUDE) \
		./proto/*.proto

protoc-tools:
	@command -v protoc-gen-go >/dev/null 2>&1 || go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.11
	@command -v protoc-gen-go-grpc >/dev/null 2>&1 || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.6.1
	@command -v protoc-gen-go-vtproto >/dev/null 2>&1 || go install github.com/planetscale/vtprotobuf/cmd/protoc-gen-go-vtproto@latest

