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
	go test -race -covermode=atomic -v -coverprofile=coverage.txt ./... || exit 1;
	for dir in `find . -type f -name "go.mod" -exec dirname {} \;`; do \
		if [ $$dir != "." ]; then \
			cd $$dir; \
			go test -race -covermode=atomic -v -coverprofile=coverage.txt ./... || exit 1; \
			cd - > /dev/null ;\
			lines=`cat $$dir/coverage.txt | wc -l`; \
			lines=`expr $$lines - 1`; \
			tail -n $$lines $$dir/coverage.txt >> coverage.txt; \
		fi; \
	done

.PHONY: benchmark
benchmark: 
	for dir in `find . -type f -name "go.mod" -exec dirname {} \;`; do \
		cd $$dir; \
		go test -bench=. -run=^Benchmark ./...; \
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
	for dir in `find . -type f -name "go.mod" -exec dirname {} \;`; do \
		cd $$dir; \
		go mod download; \
		cd - > /dev/null; \
	done

.PHOY: sync-module-version
sync-module-version:
	for dir in `find . -type f -name "go.mod" -exec dirname {} \;`; do \
		head -n 1 $$dir/go.mod | grep github.com/wosai/ultron/v2$ && continue; \
		cd $$dir; \
		go get github.com/wosai/ultron/v2@${version}; \
		go mod tidy; \
		cd - > /dev/null; \
	done