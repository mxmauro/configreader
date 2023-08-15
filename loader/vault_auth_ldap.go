package loader

import (
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/ldap"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// VaultLdapAuth contains the options to access vault with the LDAP authentication mechanism
type VaultLdapAuth struct {
	userName  string
	password  *ldap.Password
	mountPath ldap.LoginOption

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
		var err error

		userName, err = helpers.LoadAndReplaceEnvs(userName)
		if err == nil {
			a.userName = userName
		} else {
			a.err = err
		}
	}
	return a
}

// WithPassword sets the access password
func (a *VaultLdapAuth) WithPassword(password string) *VaultLdapAuth {
	if a.err == nil {
		var err error

		password, err = helpers.LoadAndReplaceEnvs(password)
		if err == nil {
			if len(password) > 0 {
				a.password = &ldap.Password{}
				a.password.FromString = password
			} else {
				a.password = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to ldap
func (a *VaultLdapAuth) WithMountPath(mountPath string) *VaultLdapAuth {
	if a.err == nil {
		var err error

		mountPath, err = helpers.LoadAndReplaceEnvs(mountPath)
		if err == nil {
			if len(mountPath) > 0 {
				a.mountPath = ldap.WithMountPath(mountPath)
			} else {
				a.mountPath = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

func (a *VaultLdapAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	opts := make([]ldap.LoginOption, 0)
	if len(a.userName) == 0 {
		return nil, errors.New("no user name specified for Vault's LDAP auth")
	}
	if a.password == nil {
		return nil, errors.New("no password specified for Vault's LDAP auth")
	}
	if a.mountPath != nil {
		opts = append(opts, a.mountPath)
	}

	// Return the authorization wrapper
	return ldap.NewLDAPAuth(a.userName, a.password, opts...)
}
