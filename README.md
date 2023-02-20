# Go Lambda

This repository demonstrates an example on how to run a Go lambda on localstack, where localstack is provisioned using tflocal (a localstack specific wrapper around Terraform). In addition, the repository contains an example on levarage localstack in order to integration test your Go Lambda.

# On my virtual machine

## Pre-requisites

- Docker

## Run integration tests

Run `make ci-init && make ci-test-integration`

# On my machine

## Pre-requisites

- Docker
- Go
- [terraform](https://www.terraform.io/) & [terraform local](https://docs.localstack.cloud/user-guide/integrations/terraform/)
- [aws cli](https://aws.amazon.com/cli/) & [awscli-local](https://github.com/localstack/awscli-local) in case you want to check what has been "deployed" in localstack

## Start the Lambda
- Run `make localstack`
- Run `make terraform`
- Use the output `api_gw_id` to compose the url
- Curl the hello world endpoint: `curl --location --request GET "http://localhost:4566/restapis/$(tflocal -chdir=infra/terraform output -raw api_gw_id)/test/_user_request_/hello-world"`

## Run integration tests

- Run `make test-integration`
