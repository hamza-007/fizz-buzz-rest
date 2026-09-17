BINARY := fizzbuzz
IMAGE  := fizzbuzz
PORT   ?= 8080

TARGETS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 linux/arm \
           windows/amd64 windows/arm64

.DEFAULT_GOAL := help

.PHONY: help
help: ## List the targets
	@grep -hE '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##/\t/' | expand -t 12

.PHONY: run
run: ## Run the server from source
	go run . serve

.PHONY: build
build: ## Build for this machine into bin/
	go build -trimpath -ldflags='-s -w' -o bin/$(BINARY) .

.PHONY: dist
dist: ## Cross-compile every target into dist/
	@for target in $(TARGETS); do \
		os=$${target%%/*}; arch=$${target##*/}; \
		out=dist/$(BINARY)_$${os}_$${arch}; \
		if [ "$$os" = windows ]; then out=$$out.exe; fi; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch GOARM=7 \
			go build -trimpath -ldflags='-s -w' -o $$out . || exit 1; \
		echo $$out; \
	done

.PHONY: fmt
fmt: ## Format the code
	gofmt -s -w .

.PHONY: lint
lint: ## Check formatting and run go vet
	@test -z "$$(gofmt -l .)" || { echo "run 'make fmt':"; gofmt -l .; exit 1; }
	go vet ./...

.PHONY: check
check: lint build ## What the CI runs

.PHONY: image
image: ## Build the container image
	docker build -t $(IMAGE) .

.PHONY: up
up: image ## Run the container on $(PORT)
	docker run --rm -p $(PORT):8080 --name $(BINARY) $(IMAGE)

.PHONY: env
env: ## Create .env from the example
	@test -f .env || cp .env.example .env

.PHONY: clean
clean: ## Remove build output
	rm -rf bin dist
