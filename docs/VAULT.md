## Hashicorp's Vault loader authorization methods

The following authorization methods are available.

### AppRole

`loader.NewVaultAppRoleAuthMethod()` creates a new `AppRole` authentication method helper.

| Method             | Description                                         |
|--------------------|-----------------------------------------------------|
| `WithRoleId`       | Sets the role id.                                   |
| `WithSecretId`     | Sets the secret id.                                 |
| `WithSecretUnwrap` | Specifies if the secret must be unwrapped.          |
| `WithMountPath`    | Sets an optional mount path. Defaults to `approle`. |

### Amazon Web Services

`loader.NewVaultAwsAuthMethod()` creates a new AWS authentication method helper.

| Method                  | Description                                                                                      |
|-------------------------|--------------------------------------------------------------------------------------------------|
| `WithRole`              | Sets the role.                                                                                   |
| `WithType`              | Sets the authentication type IAM or EC2.                                                         |
| `WithTypeEC2`           | Sets the authentication type as EC2.                                                             |
| `WithTypeIAM`           | Sets the authentication type as IAM.                                                             |
| `WithIamServerID`       | Sets the server id header when authenticating as IAM.                                            |
| `WithSignature`         | Tells the client which type of signature to use when verifying EC2 auth logins.                  |
| `WithIdentitySignature` | Tells the client to use the cryptographic identity document signature to verify EC2 auth logins. |
| `WithPKCS7Signature`    | Tells the client to use the PKCS #7 signature to verify EC2 auth logins.                         |
| `WithNonce`             | Sets nonce to use. Defaults to generate a random UUID.                                           |
| `WithRegion`            | Sets the region to use. Defaults to `us-east-1`.                                                 |
| `WithMountPath`         | Sets an optional mount path. Defaults to `aws`.                                                  |

### Microsoft Azure

`loader.NewVaultAzureAuthMethod()` creates a new Azure authentication method helper.

| Method          | Description                                                                        |
|-----------------|------------------------------------------------------------------------------------|
| `WithRole`      | Sets the role.                                                                     |
| `WithResource`  | Sets an optional different resource URL. Defaults to Azure Public Cloud's ARM URL. |
| `WithMountPath` | Sets an optional mount path. Defaults to `azure`.                                  |

### Google Cloud Platform

`loader.NewVaultGcpAuthMethod()` creates a new GCP authentication method helper.

| Method                       | Description                                                                        |
|------------------------------|------------------------------------------------------------------------------------|
| `WithRole`                   | Sets the role.                                                                     |
| `WithResource`               | Sets an optional different resource URL. Defaults to Azure Public Cloud's ARM URL. |
| `WithType`                   | Sets the authentication type IAM or EC2.                                           |
| `WithTypeGCE`                | Sets the authentication type as GCE.                                               |
| `WithTypeIAM`                | Sets the authentication type as IAM.                                               |
| `WithIamServiceAccountEmail` | Sets the service account email for IAM authentication type.                        |
| `WithMountPath`              | Sets an optional mount path. Defaults to `gcp`.                                    |

### Kubernetes

`loader.NewVaultKubernetesAuthMethod()` creates a new K8S authentication method helper.

| Method             | Description                                            |
|--------------------|--------------------------------------------------------|
| `WithRole`         | Sets the role.                                         |
| `WithAccountToken` | Sets the account access token.                         |
| `WithMountPath`    | Sets an optional mount path. Defaults to `kubernetes`. |

### LDAP

`loaderNewVaultLdapAuthMethod()` creates a new LDAP authentication method helper.

| Method          | Description                                      |
|-----------------|--------------------------------------------------|
| `WithUsername`  | Sets the username.                               |
| `WithPassword`  | Sets the access password.                        |
| `WithMountPath` | Sets an optional mount path. Defaults to `ldap`. |
