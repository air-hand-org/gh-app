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
	zizmor -c .github/linters/zizmor.yaml ./

.PHONY: test
test: build
	go test -race -cover ./...
	./bin/gh-app --help
	./bin/gh-app --version

.PHONY: check-clean
check-clean:
	git status --short
	test -z "$$(git status --porcelain)"

# Regenerate versioned documents after changing dependencies or the collector.
.PHONY: licenses
licenses:
	rm -rf third_party/licenses
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go-licenses save ./ --ignore github.com/air-hand-org/gh-app --save_path third_party/licenses/amd64
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go-licenses save ./ --ignore github.com/air-hand-org/gh-app --save_path third_party/licenses/arm64

.PHONY: release-check
release-check:
	goreleaser check

.PHONY: release-snapshot
release-snapshot: release-check
	goreleaser release --snapshot --clean

.PHONY: ci
ci: build lint test release-snapshot check-clean
