###########################################################
# ACCOUNT AUTOMATION POST LAMBDA FUNCTION ROLE AND POLICY #
###########################################################
resource "aws_iam_role" "post_lambda_role" {
  name               = "account-automation-post-lambda-role"
  description        = "IAM Role for account-automation-post-lambda-role"
  path               = "/delegated/${var.application_id}/"
  tags               = var.tags
  assume_role_policy = data.aws_iam_policy_document.post_lambda_trust_policy.json
}

resource "aws_iam_role_policy_attachment" "post_lambda_trust_policy" {
  role       = aws_iam_role.post_lambda_role.name
  policy_arn = aws_iam_policy.post_lambda_policy.arn
}

resource "aws_iam_policy" "post_lambda_policy" {
  name        = "account-automation-post-lambda-policy"
  path        = "/delegated/${var.application_id}/"
  description = "IAM policy for account-automation-post-lambda"
  policy      = data.aws_iam_policy_document.post_lambda_permissions.json
}

data "aws_iam_policy_document" "post_lambda_trust_policy" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    principals {
      type = "Service"
      identifiers = [
        "lambda.amazonaws.com",
      ]
    }
  }
}

data "aws_iam_policy_document" "post_lambda_permissions" {
  statement {
    effect    = "Allow"
    actions   = "sts:AssumeRole"
    resources = [var.create_account_role_arn]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "kms:CreateGrant",
      "kms:Decrypt",
      "kms:Encrypt",
      "kms:ListAliases",
    ]
    resources = [
      module.kms_key.key_arn,
    ]
  }
}

##########################################################
# ACCOUNT AUTOMATION PUT LAMBDA FUNCTION ROLE AND POLICY #
##########################################################
resource "aws_iam_role" "put_lambda_role" {
  name               = "account-automation-put-lambda-role"
  description        = "IAM Role for account-automation-put-lambda"
  path               = "/delegated/${var.application_id}/"
  tags               = var.tags
  assume_role_policy = data.aws_iam_policy_document.put_lambda_trust_policy.json
}

resource "aws_iam_role_policy_attachment" "put_lambda_trust_policy" {
  role       = aws_iam_role.put_lambda_role.name
  policy_arn = aws_iam_policy.put_lambda_policy.arn
}

resource "aws_iam_policy" "put_lambda_policy" {
  name        = "account-automation-put-lambda-policy"
  path        = "/delegated/${var.application_id}/"
  description = "IAM policy for account-automation-put-lambda"
  policy      = data.aws_iam_policy_document.put_lambda_permissions.json
}

data "aws_iam_policy_document" "put_lambda_trust_policy" {
  statement {
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    principals {
      type = "Service"
      identifiers = [
        "lambda.amazonaws.com",
      ]
    }
  }
}

data "aws_iam_policy_document" "put_lambda_permissions" {
  statement {
    effect    = "Allow"
    actions   = "sts:AssumeRole"
    resources = [var.create_account_role_arn]
  }

  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      "*",
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "kms:CreateGrant",
      "kms:Decrypt",
      "kms:Encrypt",
      "kms:ListAliases",
    ]
    resources = [
      module.kms_key.key_arn,
    ]
  }
}

####################################
# API GATEWAY RESOURCE PERMISSIONS #
####################################
data "aws_iam_policy_document" "api_gateway_pd" {
  statement {
    actions   = ["execute-api:Invoke"]
    resources = ["*"]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_lambda_permission" "apigw_post_lambda_permission" {
  statement_id  = "AllowInvokeAccountAutomationPostLambda"
  action        = "lambda:InvokeFunction"
  function_name = "account-automation-post-lambda"
  principal     = "apigateway.amazonaws.com"

  # The /*/*/* part allows invocation from any stage, method and resource path
  # within API Gateway REST API.
  source_arn = "${module.apigw.execution_arn}/*/*/*"
}

resource "aws_lambda_permission" "apigw_put_lambda_permission" {
  statement_id  = "AllowInvokeAccountAutomationPutLambda"
  action        = "lambda:InvokeFunction"
  function_name = "account-automation-put-lambda"
  principal     = "apigateway.amazonaws.com"

  # The /*/*/* part allows invocation from any stage, method and resource path
  # within API Gateway REST API.
  source_arn = "${module.apigw.execution_arn}/*/*/*"
}
