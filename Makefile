.PHONY: proto
proto: tools
	@protoc --proto_path=api/protobuf/ api/protobuf/ultron.proto \
	--go_out=pkg/genproto \
	--go_opt=paths=source_relative \
	--go-grpc_opt=require_unimplemented_servers=false \
	--go-grpc_out=pkg/genproto \
	--go-grpc_opt=paths=source_relative

	@protoc --proto_path=api/protobuf/ api/protobuf/statistics.proto \
	--go_out=pkg/statistics \
	--go_opt=paths=source_relative \
	--go-grpc_opt=require_unimplemented_servers=false \
	--go-grpc_out=pkg/statistics \
	--go-grpc_opt=paths=source_relative

.PHONY: test
test: 
	@for dir in `find . -type f -name "go.mod" -exec dirname {} \;`; do \
		cd $$dir; \
		go test -race -covermode=atomic -v -coverprofile=coverage.txt ./...; \
		cd - > /dev/null; \
	done

.PHONY: benchmark
benchmark: 
	@for dir in `find . -type f -name "go.mod" -exec dirname {} \;`; do \
		cd $$dir; \
		go test -bench -v . ./...; \
		cd - > /dev/null; \
	done

.PHONY: tools
tools:
	@go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
	@go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
	@go install github.com/cweill/gotests/gotests@v1.6.0

.PHONY: ultron
ultron:
	@go run cmd/ultron/main.go

.PHONY: gomod
gomod: 
	@for dir in `find . -type f -name "go.mod" -exec dirname {} \;`; do \
		cd $$dir; \
		go mod download; \
		cd - > /dev/null; \
	done