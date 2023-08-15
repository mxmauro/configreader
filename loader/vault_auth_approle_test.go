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
	appRoleRoleName = "test-approle"
)

//------------------------------------------------------------------------------

func TestVaultLoaderWithAppRoleAuth(t *testing.T) {
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

	// Add AppRole auth engine
	enableAppRoleAuthEngine(t, client)

	// Create the role, get its ID and create a secret id
	createAppRoleRoleForReadSecretPolicy(t, client)
	roleId := getAppRoleRoleID(t, client)
	secretId := createAppRoleSecretID(t, client)

	// Write the secrets
	testhelpers.WriteVaultSecret(t, client, "settings", testhelpers.GoodSettingsJSON)

	// Load configuration from vault using approle
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

func enableAppRoleAuthEngine(t *testing.T, client *api.Client) {
	// Enable AppRole auth method
	err := client.Sys().EnableAuthWithOptions("approle/", &api.EnableAuthOptions{
		Type:        "approle",
		Description: "AppRole auth method",
	})
	if err != nil {
		var apiErr *api.ResponseError
		if !(errors.As(err, &apiErr) &&
			apiErr.StatusCode == 400 &&
			len(apiErr.Errors) > 0 &&
			strings.Contains(apiErr.Errors[0], "already in use")) {
			t.Fatalf("unable to enable AppRole auth [err=%v]", err)
		}
	}
}

func createAppRoleRoleForReadSecretPolicy(t *testing.T, client *api.Client) {
	// Create an AppRole role with the created policy
	_, err := client.Logical().Write("auth/approle/role/"+appRoleRoleName, map[string]interface{}{
		"policies": []string{testhelpers.VaultReadSecretPolicy},
	})
	if err != nil {
		t.Fatalf("unable to create the test approle role [err=%v]", err)
	}
}

func getAppRoleRoleID(t *testing.T, client *api.Client) string {
	role, err := client.Logical().Read("auth/approle/role/" + appRoleRoleName + "/role-id")
	if err != nil {
		t.Fatalf("unable to get the approle role id [err=%v]", err)
	}
	if role == nil {
		t.Fatalf("unable to get the approle role id")
	}
	return role.Data["role_id"].(string)
}

func createAppRoleSecretID(t *testing.T, client *api.Client) string {
	secret, err := client.Logical().Write("auth/approle/role/"+appRoleRoleName+"/secret-id", nil)
	if err != nil {
		t.Fatalf("unable to generate the approle secret id [err=%v]", err)
	}
	if secret == nil {
		t.Fatalf("unable to generate the approle secret id")
	}
	return secret.Data["secret_id"].(string)
}
