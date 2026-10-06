.PHONY: proto test run

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/snowflake.proto

test:
	go test ./... -v
