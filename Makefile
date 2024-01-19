BINARY_NAME := dockwiz
DEV_NAME := dockwiz-dev

build:
	go build -o bin/$(BINARY_NAME) -v ./cmd

docker:
	docker build -t $(BINARY_NAME) .

dev:
	docker image inspect $(DEV_NAME) > /dev/null 2>&1 || docker build --target development -t $(DEV_NAME) .

	# Remove the containers if it exists
	docker container inspect $(DEV_NAME) > /dev/null 2>&1 && docker container rm -f $(DEV_NAME) || true
	docker container inspect redis > /dev/null 2>&1 && docker container rm -f redis || true

	docker run --name redis -d redis || exit 1
	docker container run -it --rm -p 8080:8080 -v $(shell pwd):/go/src/$(BINARY_NAME) --workdir /go/src/$(BINARY_NAME) --name $(DEV_NAME) $(DEV_NAME) /bin/sh

lint:
	golangci-lint run ./...

test:
	go test -v ./... -count=1

.PHONY: build docker dev lint test