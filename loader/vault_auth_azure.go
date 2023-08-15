package loader

import (
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/azure"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// VaultAzureAuth contains the options to access vault with the Azure authentication mechanism
type VaultAzureAuth struct {
	opts []azure.LoginOption

	role      string
	mountPath azure.LoginOption
	resource  azure.LoginOption

	err error
}

// -----------------------------------------------------------------------------

// NewVaultAzureAuthMethod creates a new Azure authentication method helper
func NewVaultAzureAuthMethod() *VaultAzureAuth {
	return &VaultAzureAuth{}
}

// WithRole sets the role
func (a *VaultAzureAuth) WithRole(role string) *VaultAzureAuth {
	if a.err == nil {
		var err error

		role, err = helpers.LoadAndReplaceEnvs(role)
		if err == nil {
			a.role = role
		} else {
			a.err = err
		}
	}
	return a
}

// WithResource sets an optional different resource URL. Defaults to Azure Public Cloud's ARM URL
func (a *VaultAzureAuth) WithResource(url string) *VaultAzureAuth {
	if a.err == nil {
		var err error

		url, err = helpers.LoadAndReplaceEnvs(url)
		if err == nil {
			if len(url) > 0 {
				a.resource = azure.WithResource(url)
			} else {
				a.resource = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to azure
func (a *VaultAzureAuth) WithMountPath(mountPath string) *VaultAzureAuth {
	if a.err == nil {
		var err error

		mountPath, err = helpers.LoadAndReplaceEnvs(mountPath)
		if err == nil {
			if len(mountPath) > 0 {
				a.mountPath = azure.WithMountPath(mountPath)
			} else {
				a.mountPath = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

func (a *VaultAzureAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	opts := make([]azure.LoginOption, 0)
	if len(a.role) == 0 {
		return nil, errors.New("no role specified for Vault's Azure auth")
	}
	if a.mountPath != nil {
		opts = append(opts, a.mountPath)
	}
	if a.resource != nil {
		opts = append(opts, a.resource)
	}

	// Return the authorization wrapper
	return azure.NewAzureAuth(a.role, opts...)
}
