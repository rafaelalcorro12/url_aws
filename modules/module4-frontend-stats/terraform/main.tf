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
    s3 = "http://localhost:4566"
  }

  s3_use_path_style = true
}

# Crear bucket S3 para estadísticas
resource "aws_s3_bucket" "stats_frontend" {
  bucket = "stats-frontend-url-shortener"
  force_destroy = true
}

# Configurar website hosting
resource "aws_s3_bucket_website_configuration" "website" {
  bucket = aws_s3_bucket.stats_frontend.id
  index_document {
    suffix = "index.html"
  }
}

# Subir index.html
resource "aws_s3_object" "index_html" {
  bucket       = aws_s3_bucket.stats_frontend.id
  key          = "index.html"
  source       = "../index.html"
  content_type = "text/html"
}

# Hacer el bucket público
resource "aws_s3_bucket_public_access_block" "public_access" {
  bucket = aws_s3_bucket.stats_frontend.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

# Política pública
resource "aws_s3_bucket_policy" "public_policy" {
  bucket = aws_s3_bucket.stats_frontend.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = "*"
        Action    = "s3:GetObject"
        Resource  = "${aws_s3_bucket.stats_frontend.arn}/*"
      }
    ]
  })

  depends_on = [aws_s3_bucket_public_access_block.public_access]
}

output "website_url" {
  value = "http://localhost:4566/stats-frontend-url-shortener/index.html"
}

output "bucket_name" {
  value = aws_s3_bucket.stats_frontend.id
}
