.PHONY: fmt
fmt:
	goimports -l -w .

.PHONY: proto
proto:
	@protoc --proto_path=api/protobuf/ api/protobuf/ultron.proto \
	--go_out=pkg/rpc \
	--go_opt=paths=source_relative \
	--go-grpc_opt=require_unimplemented_servers=false \
	--go-grpc_out=pkg/rpc \
	--go-grpc_opt=paths=source_relative
