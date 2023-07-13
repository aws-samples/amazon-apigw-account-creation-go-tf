variable "tags" {
  type = map(string)
}

variable "account_id" {
  type        = string
  description = "AWS account ID these resources are being deployed into."
}

variable "application_id" {
  type        = string
  description = "Application ID to identify resources deployed by this module."

}

variable "create_account_role_arn" {
  type        = string
  description = "The role arn in MP (or whichever account that has permission to create other organization accounts) that will be assumed and used to create accounts."
}

variable "email_domain" {
  type        = string
  description = "Email domain used in validation of the account POC email address."
}

variable "region" {
  type        = string
  description = "Region resources are being deployed in."
}

variable "runtime_env" {
  type        = string
  description = "LAB|DEV|TEST|PROD environment this is being deployed to. When elevating to PROD, ensure PROD is passed."
}

variable "infosec_ous" {
  type        = map(string)
  description = "Map of security related OU IDs."
}

variable "workload_ou" {
  type        = string
  description = "Workload (or Application OU) that requesters of new accounts will be having their accounts deployed in."
}
