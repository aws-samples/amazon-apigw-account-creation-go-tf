locals {
  template_vars = {
    account_id                 = var.account_id
    region                     = var.region
    account_provision_put_uri  = module.put_lambda_function.data.invoke_arn
    account_provision_post_uri = module.post_lambda_function.data.invoke_arn
  }
}
