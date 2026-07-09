// Khai báo terraform sẽ làm việc với provider nào để terrform biết 
// mà cài các API call của provider đó. Trong case này là AWS và Cloudflare
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
  }
}

// Khai báo provider compute là AWS
provider "aws" {
  region     = "ap-southeast-1"
  access_key = var.aws_accesskey_id
  secret_key = var.aws_secret_access_key
}

// Khai báo provider network là Cloudflare
provider "cloudflare" {
  api_token = var.cloudflare_api_token
}
