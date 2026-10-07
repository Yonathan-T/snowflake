.PHONY: proto test bench server client stress up down

proto:
	protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/snowflake/snowflake.proto

test:
	go test ./... -v

bench:
	go test -bench=BenchmarkNextID -benchmem .\pkg\snowflake

server:
	go run cmd/server/main.go --port=50051

client:
	go run cmd/client/main.go

stress:
	go run cmd/stress/main.go --workers=100 --requests=1000

up:
	docker compose up -d

down:
	docker compose down
