terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  endpoints {
    dynamodb   = "http://localhost:4566"
    lambda     = "http://localhost:4566"
    apigateway = "http://localhost:4566"
    iam        = "http://localhost:4566"
  }
}

# IAM Role para Lambda
resource "aws_iam_role" "lambda_role" {
  name = "module2-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "dynamodb_access" {
  name = "dynamodb-access"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = ["dynamodb:GetItem", "dynamodb:UpdateItem"]
      Resource = "arn:aws:dynamodb:us-east-1:000000000000:table/ShortUrls"
    }]
  })
}

# Lambda Function
resource "aws_lambda_function" "redirect_function" {
  filename         = "../function.zip"
  function_name    = "url-shortener-redirect"
  role            = aws_iam_role.lambda_role.arn
  handler         = "bootstrap"
  runtime         = "provided.al2"
  timeout         = 10
  memory_size     = 128
}

# API Gateway
resource "aws_api_gateway_rest_api" "redirect_api" {
  name        = "Redirect API"
  description = "API for URL redirection"
}

resource "aws_api_gateway_resource" "code_resource" {
  rest_api_id = aws_api_gateway_rest_api.redirect_api.id
  parent_id   = aws_api_gateway_rest_api.redirect_api.root_resource_id
  path_part   = "{code}"
}

resource "aws_api_gateway_method" "get_method" {
  rest_api_id   = aws_api_gateway_rest_api.redirect_api.id
  resource_id   = aws_api_gateway_resource.code_resource.id
  http_method   = "GET"
  authorization = "NONE"

  request_parameters = {
    "method.request.path.code" = true
  }
}

resource "aws_api_gateway_integration" "lambda_integration" {
  rest_api_id             = aws_api_gateway_rest_api.redirect_api.id
  resource_id             = aws_api_gateway_resource.code_resource.id
  http_method             = aws_api_gateway_method.get_method.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = aws_lambda_function.redirect_function.invoke_arn
}

resource "aws_api_gateway_deployment" "deployment" {
  depends_on = [aws_api_gateway_integration.lambda_integration]
  rest_api_id = aws_api_gateway_rest_api.redirect_api.id
  stage_name  = "dev"
}

resource "aws_lambda_permission" "api_gateway" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.redirect_function.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.redirect_api.execution_arn}/*/*"
}

output "api_url" {
  value = "${aws_api_gateway_deployment.deployment.invoke_url}/{code}"
}
