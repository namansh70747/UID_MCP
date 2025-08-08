.PHONY: build run test clean

build:
	go mod tidy
	go build -o kubernetes-api main.go

run:
	go run main.go

test:
	chmod +x test_api.sh
	./test_api.sh

clean:
	rm -f kubernetes-api

docker-build:
	docker build -t kubernetes-api:latest .

install-deps:
	go mod download

dev:
	go run main.go