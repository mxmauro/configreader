package testhelpers

import (
	"net"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/vault/api"
)

// -----------------------------------------------------------------------------

const (
	VaultReadSecretPolicy = "read-secret-policy"
)

// -----------------------------------------------------------------------------

func EnsureVaultAvailability(t *testing.T) string {
	// Get vault address
	vaultAddr := os.Getenv("VAULT_ADDR")
	if len(vaultAddr) == 0 {
		vaultAddr = "127.0.0.1:8200"
		t.Logf("VAULT_ADDR environment variable not set, assuming 127.0.0.1:8200")
	}

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

	// Return address
	return vaultAddr
}

func CreateRootVaultClient(t *testing.T, vaultAddr string) *api.Client {
	// Create vault accessor
	client, err := api.NewClient(&api.Config{
		Address: "http://" + vaultAddr,
	})
	if err != nil {
		t.Fatalf("unable to create Vault client [err=%v]", err.Error())
	}
	client.SetToken("root")

	// Create a policy for read-only access to "/secret/data"
	err = client.Sys().PutPolicy(VaultReadSecretPolicy, `
path "secret/data/*" {
  capabilities = ["read"]
}
`)
	if err != nil {
		t.Fatalf("unable to create the read secret policy [err=%v]", err.Error())
	}

	// Done
	return client
}

func WriteVaultSecrets(t *testing.T, client *api.Client, key string, secrets map[string]interface{}) {
	// Write the test secret. NOTE: We are following the K/V v2 specs here.
	data := map[string]interface{}{
		"data": secrets,
	}
	_, err := client.Logical().Write(PathFromSecretKey(key), data)
	if err != nil {
		t.Fatalf("unable to write Vault [key=%v] [err=%v]", key, err.Error())
	}
}

func PathFromSecretKey(key string) string {
	return "secret/data/go_config_reader_test/" + key
}
