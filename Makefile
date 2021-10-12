.PHONY: fmt
fmt:
	goimports -l -w .

#.PHONY: goproxy
#goproxy:
#	export GOPROXY=https://goproxy.cn,direct

#.PHONY: pb-tool
#pb-tool: goproxy
#	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
#	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1

.PHONY: proto
proto:
	@protoc --proto_path=api/protobuf/ api/protobuf/ultron.proto \
	--go_out=pkg/service \
	--go_opt=paths=source_relative \
	--go-grpc_opt=require_unimplemented_servers=false \
	--go-grpc_out=pkg/service \
	--go-grpc_opt=paths=source_relative
