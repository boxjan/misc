.PHONY:help
help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

CMDS=$(shell ls cmd)
COMMIT=$(shell git rev-parse --short HEAD)
REV=$(shell git describe --dirty 2>/dev/null || echo "dev-$(COMMIT)"; )
BUILD_TIME=$(shell TZ="UTC" date +'%Y-%m-%dT%H:%M:%S')
GO_VERSION=$(shell go version | cut -d ' ' -f3)

# Add go ldflags using LDFLAGS at the time of compilation.
BUILD_INFO_LDFLAGS = -X github.com/boxjan/misc/commom/cmd.version=$(REV) \
-X github.com/boxjan/misc/commom/cmd.buildTime=$(BUILD_TIME)

EXT_LDFLAGS = -extldflags "-static"
LDFLAGS =
FULL_LDFLAGS = $(LDFLAGS) $(BUILD_INFO_LDFLAGS) $(EXT_LDFLAGS)

BUILD_PLATFORMS ?=linux/amd64,linux/arm64

.PHONY: build-% build
build: ## build bin
build: $(CMDS:%=build-%)
$(CMDS:%=build-%): build-%:
	mkdir -p bin
	echo '$(BUILD_PLATFORMS)' | tr '/' ' ' | tr ',' '\n' | while read -r os arch suffix; do \
		if ! (set -x; CGO_ENABLED=0 GOOS="$$os" GOARCH="$$arch" go build $(GOFLAGS_VENDOR) -a -ldflags '$(FULL_LDFLAGS)' -o ./bin/"$*_$$os"_"$$arch" ./cmd/$*/); then \
			echo "Building $* for GOOS=$$os GOARCH=$$arch failed, see error(s) above."; \
			exit 1; \
		fi; \
	done

REGISTRY ?= boxjan/misc

.PHONY: container-% container
container: ## Package container
container: $(CMDS:%=cdontainer-%)
$(CMDS:%=container-%): container-%: build-%
	set -x; \
	echo '$(BUILD_PLATFORMS)' | tr '/' ' ' | tr ',' '\n' | while read -r os arch suffix; do \
		docker build -t $(REGISTRY)-$*:$(REV)-$$arch -t $(REGISTRY)-$*:latest-$$arch \
		--build-arg os=$$os --build-arg arch=$$arch --platform $$os/$$arch \
		-f $(shell if [ -e ./cmd/$*/Dockerfile ]; then echo ./cmd/$*/Dockerfile; else echo Dockerfile; fi) . ; \
	done; \
	docker manifest create $(REGISTRY)-$*:$(REV) $(shell echo '$(BUILD_PLATFORMS)' | tr '/' ' ' | tr ',' '\n' | while read -r os arch suffix; do echo $(REGISTRY)-$*:$(REV)-"$$arch"; done); \
    docker manifest create $(REGISTRY)-$*:latest $(shell echo '$(BUILD_PLATFORMS)' | tr '/' ' ' | tr ',' '\n' | while read -r os arch suffix; do echo $(REGISTRY)-$*:latest-"$$arch"; done); \
	docker manifest push $(REGISTRY)-$*:latest;\
	docker manifest push $(REGISTRY)-$*:$(REV)

.PHONY: test
test: ## run all test case
	@ echo; echo "### $@:"
	go test $(GOFLAGS_VENDOR) `go list $(GOFLAGS_VENDOR) ./... | grep -v -e 'vendor' -e '/test/e2e$$' $(TEST_GO_FILTER_CMD)` $(TESTARGS)