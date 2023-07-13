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
	accountID := "999999999999"
	svc := mockOrganizationsClient{
		accountName: "aws_SEC_test_Dev",
	}
	testPayload := AccountPayload{
		Name:          "aws_SEC_test_Dev",
		AccountPOC:    "01234",
		ApplicationID: "john.doe@example.com",
		Env:           "Dev",
		Lob:           "SEC",
		AccountID:     "",
	}
	error := ValidatePayload(svc, accountID, testPayload)
	if error != nil {
		t.Fatal("Payload failed validation: ", error.Error())
	}

	//test for malformed payload catch
	testPayload.Name = "aws_IS_test_Dev"
	error = ValidatePayload(svc, accountID, testPayload)
	if error == nil {
		t.Fatal("Payload was expected to fail but didn't ")
	}
}

func TestGenerateKeysAndTags(t *testing.T) {
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

	expectedKeys := []*string{&name, &costCenter, &accountPOC, &applicationID, &env, &lob}
	keys, tags := GenerateKeysAndTags(payload)
	for i := range keys {
		if !strings.EqualFold(*keys[i], *expectedKeys[i]) {
			t.Fatal("Account tagging returned unexpected output for keys")

		}
	}

	if !cmp.Equal(tags, expectedTags) {
		t.Fatal("Account tagging returned unexpected output tags")
	}
}

func TestMain(m *testing.M) {
	err := os.Setenv("RUNTIME_ENV", "prod")
	if err != nil {
		log.Panic("Issue setting env var")
	}
	_RUNTIME_ENV_ = os.Getenv("RUNTIME_ENV")

	m.Run()
}
