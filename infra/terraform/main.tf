provider "aws" {

  access_key = "mock_access_key"
  secret_key = "mock_secret_key"
  region     = "us-east-1"

  s3_use_path_style           = true
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
}

terraform {
  backend "local" {}
}

resource "aws_s3_bucket" "test-bucket" {
  bucket = "my-bucket"
}

resource "aws_lambda_function" "test_lambda" {
  # If the file is not in the current working directory you will need to include a
  # path.module in the filename.
  filename      = "app.zip"
  function_name = "test_lambda"
  role          = "arn:aws:iam::123456:role/irrelevant"
  handler       = "app"
  runtime       = "go1.x"

  environment {
    variables = {
      foo = "bar"
    }
  }
}

resource "aws_api_gateway_rest_api" "api_gw" {
  name        = "api"
  description = "API Gateway V1"
}

output "api_gw_id" {
  value = aws_api_gateway_rest_api.api_gw.id
}

resource "aws_api_gateway_resource" "proxy" {
  rest_api_id = aws_api_gateway_rest_api.api_gw.id
  parent_id   = aws_api_gateway_rest_api.api_gw.root_resource_id
  path_part   = "{somethingId}"
}

resource "aws_api_gateway_method" "proxy" {
  rest_api_id   = aws_api_gateway_rest_api.api_gw.id
  resource_id   = aws_api_gateway_resource.proxy.id
  http_method   = "ANY"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "lambda" {
  rest_api_id             = aws_api_gateway_rest_api.api_gw.id
  resource_id             = aws_api_gateway_method.proxy.resource_id
  http_method             = aws_api_gateway_method.proxy.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.test_lambda.invoke_arn
}

resource "aws_api_gateway_deployment" "apigw_deployment" {
  depends_on = [
    aws_api_gateway_integration.lambda,
  ]
  rest_api_id = aws_api_gateway_rest_api.api_gw.id
  stage_name  = "test"
}
