terraform {
  required_version = ">= 0.12"
}

provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
}

data "aws_caller_identity" "main" {
}

locals {
  account_id              = data.aws_caller_identity.main.account_id
  application_id          = "00000000-0000-0000-0000-000000000000"
  create_account_role_arn = ""
  email_domain            = "@example.com"
  runtime_env             = "dev"

  infosec_ous = {
    "SOAR" = "ou-abcd-01234567"
    "IS"   = "ou-abcd-01234567"
  }
  workload_ou = "ou-abcd-01234567"
}

module "us-east-1" {
  source = "./modules"

  account_id              = local.account_id
  application_id          = local.application_id
  create_account_role_arn = local.create_account_role_arn
  email_domain            = local.email_domain
  region                  = "us-east-1"
  runtime_env             = local.runtime_env

  infosec_ous = local.infosec_ous
  workload_ou = local.workload_ou

  tags = var.tags

  providers = {
    aws = aws.us-east-1
  }
}
