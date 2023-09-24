package testhelpers

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

//------------------------------------------------------------------------------

const (
	VaultAwsRoleName = "test-aws"
)

//------------------------------------------------------------------------------

func EnableAwsAuthEngine(t *testing.T, client *api.Client) {
	// Enable AppRole auth method
	err := client.Sys().EnableAuthWithOptions("aws/", &api.EnableAuthOptions{
		Type:        "aws",
		Description: "AWS auth method",
	})
	if err != nil {
		var apiErr *api.ResponseError

		if !(errors.As(err, &apiErr) && apiErr.StatusCode == 400 && len(apiErr.Errors) > 0 &&
			strings.Contains(apiErr.Errors[0], "already in use")) {
			t.Fatalf("unable to enable AWS auth [err=%v]", err)
		}
	}
}

func CreateAwsRoleForReadSecretPolicy(t *testing.T, client *api.Client) {
	vpcId := GetEC2InstanceVpcId(t)
	// Create an AppRole role with the created policy
	_, err := client.Logical().Write("auth/aws/role/"+VaultAwsRoleName, map[string]interface{}{
		"auth_type":    "ec2",
		"policies":     []string{VaultReadSecretPolicy},
		"bound_vpc_id": []string{vpcId},
		"token_ttl":    30, // Tokens must be renewed within 30 seconds
	})
	if err != nil {
		t.Fatalf("unable to create the test aws role [err=%v]", err)
	}
}
