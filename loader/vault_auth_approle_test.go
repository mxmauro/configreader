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

func TestVaultLoaderWithAppRoleAuth(t *testing.T) {
	// Check if vault is running
	vaultAddr := testhelpers.EnsureVaultAvailability(t)

	// Create vault accessor with root privileges
	client := testhelpers.CreateRootVaultClient(t, vaultAddr)

	// Write the secrets
	testhelpers.WriteVaultSecret(t, client, "settings", testhelpers.GoodSettingsJSON)

	// Add AppRole auth engine
	testhelpers.EnableAppRoleAuthEngine(t, client)

	// Create the role, get its ID and create a secret id
	testhelpers.CreateAppRoleRoleForReadSecretPolicy(t, client)
	roleId := testhelpers.GetAppRoleRoleID(t, client)
	secretId := testhelpers.CreateAppRoleSecretID(t, client)

	// Load configuration from vault using app-role
	settings, err := configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewVault().
			WithHost(vaultAddr).
			WithPath(testhelpers.PathFromSecretKey("settings")).
			WithAuth(loader.NewVaultAppRoleAuthMethod().
				WithRoleId(roleId).
				WithSecretId(secretId))).
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
