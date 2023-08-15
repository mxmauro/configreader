package loader

import (
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// VaultKubernetesAuth contains the options to access vault with the Kubernetes authentication mechanism
type VaultKubernetesAuth struct {
	role         string
	accountToken kubernetes.LoginOption
	mountPath    kubernetes.LoginOption

	err error
}

// -----------------------------------------------------------------------------

// NewVaultKubernetesAuthMethod creates a new Kubernetes authentication method helper
func NewVaultKubernetesAuthMethod() *VaultKubernetesAuth {
	return &VaultKubernetesAuth{}
}

// WithRole sets the role
func (a *VaultKubernetesAuth) WithRole(role string) *VaultKubernetesAuth {
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

// WithAccountToken sets the account access token
func (a *VaultKubernetesAuth) WithAccountToken(token string) *VaultKubernetesAuth {
	if a.err == nil {
		var err error

		token, err = helpers.LoadAndReplaceEnvs(token)
		if err == nil {
			if len(token) > 0 {
				a.accountToken = kubernetes.WithServiceAccountToken(token)
			} else {
				a.accountToken = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to kubernetes
func (a *VaultKubernetesAuth) WithMountPath(mountPath string) *VaultKubernetesAuth {
	if a.err == nil {
		var err error

		mountPath, err = helpers.LoadAndReplaceEnvs(mountPath)
		if err == nil {
			if len(mountPath) > 0 {
				a.mountPath = kubernetes.WithMountPath(mountPath)
			} else {
				a.mountPath = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

func (a *VaultKubernetesAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	opts := make([]kubernetes.LoginOption, 0)
	if len(a.role) == 0 {
		return nil, errors.New("no role specified for Vault's Kubernetes auth")
	}
	if a.accountToken == nil {
		return nil, errors.New("no access token specified for Vault's Kubernetes auth")
	}
	opts = append(opts, a.accountToken)
	if a.mountPath != nil {
		opts = append(opts, a.mountPath)
	}

	// Return the authorization wrapper
	return kubernetes.NewKubernetesAuth(a.role, opts...)
}
