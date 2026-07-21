terraform {
  backend "s3" {
    bucket       = "chuong-logistic-bucket"
    key          = "logistic/dev/compute/terraform.tfstate"
    region       = "ap-southeast-1"
    use_lockfile = true
  }
}
