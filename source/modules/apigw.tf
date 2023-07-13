resource "aws_api_gateway_rest_api" "api" {
  name        = "AccountAutomationAPI"
  description = "API GW for Account Creation workflows. This API GW has a POST and PUT method with lambda proxy integration acting as an account create and account update, respectively."
  policy      = data.aws_iam_policy_document.api_gateway_pd.json
  body        = templatefile("${path.module}/src/json/swagger.json", local.template_vars)
  tags        = var.tags

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_deployment" "api_deployment" {
  rest_api_id       = aws_api_gateway_rest_api.api.id
  stage_description = "Initial stage deployment."

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_api_gateway_stage" "stage" {
  stage_name           = "v1"
  rest_api_id          = aws_api_gateway_rest_api.api.id
  deployment_id        = aws_api_gateway_deployment.api_deployment.id
  xray_tracing_enabled = true
  tags                 = var.tags
}

resource "aws_api_gateway_usage_plan" "usage_plan" {
  name = "usage_plan"

  api_stages {
    api_id = module.apigw.rest_api_id
    stage  = "v1"
  }

  depends_on = [module.apigw.api_stage_id]
}

resource "aws_api_gateway_method_settings" "api_method_settings" {
  rest_api_id = module.apigw.rest_api_id
  stage_name  = "v1"
  method_path = "*/*"

  settings {
    metrics_enabled = true
    logging_level   = "ERROR"
  }

  depends_on = [module.apigw.api_stage_id]
}
