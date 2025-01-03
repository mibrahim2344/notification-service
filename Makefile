.PHONY: build test run docker-build docker-run clean migrate-up migrate-down migrate-force migrate-steps migrate-version migrate-create

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

## Database migration commands
migrate-up:
	go run cmd/migrate/main.go -command=up

migrate-down:
	go run cmd/migrate/main.go -command=down

migrate-force:
	go run cmd/migrate/main.go -command=force -version=$(version)

migrate-steps:
	go run cmd/migrate/main.go -command=steps -steps=$(steps)

migrate-version:
	go run cmd/migrate/main.go -command=version

## Create a new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	up_file="migrations/$${timestamp}_$${name}.up.sql"; \
	down_file="migrations/$${timestamp}_$${name}.down.sql"; \
	touch $$up_file $$down_file; \
	echo "Created migration files: $$up_file, $$down_file"
