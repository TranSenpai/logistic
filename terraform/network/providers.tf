// Khai báo terraform sẽ làm việc với provider nào để terrform biết 
// mà cài các API call của provider đó. Trong case này là AWS và Cloudflare
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.54.0"
    }
  }
}

// Khai báo provider compute là AWS
provider "aws" {
  region     = var.aws_region
  access_key = var.aws_access_key_id
  secret_key = var.aws_secret_access_key
}
