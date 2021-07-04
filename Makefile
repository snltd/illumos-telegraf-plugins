test: lint
	go test ./...

lint:
	golangci-lint run ./...

build:
	@echo This software cannot be built standalone.
	@echo Please look at the README for build instructions.
