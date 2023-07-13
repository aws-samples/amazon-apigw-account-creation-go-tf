// Application KMS Key
resource "aws_kms_key" "kms_key" {
  description         = "The key used to encrypt and decrypt data for Account Automation API related resources"
  enable_key_rotation = true
  tags                = var.tags

  integrated_aws_services = ["lambda", "ssm"]
}

resource "aws_kms_alias" "kms_key_alias" {
  name          = "alias/account-automation-${var.region}"
  target_key_id = aws_kms_key.kms_key.key_id
}
