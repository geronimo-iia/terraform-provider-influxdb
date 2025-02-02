export TF_ACC_TERRAFORM_VERSION=1.5.0

TEST?=$$(go list ./... |grep -v 'vendor')
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

PKG_NAME=influxdb

default: build

build:
	go install

.PHONY: dev-setup
dev-setup: ## setup development dependencies
	@which ./bin/golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v1.53.3


.PHONY: dev-cleanup
dev-cleanup: ## cleanup development dependencies
	rm -rf bin/*

.PHONY: mod
mod: ## add missing and remove unused modules
	go mod tidy -compat=1.20

.PHONY: lint-check
lint-check: ## Run static code analysis and check formatting
	./bin/golangci-lint run ./... -v

.PHONY: lint-fix
lint-fix: ## Run static code analysis, check formatting and try to fix findings
	./bin/golangci-lint run ./... -v --fix

# go generate ./...
test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | \
		xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: fmtcheck
	TF_ACC=1 \
	INFLUXDB_USERNAME=test INFLUXDB_PASSWORD=test \
	go test -v ./... -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

test-compile:
	@if [ "$(TEST)" = "./..." ]; then \
		echo "ERROR: Set TEST to a specific package. For example,"; \
		echo "  make test-compile TEST=./$(PKG_NAME)"; \
		exit 1; \
	fi
	go test -c $(TEST) $(TESTARGS)
