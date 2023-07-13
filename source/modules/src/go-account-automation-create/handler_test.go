package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/google/go-cmp/cmp"
)

func TestProcessRequestPayload(t *testing.T) {
	// pwd, _ := os.Getwd()
	jsonFile, error := os.Open("testdata/request.json")
	if error != nil {
		log.Panic("Error setting up TestProcessRequestPayload - error opening request.json: ", error.Error())
	}

	byteValue, _ := ioutil.ReadAll(jsonFile)
	if error != nil {
		log.Panic("Error setting up TestProcessRequestPayload - error reading request.json: ", error.Error())
	}

	//test
	payload, error := ProcessRequestPayload(string(byteValue))
	if error != nil {
		t.Fatal("Error serializing payload: ", error.Error())
	}

	if !strings.EqualFold(payload.Name, "aws_SEC_test_Dev") ||
		!strings.EqualFold(payload.CostCenter, "01234") ||
		!strings.EqualFold(payload.AccountPOC, "john.doe@example.com") ||
		!strings.EqualFold(payload.ApplicationID, "00000000-0000-0000-0000-000000000000") ||
		!strings.EqualFold(payload.Env, "Dev") ||
		!strings.EqualFold(payload.Lob, "SEC") ||
		!strings.EqualFold(payload.AccountID, "") {
		t.Fatal("Payload object fields not as expected")
	}
}

func TestValidatePayload(t *testing.T) {
	testPayload := AccountPayload{
		Name:          "aws_SEC_test_Dev",
		AccountPOC:    "01234",
		ApplicationID: "john.doe@example.com",
		Env:           "Dev",
		Lob:           "SEC",
		AccountID:     "",
	}
	error := ValidatePayload(testPayload)
	if error != nil {
		t.Fatal("Payload failed validation: ", error.Error())
	}

	//test for malformed payload catch
	testPayload.Name = "aws_IS_test_Dev"
	error = ValidatePayload(testPayload)
	if error == nil {
		t.Fatal("Payload was expected to fail but didn't ")

	}
}

func TestRetrieveOUs(t *testing.T) {
	svc := mockOrganizationsClient{
		destENV:   "Dev",
		destOUID:  "ou-abcd-12345678",
		orgRootID: "r-abcd",
		createErr: nil,
	}
	payload := AccountPayload{
		Name:          "aws_SEC_test_Dev",
		CostCenter:    "01234",
		AccountPOC:    "john.doe@example.com",
		ApplicationID: "00000000-0000-0000-0000-000000000000",
		Env:           "DEV",
		Lob:           "SEC",
	}
	root, ou, error := RetrieveOUs(svc, payload)
	if error != nil {
		t.Fatal("RetrieveOUs failed")
	}

	if !strings.EqualFold(root, svc.orgRootID) || !strings.EqualFold(ou, svc.destOUID) {
		t.Fatal("RetrieveOUs output not as expected")
	}
}

func TestCreateAccount(t *testing.T) {
	accountName := "aws_SEC_test_Dev"
	svc := mockOrganizationsClient{
		createID:  "car-012345678912",
		createErr: nil,
	}

	requestID, error := CreateAccount(svc, accountName)
	if error != nil {
		t.Fatal("Account Creation Failed:", error.Error())
	}
	if requestID != "car-012345678912" {
		log.Println("requestID:", requestID)
		t.Fatal("Account Creation returned unexpected output")
	}
}

func TestValidateAccountStatus(t *testing.T) {
	svc := mockOrganizationsClient{
		createState: organizations.CreateAccountStateSucceeded,
		createErr:   nil,
		createID:    "car-012345678912",
	}
	accountID, error := ValidateAccountStatus(svc, svc.createID)
	if error != nil {
		t.Fatal(error.Error())
	}
	if accountID != "999999999999" {
		t.Fatal("Account ID was not as expected")
	}
}

func TestGenerateTags(t *testing.T) {
	payload := AccountPayload{
		Name:          "aws_SEC_test_Dev",
		CostCenter:    "01234",
		AccountPOC:    "john.doe@example.com",
		ApplicationID: "00000000-0000-0000-0000-000000000000",
		Env:           "DEV",
		Lob:           "SEC",
	}
	name, costCenter, accountPOC, applicationID, env, lob := "Name", "CostCenter", "AccountPOC", "ApplicationID", "Env", "Lob"
	expectedTags := []*organizations.Tag{
		{
			Key:   &name,
			Value: &payload.Name,
		},
		{
			Key:   &costCenter,
			Value: &payload.CostCenter,
		},
		{
			Key:   &accountPOC,
			Value: &payload.AccountPOC,
		},
		{
			Key:   &applicationID,
			Value: &payload.ApplicationID,
		},
		{
			Key:   &env,
			Value: &payload.Env,
		},
		{
			Key:   &lob,
			Value: &payload.Lob,
		},
	}
	tags := GenerateTags(payload)

	if !cmp.Equal(tags, expectedTags) {
		t.Fatal("Account tagging returned unexpected output")
	}
}

func TestMain(m *testing.M) {
	var err error

	err = os.Setenv("RUNTIME_ENV", "prod")
	if err != nil {
		log.Panic("Issue setting env var")
	}
	_RUNTIME_ENV_ = os.Getenv("RUNTIME_ENV")

	err = os.Setenv("EMAIL_DOMAIN", "@example.com")
	if err != nil {
		log.Panic("Issue setting env var")
	}

	err = os.Setenv("WORKLOAD_OU", "ou-abcd-01234567")
	if err != nil {
		log.Panic("Issue setting env var")
	}

	err = os.Setenv("SEC_OU", "{\"SEC\":\"ou-abcd-12345678\",\"IS\":\"ou-abcd-23456789\"}")
	if err != nil {
		log.Panic("Issue setting env var")
	}

	m.Run()
}
