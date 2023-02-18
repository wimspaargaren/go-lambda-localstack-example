FROM golang:1.20.1-alpine3.17

RUN apk update && apk add --no-cache gcc && apk add --no-cache libc-dev

# Install terraform local
RUN apk add --no-cache python3 py3-pip
RUN pip install wheel
RUN pip install terraform-local

# Install terraform
RUN wget https://releases.hashicorp.com/terraform/1.3.9/terraform_1.3.9_linux_amd64.zip
RUN unzip terraform_1.3.9_linux_amd64.zip -d /usr/bin

COPY . /go/src/github.com/wimspaargaren/localstack-lambda-example
WORKDIR /go/src/github.com/wimspaargaren/localstack-lambda-example

RUN tflocal -chdir=infra/terraform init
RUN go mod download
