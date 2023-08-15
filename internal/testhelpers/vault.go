package testhelpers

import (
	"net"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
)

//------------------------------------------------------------------------------

const (
	VaultReadSecretPolicy = "read-secret-policy"
)

//------------------------------------------------------------------------------

func EnsureVaultAvailability(t *testing.T, vaultAddr string) {
	conn, err := net.DialTimeout("tcp", vaultAddr, time.Second)
	if err != nil {
		t.Logf("Vault server is not running. Skipping....")
		t.Logf("")
		t.Logf("Use the following command in order to run Vault tests:")
		t.Logf("")
		t.Logf(`    vault server -dev -dev-root-token-id="root" -dev-listen-address="0.0.0.0:8200" -dev-no-store-token`)
		t.Logf("")
		t.Logf("If you are running Vault on an Amazon EC2 instance, ensure it has attached an IAM role with the following policies:")
		t.Logf("")
		t.Logf("    * ec2:DescribeInstances")
		t.Logf("    * iam:GetInstanceProfile (if IAM Role binding is used)")
		t.SkipNow()
	}
	if conn != nil {
		_ = conn.Close()
	}
}

func CreateRootVaultClient(t *testing.T, vaultAddr string) *api.Client {
	// Create vault accessor
	client, err := api.NewClient(&api.Config{
		Address: "http://" + vaultAddr,
	})
	if err != nil {
		t.Fatalf("unable to create Vault client [err=%v]", err)
	}
	client.SetToken("root")

	// Create a policy for read-only access to "/secret/data"
	err = client.Sys().PutPolicy(VaultReadSecretPolicy, `
path "secret/data/*" {
  capabilities = ["read"]
}
`)
	if err != nil {
		t.Fatalf("unable to create the read secret policy [err=%v]", err)
	}

	// Done
	return client
}

func WriteVaultSecret(t *testing.T, client *api.Client, key string, secret string) {
	// Write the test secret. NOTE: We are following the K/V v2 specs here.
	value := `{ "data": ` + secret + `}`
	_, err := client.Logical().WriteBytes(PathFromSecretKey(key), []byte(value))
	if err != nil {
		t.Fatalf("unable to write Vault [key=%v] [err=%v]", key, err)
	}
}

func PathFromSecretKey(key string) string {
	return "secret/data/go_config_reader_test/" + key
}