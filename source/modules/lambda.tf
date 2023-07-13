####################################################
# ACCOUNT AUTOMATION POST LAMBDA FUNCTION RESOURCE #
####################################################
data "archive_file" "post_lambda_function_archive" {
  type        = "zip"
  source_dir  = "${path.module}/src/lambda/go-account-automation-create/"
  output_path = "${path.module}/src/lambda/go-account-automation-create-archive/go-account-automation-create.zip"
}

resource "aws_lambda_function" "post_lambda_function" {
  filename = "${path.module}/src/lambda/go-account-automation-create-archive/go-account-automation-create.zip"

  lambda_name = "account-automation-post-lambda"
  description = "This lambda acts as an API GW POST method handler and will create accounts based on an API GW request event"

  runtime     = "go1.x"
  handler     = "HandleRequest"
  timeout     = 60
  memory_size = 512
  role        = aws_iam_role.post_lambda_role.arn
  kms_key_arn = aws_kms_key.kms_key.key_arn
  tags        = var.tags

  environment {
    variables = {
      ASSUME_ROLE_ARN = var.create_account_role_arn
      EMAIL_DOMAIN    = var.email_domain
      RUNTIME_ENV     = var.runtime_env
      SEC_OU          = jsonencode(var.infosec_ous)
      WORKLOAD_OU     = var.workload_ou
    }
  }
}

###################################################
# ACCOUNT AUTOMATION PUT LAMBDA FUNCTION RESOURCE #
###################################################
data "archive_file" "put_lambda_function_archive" {
  type        = "zip"
  source_dir  = "${path.module}/src/lambda/go-account-automation-update/"
  output_path = "${path.module}/src/lambda/go-account-automation-update-archive/go-account-automation-update.zip"
}

resource "aws_lambda_function" "put_lambda_function" {
  filename = "${path.module}/src/lambda/go-account-automation-update-archive/go-account-automation-update.zip"

  lambda_name = "account-automation-put-lambda"
  description = "This lambda acts as an API GW PUT method handler and will update an account based on an API GW request event with a request body and query string"

  runtime     = "go1.x"
  handler     = "HandleRequest"
  timeout     = 60
  memory_size = 512
  role        = aws_iam_role.put_lambda_role.arn
  cmk_key_arn = aws_kms_key.kms_key.key_arn
  tags        = var.tags

  environment {
    variables = {
      ASSUME_ROLE_ARN = var.create_account_role_arn
      RUNTIME_ENV     = var.runtime_env
    }
  }
}
