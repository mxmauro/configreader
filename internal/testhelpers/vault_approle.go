package testhelpers

import (
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/vault/api"
)

//------------------------------------------------------------------------------

const (
	VaultAppRoleRoleName = "test-approle"

	VaultAppRoleRenewPolicy = "approle-renew-policy"
)

//------------------------------------------------------------------------------

func EnableAppRoleAuthEngine(t *testing.T, client *api.Client) {
	// Enable AppRole auth method
	err := client.Sys().EnableAuthWithOptions("approle/", &api.EnableAuthOptions{
		Type:        "approle",
		Description: "AppRole auth method",
	})
	if err != nil {
		var apiErr *api.ResponseError

		if !(errors.As(err, &apiErr) && apiErr.StatusCode == 400 && len(apiErr.Errors) > 0 &&
			strings.Contains(apiErr.Errors[0], "already in use")) {
			t.Fatalf("unable to enable AppRole auth [err=%v]", err)
		}
	}

	// Create a policy for read-only access to "/auth/token/renew-self"
	err = client.Sys().PutPolicy(VaultAppRoleRenewPolicy, `
path "/auth/token/renew-self" {
  capabilities = ["update"]
}
`)
	if err != nil {
		t.Fatalf("unable to create the auth token self renew policy [err=%v]", err)
	}
	/*
		var apiErr *api.ResponseError

		if !(errors.As(err, &apiErr) && apiErr.StatusCode == 400 && len(apiErr.Errors) > 0 &&
			strings.Contains(apiErr.Errors[0], "already in use")) {
			t.Fatalf("unable to enable AppRole auth [err=%v]", err)
		}

	*/
}

func CreateAppRoleRoleForReadSecretPolicy(t *testing.T, client *api.Client) {
	// Create an AppRole role with the created policy
	_, err := client.Logical().Write("auth/approle/role/"+VaultAppRoleRoleName, map[string]interface{}{
		"policies":  []string{VaultReadSecretPolicy, VaultAppRoleRenewPolicy},
		"token_ttl": 30, // Tokens must be renewed within 30 seconds
	})
	if err != nil {
		t.Fatalf("unable to create the test approle role [err=%v]", err)
	}
}

func GetAppRoleRoleID(t *testing.T, client *api.Client) string {
	role, err := client.Logical().Read("auth/approle/role/" + VaultAppRoleRoleName + "/role-id")
	if err != nil {
		t.Fatalf("unable to get the approle role id [err=%v]", err)
	}
	if role == nil {
		t.Fatalf("unable to get the approle role id")
	}
	return role.Data["role_id"].(string)
}

func CreateAppRoleSecretID(t *testing.T, client *api.Client) string {
	secret, err := client.Logical().Write("auth/approle/role/"+VaultAppRoleRoleName+"/secret-id", nil)
	if err != nil {
		t.Fatalf("unable to generate the approle secret id [err=%v]", err)
	}
	if secret == nil {
		t.Fatalf("unable to generate the approle secret id")
	}
	return secret.Data["secret_id"].(string)
}
