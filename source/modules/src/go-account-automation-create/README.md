# go-aws-app-account-automation-create

HandleRequest, acting as an account creator, receives a request payload of type *events.APIGatewayProxyRequest* from an API GW endpoint. Using this request, the script will first validate the request, then it will create an account, validate the account creation status, then if successful, tag the account, and finally move the account to the correct OU based on this payload. This process is "mocked" in non-production environments and logs that it was a test run and no APIs have been invoked.

Before deploying using Terraform, the Golang code must be compiled, built, and zipped into the file specified in the Terraform aws_lambda_function resource in lambda.tf.

## Example Input & Output
#### Input
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

## Finding the correct OU
A diagram of the current (at the time of this readme) OU structure can be found on the 

There a two variables passed into the lambda to help with this: a map of Security OU IDs and the Workload OU ID.

From there, using these OU IDs, multiple calls are made to [ListOrganizationalUnitsForParent](https://docs.aws.amazon.com/sdk-for-go/api/service/organizations/#Organizations.ListOrganizationalUnitsForParent) based on the LOB and ENV from the client request to eventually return the correct OU ID to move the account to.

## Resource Deployment 
This resource, among others, is deployed via Terraform.

## Unit Testing
handler_test.go handles test invocation and setting up the test environment inside the TestMain() function.

testutil.go is a list of mock methods and is used for both testing and mocking the sdk in non-production environments.
