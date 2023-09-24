package loader

import (
	"crypto/sha256"
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/ldap"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// VaultLdapAuth contains the options to access vault with the LDAP authentication mechanism
type VaultLdapAuth struct {
	userName  string
	password  string
	mountPath string

	err error
}

// -----------------------------------------------------------------------------

// NewVaultLdapAuthMethod creates a new LDAP authentication method helper
func NewVaultLdapAuthMethod() *VaultLdapAuth {
	return &VaultLdapAuth{}
}

// WithUsername sets the username
func (a *VaultLdapAuth) WithUsername(userName string) *VaultLdapAuth {
	if a.err == nil {
		userName, a.err = helpers.LoadAndReplaceEnvs(userName)
		if a.err == nil {
			a.userName = userName
		}
	}
	return a
}

// WithPassword sets the access password
func (a *VaultLdapAuth) WithPassword(password string) *VaultLdapAuth {
	if a.err == nil {
		password, a.err = helpers.LoadAndReplaceEnvs(password)
		if a.err == nil {
			a.password = password
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to ldap
func (a *VaultLdapAuth) WithMountPath(mountPath string) *VaultLdapAuth {
	if a.err == nil {
		mountPath, a.err = helpers.LoadAndReplaceEnvs(mountPath)
		if a.err == nil {
			a.mountPath = mountPath
		}
	}
	return a
}

func (a *VaultLdapAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	if len(a.userName) == 0 {
		return nil, errors.New("no user name specified for Vault's LDAP auth")
	}
	if len(a.password) > 0 {
		return nil, errors.New("no password specified for Vault's LDAP auth")
	}

	opts := make([]ldap.LoginOption, 0)
	if len(a.mountPath) > 0 {
		opts = append(opts, ldap.WithMountPath(a.mountPath))
	}

	// Return the authorization wrapper
	return ldap.NewLDAPAuth(a.userName, &ldap.Password{
		FromString: a.password,
	}, opts...)
}

func (a *VaultLdapAuth) hash() (res [32]byte) {
	h := sha256.New()
	_, _ = h.Write([]byte(a.userName))
	_, _ = h.Write([]byte(a.password))
	_, _ = h.Write([]byte(a.mountPath))
	copy(res[:], h.Sum(nil))
	return
}
