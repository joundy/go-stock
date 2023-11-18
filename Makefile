build:
	go build ./cmd/main.go

run:
	GRPC_ADDRESS=127.0.0.1:9090 \
	HTTP_ADDRESS=127.0.0.1:3001 \
	REDPANDA_ADDRESS=127.0.0.1:9092 \
	REDIS_ADDRESS=127.0.0.1:6379 \
	REDIS_AUTH=eYVX7EwVmmxKPCDmwMtyKVge8oLd2t81 \
	go run ./cmd/main.go

test: 
	go test ./...

proto-gen:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	proto/*.proto

mock-gen: 
	mockery --dir=ohlc --output=./ohlc/mocks --outpkg=mocks --all

run-docker-panda-redis:
	docker compose up redpanda redis

run-docker-full:
	docker compose up --build
