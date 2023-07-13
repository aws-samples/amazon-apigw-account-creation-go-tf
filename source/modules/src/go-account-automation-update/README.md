# go-aws-app-account-automation-update

Acting as an updater for the Account Automation, this Lambda receives a request payload of type events.APIGatewayProxyRequest from an API GW endpoint. Using this request, the script will first validate the request, then if successful, create tags based on the payload, untag the account, and finally retag the account with the updated tags provided in the payload. This process is "mocked" in non-production environments and logs that it was a test run and no APIs have been invoked.

Before deploying using Terraform, the Golang code must be compiled, built, and zipped into the file specified in the Terraform aws_lambda_function resource in lambda.tf.

## Example Input & Output
#### Input
##### Payload Query String

| Field           | Value        |
| --------------- | ------------ |
| account-id      | 123456789012 |

##### Payload Body
```javascript
{
  "name": "AWS_SEC_Example_Dev",
  "costCenter": "01234",
  "accountPOC": "john.doe@example.com",
  "applicationId": "00000000-0000-0000-0000-000000000000",
  "env": "DEV",
  "lob": "SEC"
}
```

#### Output
```javascript
{
  "name": "AWS_SEC_Example_Dev",
  "accountId": "123456789012",
  "costCenter": "01234",
  "accountPOC": "john.doe@example.com",
  "applicationId": "00000000-0000-0000-0000-000000000000",
  "env": "DEV",
  "lob": "SEC"
}
```

Notice the appended accountId.

## Validation
Most simple validation (such as the accountPOC ending in @example.com) is handled on the API GW.

The script will manually validate that the LOB and ENV string slices on the name field match the env and lob fields.

Further, the following fields are immutable and will throw an exception if the fields in the request payload do not match those that are already tagged to the account:

* name
* env
* lob

## Resource Deployment 
This resource, among others, is deployed via terraform.

## Unit Testing
handler_test.go handles test invocation and setting up the test environment inside the TestMain() function.

testutil.go is a list of mock methods and is used for both testing and mocking the sdk in non-production enviornments.
