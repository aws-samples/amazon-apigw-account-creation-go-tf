package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

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

func ValidatePayload(payload AccountPayload) error {
	accountName := &payload.Name

	s := strings.Split(*accountName, "_")
	_, lob, _, env := s[0], s[1], s[2], s[3] //'aws'_lob_name_env

	if !strings.EqualFold(lob, payload.Lob) || !strings.EqualFold(env, payload.Env) {
		error := errors.New("error: lob or env provided seems to differ from those in the account name")
		return error
	}
	return nil
}

func CreateAccount(svc organizationsiface.OrganizationsAPI, accountName string) (string, error) {
	if strings.EqualFold(_RUNTIME_ENV_, "prod") {
		email := accountName + os.Getenv("EMAIL_DOMAIN")
		input := organizations.CreateAccountInput{
			AccountName: &accountName,
			Email:       &email,
		}
		accountOutput, error := svc.CreateAccount(&input)
		if error != nil {
			return "", error
		}

		requestID := accountOutput.CreateAccountStatus.Id
		return *requestID, nil
	} else {
		log.Println("Dev/Test environment detected - Your request will not trigger an account creation")
		return "test", nil
	}
}

func ValidateAccountStatus(svc organizationsiface.OrganizationsAPI, requestID string) (string, error) {
	if strings.EqualFold(_RUNTIME_ENV_, "prod") {
		for {
			status, error := svc.DescribeCreateAccountStatus(&organizations.DescribeCreateAccountStatusInput{CreateAccountRequestId: &requestID})
			if error != nil {
				return "", error
			}
			state := *status.CreateAccountStatus.State
			if state == "FAILED" {
				log.Println("Failed creating Account")
				error = errors.New(*status.CreateAccountStatus.FailureReason)
				return "", error
			} else if state == "IN_PROGRESS" {
				log.Println("In Progress for creating Account")
				time.Sleep(5 * time.Second)
			} else {
				log.Println("Success")
				return *status.CreateAccountStatus.AccountId, nil
			}
		}
	} else {
		log.Println("Dev/Test environment detected - Your request will not trigger ValidateAccountStatus")
		return "test_account_id", nil
	}
}

func RetrieveOUs(svc organizationsiface.OrganizationsAPI, payload AccountPayload) (string, string, error) {
	workloadOU := os.Getenv("WORKLOAD_OU")
	infraSecOUs, error := RetrieveInfraSecOUs()
	if error != nil {
		return "", "", error
	}
	// log.Println("infraSecOUs, workload:", infraSecOUs, ",", workloadOU)

	parentOU := RetrieveParentOU(infraSecOUs, workloadOU, payload.Lob)
	// log.Println("parentOU:", parentOU)

	envOU, root, error := RetrieveEnvOU(svc, infraSecOUs, parentOU, payload.Env)
	if error != nil {
		return "", "", error
	}
	// log.Println("envOU,root:", envOU, ",", root)

	ou, error := DetermineDestinationOU(svc, infraSecOUs, envOU, payload.Lob)
	if error != nil {
		return "", "", error
	}
	if ou == "" && strings.EqualFold(_RUNTIME_ENV_, "prod") {
		error = errors.New("error: Destination OU not found")
		return "", "", error
	}
	// log.Println("ou", ou)

	return root, ou, nil
}

func RetrieveInfraSecOUs() (map[string]string, error) {
	var infraSecOUs map[string]string
	// Serialize json to map
	jsonMap := os.Getenv("SEC_OU")
	error := json.Unmarshal([]byte(jsonMap), &infraSecOUs)
	if error != nil {
		return nil, error
	}
	return infraSecOUs, nil
}

func RetrieveParentOU(infraSecOUs map[string]string, workloadOU string, lob string) string {
	if infraSecOU, ok := infraSecOUs[lob]; ok {
		return infraSecOU
	} else {
		return workloadOU
	}
}

func RetrieveEnvOU(svc organizationsiface.OrganizationsAPI, infraSecOUs map[string]string, parentOU string, env string) (string, string, error) {
	rootList, error := svc.ListRoots(&organizations.ListRootsInput{})
	if error != nil {
		return "", "", error
	}
	root := rootList.Roots[0].Id

	envOUs, error := svc.ListOrganizationalUnitsForParent(&organizations.ListOrganizationalUnitsForParentInput{ParentId: &parentOU})
	if error != nil {
		return "", "", error
	}

	var envOU string
	for _, tempOU := range envOUs.OrganizationalUnits {
		if strings.EqualFold(*tempOU.Name, env) {
			envOU = *tempOU.Id
			break
		}
	}

	return envOU, *root, nil
}

func DetermineDestinationOU(svc organizationsiface.OrganizationsAPI, infraSecOUs map[string]string, envOU string, lob string) (string, error) {
	if _, ok := infraSecOUs[lob]; ok {
		return envOU, nil
	}

	lobOUs, error := svc.ListOrganizationalUnitsForParent(&organizations.ListOrganizationalUnitsForParentInput{ParentId: &envOU})
	if error != nil {
		return "", error
	}

	var ou string
	for _, tempOU := range lobOUs.OrganizationalUnits {
		if *tempOU.Name == lob {
			ou = *tempOU.Id
			break
		}
	}

	return ou, nil
}

func MoveAccount(svc organizationsiface.OrganizationsAPI, accountID string, root string, ou string) error {
	if strings.EqualFold(_RUNTIME_ENV_, "prod") {
		input := organizations.MoveAccountInput{
			AccountId:           &accountID,
			SourceParentId:      &root,
			DestinationParentId: &ou,
		}
		_, error := svc.MoveAccount(&input)
		return error
	} else {
		log.Println("Dev/Test environment detected - Your request will not trigger MoveAccount")
		return nil
	}
}

func GenerateTags(payload AccountPayload) []*organizations.Tag {
	var tags []*organizations.Tag
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
			tags = append(tags, tag)
		}
	}

	return tags
}

func TagAccount(svc organizationsiface.OrganizationsAPI, tags []*organizations.Tag, accountID string) error {
	  else {
		log.Println("Dev/Test environment detected - Your request will not trigger TagAccount")
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
		svc = mockOrganizationsClient{}
	}
	return svc
}

func HandleRequest(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	log.Println("Setting up session, assume role, and org client...")
	svc := GetClient()

	log.Println("Serializing Payload...")
	payload, error := ProcessRequestPayload(request.Body)
	if error != nil {
		return HandleErrors(error, 500)
	}
	log.Println("Payload serialized without error...")

	log.Println("Validating payload...")
	error = ValidatePayload(payload)
	if error != nil {
		return HandleErrors(error, 400)
	}
	log.Println("Payload does not appear malformed...")

	log.Println("Creating Account...")
	requestID, error := CreateAccount(svc, payload.Name)
	if error != nil {
		return HandleErrors(error, 500)
	}

	log.Println("Validating account creation status...")
	accountID, error := ValidateAccountStatus(svc, requestID)
	if error != nil {
		return HandleErrors(error, 500)
	}

	log.Println("Retrieving Root ID and correct OU ID based on payload...")
	root, ou, error := RetrieveOUs(svc, payload)
	if error != nil {
		return HandleErrors(error, 500)
	}

	log.Println("Moving account to correct OU...")
	error = MoveAccount(svc, accountID, root, ou)
	if error != nil {
		return HandleErrors(error, 500)
	}

	log.Println("Generating a list of Tag objects from payload...")
	tags := GenerateTags(payload)
	log.Println("Tags: ", tags)

	log.Println("Tagging Account...")
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
