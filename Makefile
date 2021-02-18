proto:
	@protoc \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		dbmigration/dbmigration.proto
	@protoc \
		--go_out=paths=source_relative:. \
		--go-grpc_out=paths=source_relative:. \
		diagnostic/diagnostic.proto

install:
	@go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc