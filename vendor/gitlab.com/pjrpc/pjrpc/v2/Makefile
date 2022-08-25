.PHONY: test coverage lints lints_fix gogen

test:
	go test -cover ./...
	cd ./cmd/genpjrpc/ && make test

coverage:
	go test -coverpkg=./... -coverprofile=profile.cov ./...
	go tool cover -func profile.cov

	# last line will be number of total coverage
	go tool cover -func profile.cov | tail -n1 | awk '{print $$3}'

lints:
	golangci-lint version
	golangci-lint run -c ./.linters.yml
	cd ./cmd/genpjrpc && golangci-lint run -c ../../.linters.yml

lints_fix:
	golangci-lint run --fix -c ./.linters.yml
	cd ./cmd/genpjrpc && golangci-lint run --fix -c ../../.linters.yml

gogen:
	cd ./_examples/helloworld/ && go generate ./...
	cd ./_examples/middlewares/ && go generate ./...
	cd ./_examples/wasm/ && go generate ./...
