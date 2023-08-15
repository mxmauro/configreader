package loader_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
	"github.com/mxmauro/configreader/loader"
)

//------------------------------------------------------------------------------

const (
	awsRoleName = "test-aws"
)

//------------------------------------------------------------------------------

func TestVaultLoaderWithAwsAuth(t *testing.T) {
	testhelpers.EnsureAwsEC2Instance(t)
	/*
		t.Logf("please execute this command in order to run Vault tests:")
		t.Logf(`    vault server -dev -dev-root-token-id="root" --dev-no-store-token`)

		* ec2:DescribeInstances
		* iam:GetInstanceProfile (if IAM Role binding is used)
	*/

	// Get vault address
	vaultAddr := os.Getenv("VAULT_ADDR")
	if len(vaultAddr) == 0 {
		vaultAddr = "127.0.0.1:8200"
		t.Logf("VAULT_ADDR environment variable not set, assuming 127.0.0.1:8200")
	}

	// Check if vault is running
	testhelpers.EnsureVaultAvailability(t, vaultAddr)

	// Create vault accessor with root privileges
	client := testhelpers.CreateRootVaultClient(t, vaultAddr)

	// Add AWS auth engine
	enableAwsAuthEngine(t, client)

	// Assume vault is running in a EC2 instance with an IAM role with the following polices attached:
	// * ec2:DescribeInstances
	// * iam:GetInstanceProfile (if IAM Role binding is used)

	// Create the role
	createAwsRoleForReadSecretPolicy(t, client)

	// Write the secrets
	testhelpers.WriteVaultSecret(t, client, "settings", testhelpers.GoodSettingsJSON)

	// Load configuration from vault using AWS
	settings, err := configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewVault().
			WithHost(vaultAddr).
			WithPath(testhelpers.PathFromSecretKey("settings")).
			WithAuth(loader.NewVaultAwsAuthMethod().
				WithRole(awsRoleName).
				WithTypeEC2().
				WithPKCS7Signature().
				WithRegion("us-east-1").
				WithNonce(""))).
		WithSchema(testhelpers.SchemaJSON).
		Load(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, &testhelpers.GoodSettings) {
		t.Fatalf("settings mismatch")
	}
}

func enableAwsAuthEngine(t *testing.T, client *api.Client) {
	// Enable AppRole auth method
	err := client.Sys().EnableAuthWithOptions("aws/", &api.EnableAuthOptions{
		Type:        "aws",
		Description: "AWS auth method",
	})
	if err != nil {
		var apiErr *api.ResponseError
		if !(errors.As(err, &apiErr) &&
			apiErr.StatusCode == 400 &&
			len(apiErr.Errors) > 0 &&
			strings.Contains(apiErr.Errors[0], "already in use")) {
			t.Fatalf("unable to enable AWS auth [err=%v]", err)
		}
	}
}

func createAwsRoleForReadSecretPolicy(t *testing.T, client *api.Client) {
	vpcId := testhelpers.GetEC2InstanceVpcId(t)
	// Create an AppRole role with the created policy
	_, err := client.Logical().Write("auth/aws/role/"+awsRoleName, map[string]interface{}{
		"auth_type":    "ec2",
		"policies":     []string{testhelpers.VaultReadSecretPolicy},
		"bound_vpc_id": []string{vpcId},
	})
	if err != nil {
		t.Fatalf("unable to create the test aws role [err=%v]", err)
	}
}
