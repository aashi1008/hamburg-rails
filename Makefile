run:
	go run ./cmd/service

test:
	go test ./...

lint:
	golangci-lint run

fuzz:
	go test -fuzz=Fuzz -fuzztime=10s ./internal/graph

docker-build:
	docker build -t hamburg-rails .
