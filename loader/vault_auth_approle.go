package loader

import (
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// VaultAppRoleAuth contains the options to access vault with the AppRole authentication mechanism
type VaultAppRoleAuth struct {
	roleId      string
	secretId    *approle.SecretID
	mountPath   approle.LoginOption
	unwrapToken approle.LoginOption

	err error
}

// -----------------------------------------------------------------------------

// NewVaultAppRoleAuthMethod creates a new AppRole authentication method helper
func NewVaultAppRoleAuthMethod() *VaultAppRoleAuth {
	return &VaultAppRoleAuth{}
}

// WithRoleId sets the role id
func (a *VaultAppRoleAuth) WithRoleId(roleId string) *VaultAppRoleAuth {
	if a.err == nil {
		var err error

		roleId, err = helpers.LoadAndReplaceEnvs(roleId)
		if err == nil {
			a.roleId = roleId
		} else {
			a.err = err
		}
	}
	return a
}

// WithSecretId sets the secret id
func (a *VaultAppRoleAuth) WithSecretId(secretId string) *VaultAppRoleAuth {
	if a.err == nil {
		var err error

		secretId, err = helpers.LoadAndReplaceEnvs(secretId)
		if err == nil {
			if len(secretId) > 0 {
				a.secretId = &approle.SecretID{
					FromString: secretId,
				}
			} else {
				a.secretId = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithSecretUnwrap specifies if the secret must be unwrapped
func (a *VaultAppRoleAuth) WithSecretUnwrap(unwrap interface{}) *VaultAppRoleAuth {
	if a.err == nil {
		b, err := helpers.GetBoolEnv(unwrap)
		if err == nil {
			if b {
				a.unwrapToken = approle.WithWrappingToken()
			} else {
				a.unwrapToken = nil
			}
		} else if errors.Is(err, helpers.ErrIsNil) {
			a.unwrapToken = nil
		} else {
			a.err = err
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to approle
func (a *VaultAppRoleAuth) WithMountPath(mountPath string) *VaultAppRoleAuth {
	if a.err == nil {
		var err error

		mountPath, err = helpers.LoadAndReplaceEnvs(mountPath)
		if err == nil {
			if len(mountPath) > 0 {
				a.mountPath = approle.WithMountPath(mountPath)
			} else {
				a.mountPath = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

func (a *VaultAppRoleAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	// Get the Role ID
	if len(a.roleId) == 0 {
		return nil, errors.New("no roleID specified for Vault's AppRole auth")
	}

	// Get secret ID
	if a.secretId == nil {
		return nil, errors.New("no secretID specified for Vault's AppRole auth")
	}

	opts := make([]approle.LoginOption, 0)
	if a.mountPath != nil {
		opts = append(opts, a.mountPath)
	}
	if a.unwrapToken != nil {
		opts = append(opts, a.unwrapToken)
	}

	// Return the authorization wrapper
	return approle.NewAppRoleAuth(a.roleId, a.secretId, opts...)
}
