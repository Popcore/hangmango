EXECUTABLE=hangmango
PACKAGE=github.com/Popcore/hangmango

BUILD_DIR=build
ARTEFACT_DIR=artefacts

all: help

.PHONY: build
build: ## builds executable
	@echo "==> building executable"
	@mkdir -p build/
	go build -o $(BUILD_DIR)/$(EXECUTABLE) $(PACKAGE)

.PHONY: test
test: ## runs tests
	@echo "==> running tests"
	@mkdir -p $(ARTEFACT_DIR)
	@echo 'mode: atomic' > $(ARTEFACT_DIR)/coverage.out
	@go test ./... -coverprofile=$(ARTEFACT_DIR)/coverage.out
	@go tool cover -html=$(ARTEFACT_DIR)/coverage.out -o $(ARTEFACT_DIR)/coverage.html

.PHONY: help
help: ## shows this help message
	@echo 'usage: make [target] ...'
	@echo
	@echo 'targets:'
	@echo
	@echo "$$(grep -hE '^\S+:.*##' $(MAKEFILE_LIST) | sed -e 's/:.*##\s*/:/' -e 's/^\(.\+\):\(.*\)/\\x1b[36m\1\\x1b[m:\2/' | column -c2 -t -s : | sed -e 's/^/  /')"
