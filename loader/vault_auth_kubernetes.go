package loader

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/kubernetes"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// VaultKubernetesAuth contains the options to access vault with the Kubernetes authentication mechanism
type VaultKubernetesAuth struct {
	role         string
	accountToken string
	mountPath    string

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
		token, a.err = helpers.LoadAndReplaceEnvs(token)
		if a.err == nil {
			a.accountToken = token
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to kubernetes
func (a *VaultKubernetesAuth) WithMountPath(mountPath string) *VaultKubernetesAuth {
	if a.err == nil {
		mountPath, a.err = helpers.LoadAndReplaceEnvs(mountPath)
		if a.err == nil {
			a.mountPath = mountPath
		}
	}
	return a
}

func (a *VaultKubernetesAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	if len(a.role) == 0 {
		return nil, errors.New("no role specified for Vault's Kubernetes auth")
	}

	opts := make([]kubernetes.LoginOption, 0)

	token := a.accountToken
	if len(token) == 0 {
		for _, filename := range k8sTokenFiles {
			fileInfo, err := os.Stat(filename)
			if err == nil && fileInfo.Size() > 0 {
				var content []byte

				content, err = os.ReadFile(filename)
				if err != nil {
					return nil, fmt.Errorf("unable to read K8S service account token file [err=%w]", err)
				}

				token = string(content)
				break
			}
		}

		if len(token) == 0 {
			return nil, errors.New("no access token specified for Vault's Kubernetes auth and unable to locate K8S service account token file")
		}
	}
	opts = append(opts, kubernetes.WithServiceAccountToken(token))

	if len(a.mountPath) > 0 {
		opts = append(opts, kubernetes.WithMountPath(a.mountPath))
	}

	// Return the authorization wrapper
	return kubernetes.NewKubernetesAuth(a.role, opts...)
}

func (a *VaultKubernetesAuth) hash() (res [32]byte) {
	h := sha256.New()
	_, _ = h.Write([]byte(a.role))
	_, _ = h.Write([]byte(a.accountToken))
	_, _ = h.Write([]byte(a.mountPath))
	copy(res[:], h.Sum(nil))
	return
}
