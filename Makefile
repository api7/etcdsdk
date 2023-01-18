
### lint:		Lint Go source codes
.PHONY: lint
lint: ## Run the golangci-lint application (install if not found)
	@#Brew - MacOS
	@if [ "$(shell command -v golangci-lint)" = "" ] && [ "$(shell command -v brew)" != "" ] && [ "$(UNAME)" = "Darwin" ]; then brew install golangci-lint; fi;
	@#has sudo
	@if [ "$(shell command -v golangci-lint)" = "" ]; then curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.49.0 && sudo cp ./bin/golangci-lint $(go env GOPATH)/bin/; fi;
	@echo "running golangci-lint..."
	@golangci-lint run --tests=false ./... --timeout 5m

### test:		Run unit test
.PHONY: test
test:
	@go test -race -cover -coverprofile=coverage.txt ./...
