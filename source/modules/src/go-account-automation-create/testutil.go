package main

import (
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
)

type mockOrganizationsClient struct {
	organizationsiface.OrganizationsAPI
	createErr   error
	createState string
	createID    string
	orgRootID   string
	destENV     string
	destOUID    string
}

func (m mockOrganizationsClient) CreateAccount(input *organizations.CreateAccountInput) (*organizations.CreateAccountOutput, error) {
	id := &m.createID
	output := organizations.CreateAccountOutput{
		CreateAccountStatus: &organizations.CreateAccountStatus{
			Id:          id,
			AccountName: input.AccountName,
			State:       &m.createState,
		},
	}

	return &output, m.createErr
}

func (m mockOrganizationsClient) DescribeCreateAccountStatus(input *organizations.DescribeCreateAccountStatusInput) (*organizations.DescribeCreateAccountStatusOutput, error) {
	accountID := "999999999999"
	accountName := "test-account"

	output := organizations.DescribeCreateAccountStatusOutput{
		CreateAccountStatus: &organizations.CreateAccountStatus{
			AccountId:   &accountID,
			AccountName: &accountName,
			Id:          input.CreateAccountRequestId,
			State:       &m.createState,
		},
	}
	return &output, m.createErr
}

func (m mockOrganizationsClient) MoveAccount(input *organizations.MoveAccountInput) (*organizations.MoveAccountOutput, error) {
	output := organizations.MoveAccountOutput{}
	return &output, m.createErr
}

func (m mockOrganizationsClient) TagResource(input *organizations.TagResourceInput) (*organizations.TagResourceOutput, error) {
	output := organizations.TagResourceOutput{}
	return &output, m.createErr
}

func (m mockOrganizationsClient) ListRoots(input *organizations.ListRootsInput) (*organizations.ListRootsOutput, error) {
	var roots []*organizations.Root
	root := &organizations.Root{
		Id: &m.orgRootID,
	}
	roots = append(roots, root)
	output := &organizations.ListRootsOutput{Roots: roots}
	return output, m.createErr
}

func (m mockOrganizationsClient) ListOrganizationalUnitsForParent(input *organizations.ListOrganizationalUnitsForParentInput) (*organizations.ListOrganizationalUnitsForParentOutput, error) {
	var returnOUs []*organizations.OrganizationalUnit
	ou := &organizations.OrganizationalUnit{
		Id:   &m.destOUID,
		Name: &m.destENV,
	}
	returnOUs = append(returnOUs, ou)
	output := &organizations.ListOrganizationalUnitsForParentOutput{OrganizationalUnits: returnOUs}
	return output, m.createErr
}
