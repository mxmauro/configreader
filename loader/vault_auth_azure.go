package loader

import (
	"crypto/sha256"
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/azure"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// VaultAzureAuth contains the options to access vault with the Azure authentication mechanism
type VaultAzureAuth struct {
	role      string
	mountPath string
	resource  string

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
		role, a.err = helpers.LoadAndReplaceEnvs(role)
		if a.err == nil {
			a.role = role
		}
	}
	return a
}

// WithResource sets an optional different resource URL. Defaults to Azure Public Cloud's ARM URL
func (a *VaultAzureAuth) WithResource(url string) *VaultAzureAuth {
	if a.err == nil {
		url, a.err = helpers.LoadAndReplaceEnvs(url)
		if a.err == nil {
			a.resource = url
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to azure
func (a *VaultAzureAuth) WithMountPath(mountPath string) *VaultAzureAuth {
	if a.err == nil {
		mountPath, a.err = helpers.LoadAndReplaceEnvs(mountPath)
		if a.err == nil {
			a.mountPath = mountPath
		}
	}
	return a
}

func (a *VaultAzureAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	if len(a.role) == 0 {
		return nil, errors.New("no role specified for Vault's Azure auth")
	}

	opts := make([]azure.LoginOption, 0)
	if len(a.mountPath) > 0 {
		opts = append(opts, azure.WithMountPath(a.mountPath))
	}
	if len(a.resource) > 0 {
		opts = append(opts, azure.WithResource(a.resource))
	}

	// Return the authorization wrapper
	return azure.NewAzureAuth(a.role, opts...)
}

func (a *VaultAzureAuth) hash() (res [32]byte) {
	h := sha256.New()
	_, _ = h.Write([]byte(a.role))
	_, _ = h.Write([]byte(a.mountPath))
	_, _ = h.Write([]byte(a.resource))
	copy(res[:], h.Sum(nil))
	return
}
