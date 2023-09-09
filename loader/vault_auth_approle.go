package loader

import (
	"crypto/sha256"
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/approle"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// VaultAppRoleAuth contains the options to access vault with the AppRole authentication mechanism
type VaultAppRoleAuth struct {
	roleId      string
	secretId    string
	mountPath   string
	unwrapToken bool

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
		roleId, a.err = helpers.LoadAndReplaceEnvs(roleId)
		if a.err == nil {
			a.roleId = roleId
		}
	}
	return a
}

// WithSecretId sets the secret id
func (a *VaultAppRoleAuth) WithSecretId(secretId string) *VaultAppRoleAuth {
	if a.err == nil {
		secretId, a.err = helpers.LoadAndReplaceEnvs(secretId)
		if a.err == nil {
			a.secretId = secretId
		}
	}
	return a
}

// WithSecretUnwrap specifies if the secret must be unwrapped
func (a *VaultAppRoleAuth) WithSecretUnwrap(unwrap interface{}) *VaultAppRoleAuth {
	if a.err == nil {
		b, err := helpers.GetBoolEnv(unwrap)
		if err == nil {
			a.unwrapToken = b
		} else if errors.Is(err, helpers.ErrIsNil) {
			a.unwrapToken = false
		} else {
			a.err = err
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to approle
func (a *VaultAppRoleAuth) WithMountPath(mountPath string) *VaultAppRoleAuth {
	if a.err == nil {
		mountPath, a.err = helpers.LoadAndReplaceEnvs(mountPath)
		if a.err == nil {
			a.mountPath = mountPath
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
	if len(a.secretId) == 0 {
		return nil, errors.New("no secretID specified for Vault's AppRole auth")
	}

	opts := make([]approle.LoginOption, 0)
	if len(a.mountPath) > 0 {
		opts = append(opts, approle.WithMountPath(a.mountPath))
	}
	if a.unwrapToken {
		opts = append(opts, approle.WithWrappingToken())
	}

	// Return the authorization wrapper
	return approle.NewAppRoleAuth(a.roleId, &approle.SecretID{
		FromString: a.secretId,
	}, opts...)
}

func (a *VaultAppRoleAuth) hash() (res [32]byte) {
	h := sha256.New()
	_, _ = h.Write([]byte(a.roleId))
	_, _ = h.Write([]byte(a.secretId))
	_, _ = h.Write([]byte(a.mountPath))
	if a.unwrapToken {
		_, _ = h.Write([]byte{1})
	} else {
		_, _ = h.Write([]byte{0})
	}
	copy(res[:], h.Sum(nil))
	return
}
