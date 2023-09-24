package loader_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
	"github.com/mxmauro/configreader/loader"
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

	// Check if vault is running
	vaultAddr := testhelpers.EnsureVaultAvailability(t)

	// Create vault accessor with root privileges
	client := testhelpers.CreateRootVaultClient(t, vaultAddr)

	// Add AWS auth engine
	testhelpers.EnableAwsAuthEngine(t, client)

	// Assume vault is running in a EC2 instance with an IAM role with the following polices attached:
	// * ec2:DescribeInstances
	// * iam:GetInstanceProfile (if IAM Role binding is used)

	// Create the role
	testhelpers.CreateAwsRoleForReadSecretPolicy(t, client)

	// Write the secrets
	testhelpers.WriteVaultSecret(t, client, "settings", testhelpers.GoodSettingsJSON)

	// Load configuration from vault using AWS
	settings, err := configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewVault().
			WithHost(vaultAddr).
			WithPath(testhelpers.PathFromSecretKey("settings")).
			WithAuth(loader.NewVaultAwsAuthMethod().
				WithRole(testhelpers.VaultAwsRoleName).
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
