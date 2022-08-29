all: generate golangci

golangci:
	if [ -z "$(shell which golangci-lint)" ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
			sh -s -- -b $(shell go env GOPATH)/bin v1.48.0; \
	fi
	golangci-lint version
	golangci-lint -v run --timeout 5m

vendor:
	go mod vendor

generate:
	cd togenerate && go generate && cd ..

.PHONY: golangci vendor generate
