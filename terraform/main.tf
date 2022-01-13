provider "aws" {
  region  = var.region
}

resource "aws_kms_key" "migration_key" {
  customer_master_key_spec = "RSA_4096"
}

resource "aws_kms_alias" "migration_key" {
  name          = "alias/sast-migration-key"
  target_key_id = aws_kms_key.migration_key.key_id
}
