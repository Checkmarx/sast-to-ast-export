terraform {
  backend "s3" {
    bucket  = "terraform-state-941355383184"
    key     = "terraform/ast-sast-export/terraform.tfstate"
    region  = "eu-west-1"
    encrypt = true
  }
}
