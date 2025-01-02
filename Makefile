.PHONY: build test run docker-build docker-run clean

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
BINARY_NAME=notification-service
MAIN_PATH=cmd/notification/main.go

all: test build

build:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(MAIN_PATH)

test:
	$(GOTEST) -v ./...

run:
	$(GOBUILD) -o bin/$(BINARY_NAME) $(MAIN_PATH)
	./bin/$(BINARY_NAME)

clean:
	$(GOCLEAN)
	rm -f bin/$(BINARY_NAME)

docker-build:
	docker build -t notification-service .

docker-run:
	docker run -p 8080:8080 notification-service
