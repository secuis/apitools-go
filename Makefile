proto:
	@protoc \
		--go_out paths=source_relative:. \
		--go-grpc_out paths=source_relative:. \
		--grpc-gateway_out paths=source_relative:. \
		dbmigration/dbmigration.proto

install:
	@go install \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway \
    github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2 \
    google.golang.org/protobuf/cmd/protoc-gen-go \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc