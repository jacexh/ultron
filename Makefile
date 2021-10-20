.PHONY: proto
proto:
	protoc --proto_path=api/protobuf/ api/protobuf/ultron.proto \
	--go_out=internal/genproto \
	--go_opt=paths=source_relative \
	--go-grpc_opt=require_unimplemented_servers=false \
	--go-grpc_out=internal/genproto \
	--go-grpc_opt=paths=source_relative

	protoc --proto_path=api/protobuf/ api/protobuf/statistics.proto \
	--go_out=pkg/statistics \
	--go_opt=paths=source_relative \
	--go-grpc_opt=require_unimplemented_servers=false \
	--go-grpc_out=pkg/statistics \
	--go-grpc_opt=paths=source_relative