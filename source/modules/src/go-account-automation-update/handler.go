package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
)

// Global
var _RUNTIME_ENV_ = os.Getenv("RUNTIME_ENV")

type AccountPayload struct {
	Name          string `json:"name"`
	CostCenter    string `json:"costCenter"`
	AccountPOC    string `json:"accountPOC"`
	ApplicationID string `json:"applicationId"`
	Env           string `json:"env"`
	Lob           string `json:"lob"`
	AccountID     string `json:"accountId"`
}

func ProcessRequestPayload(request string) (AccountPayload, error) {
	var payload AccountPayload
	error := json.Unmarshal([]byte(request), &payload)
	return payload, error
}

func ValidatePayload(svc organizationsiface.OrganizationsAPI, accountID string, payload AccountPayload) error {
	account, error := svc.DescribeAccount(&organizations.DescribeAccountInput{AccountId: &accountID})
	if error != nil {
		return error
	}
	accountName := account.Account.Name
	if !strings.EqualFold(*accountName, payload.Name) {
		error = errors.New("error: The account name provided seems to differ from the actual account name")
		return error
	}
	s := strings.Split(*accountName, "_")
	_, lob, _, env := s[0], s[1], s[2], s[3] //'aws'_lob_name_env

	if !strings.EqualFold(lob, payload.Lob) || !strings.EqualFold(env, payload.Env) {
		error = errors.New("error: lob or env provided seems to differ from the actual values for account")
		return error
	}
	return nil
}

func GenerateKeysAndTags(payload AccountPayload) ([]*string, []*organizations.Tag) {
	var tags []*organizations.Tag
	var keys []*string
	val := reflect.ValueOf(payload)
	typeOfS := val.Type()

	// Iterate over fields of struct
	for i := 0; i < val.NumField(); i++ {
		key, value := typeOfS.Field(i).Name, fmt.Sprintf("%v", val.Field(i).Interface())
		tag := &organizations.Tag{
			Key:   &key,
			Value: &value,
		}

		if !strings.EqualFold(key, "AccountID") {
			keys = append(keys, &key)
			tags = append(tags, tag)
		}
	}

	return keys, tags
}

func UntagAccount(svc organizationsiface.OrganizationsAPI, keys []*string, accountID string) error {
	if strings.EqualFold(_RUNTIME_ENV_, "prod") {
		input := &organizations.UntagResourceInput{
			ResourceId: &accountID,
			TagKeys:    keys,
		}

		_, error := svc.UntagResource(input)
		return error
	} else {
		log.Println("Dev/Test environment detected - Your request will not trigger UntagResource")
		return nil
	}
}

func TagAccount(svc organizationsiface.OrganizationsAPI, tags []*organizations.Tag, accountID string) error {
	if strings.EqualFold(_RUNTIME_ENV_, "prod") {
		// log.Println("Tags:", tags)
		tagInput := &organizations.TagResourceInput{
			ResourceId: &accountID,
			Tags:       tags,
		}
		_, error := svc.TagResource(tagInput)
		return error
	} else {
		log.Println("Dev/Test environment detected - Your request will not trigger TagResource")
		return nil
	}
}

func HandleErrors(error error, statusCode int) (*events.APIGatewayProxyResponse, error) {
	log.Println("ERROR: ", error.Error())
	response := &events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Body:       error.Error(),
	}

	// return empty error to allow apigw to accurately represent statusCode and error.Error()
	// removing nil will 'make' apigw think the lambda is not exiting gracefully
	return response, nil
}

func GetClient() organizationsiface.OrganizationsAPI {
	sess := session.Must(session.NewSession())
	var svc organizationsiface.OrganizationsAPI
	if strings.EqualFold(_RUNTIME_ENV_, "prod") {
		log.Println("Production environment detected, setting up svc w/ MP creds...")
		creds := stscreds.NewCredentials(sess, os.Getenv("ASSUME_ROLE_ARN"))
		svc = organizations.New(sess, &aws.Config{Credentials: creds})
	} else {
		log.Println("Non-production environment detected, setting up mock svc...")
		svc = mockOrganizationsClient{accountName: "AWS_SEC_test_Dev"}
	}
	return svc
}

func HandleRequest(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println("Setting up session, assume role, and org client")
	svc := GetClient()

	accountID := request.QueryStringParameters["account-id"]

	log.Println("Serializing Payload")
	payload, error := ProcessRequestPayload(request.Body)
	if error != nil {
		return HandleErrors(error, 500)
	}
	log.Println(("Payload serialized without error..."))

	log.Println("Validating payload")
	error = ValidatePayload(svc, accountID, payload)
	if error != nil {
		return HandleErrors(error, 400)
	}
	log.Println(("Payload does not appear malformed..."))

	log.Println("Generating a list of keys and Tag objects from payload...")
	keys, tags := GenerateKeysAndTags(payload)

	log.Println("Untagging Account")
	error = UntagAccount(svc, keys, accountID)
	if error != nil {
		return HandleErrors(error, 500)
	}

	log.Println("Tagging Account")
	error = TagAccount(svc, tags, accountID)
	if error != nil {
		return HandleErrors(error, 500)
	}

	log.Println("Stringifying response body...")
	payload.AccountID = accountID
	jsonResponseBody, error := json.Marshal(payload)
	if error != nil {
		return HandleErrors(error, 500)
	}
	log.Println("Response payload: ", payload)

	response := &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResponseBody),
	}
	return response, nil
}

func main() {
	lambda.Start(HandleRequest)
}
