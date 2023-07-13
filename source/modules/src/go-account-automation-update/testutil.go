package main

import (
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/aws/aws-sdk-go/service/organizations/organizationsiface"
)

type mockOrganizationsClient struct {
	organizationsiface.OrganizationsAPI
	createErr   error
	accountName string
}

func (m mockOrganizationsClient) DescribeAccount(*organizations.DescribeAccountInput) (*organizations.DescribeAccountOutput, error) {
	account := &organizations.Account{
		Name: &m.accountName,
	}
	output := &organizations.DescribeAccountOutput{
		Account: account,
	}
	return output, m.createErr
}

func (m mockOrganizationsClient) TagResource(input *organizations.TagResourceInput) (*organizations.TagResourceOutput, error) {
	output := &organizations.TagResourceOutput{}
	return output, m.createErr
}

func (m mockOrganizationsClient) UntagResource(input *organizations.UntagResourceInput) (*organizations.UntagResourceOutput, error) {
	output := &organizations.UntagResourceOutput{}
	return output, m.createErr
}
