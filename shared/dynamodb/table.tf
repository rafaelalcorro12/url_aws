terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "local" {
    path = "terraform.tfstate"
  }
}

provider "aws" {
  region                      = "us-east-1"
  access_key                  = "test"
  secret_key                  = "test"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  s3_use_path_style           = true
  endpoints {
    dynamodb = "http://localhost:4566"
    lambda   = "http://localhost:4566"
    apigateway = "http://localhost:4566"
    iam      = "http://localhost:4566"
    s3       = "http://localhost:4566"
    cloudfront = "http://localhost:4566"
  }
}

resource "aws_dynamodb_table" "urls_table" {
  name           = "ShortUrls"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "ShortCode"

  attribute {
    name = "ShortCode"
    type = "S"
  }

  attribute {
    name = "CreatedAt"
    type = "S"
  }

  # GSI para queries por fecha
  global_secondary_index {
    name               = "CreatedAtIndex"
    hash_key           = "CreatedAt"
    projection_type    = "ALL"
  }

  ttl {
    attribute_name = "ExpiresAt"
    enabled        = true
  }

  tags = {
    Name        = "url-shortener-table"
    Environment = "development"
  }
}

output "table_name" {
  value = aws_dynamodb_table.urls_table.name
}
