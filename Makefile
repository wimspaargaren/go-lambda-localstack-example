.PHONY: zip terraform localstack test test-integration lint gofumpt gci ci-init ci-test ci-test-integration

zip:
	@CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o bootstrap ./cmd/api
	@zip infra/terraform/bootstrap.zip bootstrap
	@rm bootstrap

terraform:
	@tflocal -chdir=infra/terraform init
	@tflocal -chdir=infra/terraform apply --auto-approve

localstack:
	@docker compose build
	@docker compose up -d

test:
	@mkdir -p reports
	@go test -coverprofile=reports/codecoverage_all.cov ./... -cover -race -p=4
	@go tool cover -func=reports/codecoverage_all.cov > reports/functioncoverage.out
	@go tool cover -html=reports/codecoverage_all.cov -o reports/coverage.html
	@echo "View report at $(PWD)/reports/coverage.html"
	@tail -n 1 reports/functioncoverage.out

test-integration: | zip
	@mkdir -p reports
	@go test -coverprofile=reports/codecoverage_all.cov ./... --tags=integration -cover -race -p=4 -v
	@go tool cover -func=reports/codecoverage_all.cov > reports/functioncoverage.out
	@go tool cover -html=reports/codecoverage_all.cov -o reports/coverage.html
	@echo "View report at $(PWD)/reports/coverage.html"
	@tail -n 1 reports/functioncoverage.out

$(GOBIN)/golangci-lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1

lint: | $(GOBIN)/golangci-lint
	@echo Linting...
	@golangci-lint  -v --concurrency=3 --config=.golangci.yml --issues-exit-code=1 run \
	--out-format=colored-line-number 

$(GOBIN)/gofumpt:
	@go install mvdan.cc/gofumpt@latest
	@go mod tidy

gofumpt: | $(GOBIN)/gofumpt
	@gofumpt -w $(shell ls  -d $(PWD)/*/)

$(GOBIN)/gci:
	@go install github.com/daixiang0/gci@latest
	@go mod tidy

gci: | $(GOBIN)/gci
	@gci write --section Standard --section Default --section "Prefix(github.com/wimspaargaren/go-lambda-localstack-example)" $(shell ls  -d $(PWD)/*)

# Debug: @docker build -t localstack-lambda-ci --progress=plain . #debug docker build
ci-init: | zip
	@docker build -t localstack-lambda-ci .

ci-test:
	@docker run localstack-lambda-ci go test ./...

ci-test-integration:
	@docker run --network=host -v "/var/run/docker.sock:/var/run/docker.sock:rw" localstack-lambda-ci go test --tags=integration -v ./...
