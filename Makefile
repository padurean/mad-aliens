# NOTE: VERSION can also be set when running make. Example:
# make <target> VERSION=v1.2.3

# VERSION=$(shell git describe --tags)
VERSION ?= devbuild
COMMIT ?= $(shell git rev-parse HEAD|head -c 7)
BUILT_AT = $(shell date +%s)
BIN = "./invasion"
ARGS ?= -i 5
COVERPROFILE = "coverage.txt"
COMMON_GO_TEST_FLAGS = -count=1 -failfast

DATE=$(shell date "+%Y-%m-%d")

LDFLAGS = -X "main.Version=$(VERSION)"\
	-X "main.Commit=$(COMMIT)"\
	-X "main.Buildtime=$(BUILT_AT)"

.PHONY: *

build:
	go build ./...

build_binary:
	$(info Building the invasion binary ...)
	@CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -a ./cmd/invasion

run: build_binary
	@$(BIN) $(ARGS)

run_dev:
	go run ./cmd/invasion $(ARGS)

lint:
	@golangci-lint run ./... --timeout 2m

test:
	@go test ./... $(COMMON_GO_TEST_FLAGS)

test_verbose:
	@go test ./... $(COMMON_GO_TEST_FLAGS) -v

test_race:
	@go test ./... $(COMMON_GO_TEST_FLAGS) -race

bench:
	@go test ./... $(COMMON_GO_TEST_FLAGS) -bench=.

coverage:
	@go test ./... $(COMMON_GO_TEST_FLAGS) -cover -coverpkg=./...

coverprofile:
	@go test ./... $(COMMON_GO_TEST_FLAGS) -coverprofile $(COVERPROFILE) -coverpkg=./...

coverprofile_verbose:
	@go test ./... $(COMMON_GO_TEST_FLAGS) -coverprofile $(COVERPROFILE) -coverpkg=./... -v

show_coverage: coverprofile
	@go tool cover -html $(COVERPROFILE)

show_function_coverage: coverprofile
	@go tool cover -func $(COVERPROFILE)

show_function_coverage_verbose: coverprofile_verbose
	@go tool cover -func $(COVERPROFILE)

# Runs the tests during CI/CD (i.e. without the "-race" flag)
# and reports coverage.
coverage-report-ci:
	@./coverage-report.sh skip-race

coverage-report: # Runs the tests and reports coverage.
	@./coverage-report.sh

build_docker_image:
	@docker build -f Dockerfile \
		--label=org.opencontainers.image.created=$(DATE) \
		--label=org.opencontainers.image.name=padurean/mad-aliens \
		--label=org.opencontainers.image.revision=$(COMMIT) \
		--label=org.opencontainers.image.version=$(COMMIT) \
		--label=org.opencontainers.image.source=https://github.com/padurean/mad-aliens \
		--label=repository=https://github.com/padurean/mad-aliens \
		--tag padurean/mad-aliens:$(COMMIT) \
		--tag padurean/mad-aliens:latest .

push_docker_image:
	@docker push padurean/mad-aliens --all-tags
