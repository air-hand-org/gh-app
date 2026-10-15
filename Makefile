.PHONY: setup
setup:
	mise i
	aqua i

.PHONY: build
build:
	go mod tidy
	go generate ./...
	go build -o bin/gh-app ./

.PHONY: fmt
fmt:
	golangci-lint fmt ./...

.PHONY: lint
lint:
	golangci-lint fmt --diff ./...
	golangci-lint run ./...
	zizmor ./

.PHONY: test
test: build
	go test -race -cover ./...
	./bin/gh-app --help
	./bin/gh-app --version

.PHONY: check-clean
check-clean:
	git status --short
	test -z "$$(git status --porcelain)"

.PHONY: release-check
release-check:
	goreleaser check

.PHONY: release-snapshot
release-snapshot: release-check
	goreleaser release --snapshot --clean

.PHONY: ci
ci: build lint test release-snapshot check-clean
