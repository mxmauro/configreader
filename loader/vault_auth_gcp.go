package loader

import (
	"crypto/sha256"
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/gcp"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

const (
	VaultGcpAuthTypeGCE = iota + 1
	VaultGcpAuthTypeIAM
)

// -----------------------------------------------------------------------------

// VaultGcpAuth contains the options to access vault with the GCP authentication mechanism
type VaultGcpAuth struct {
	role                   string
	mountPath              string
	_type                  int
	iamServiceAccountEmail string

	err error
}

// -----------------------------------------------------------------------------

// NewVaultGcpAuthMethod creates a new GCP authentication method helper
func NewVaultGcpAuthMethod() *VaultGcpAuth {
	return &VaultGcpAuth{
		_type: VaultGcpAuthTypeGCE,
	}
}

// WithRole sets the role
func (a *VaultGcpAuth) WithRole(role string) *VaultGcpAuth {
	if a.err == nil {
		var err error

		role, err = helpers.ExpandEnvVars(role)
		if err == nil {
			a.role = role
		} else {
			a.err = err
		}
	}
	return a
}

// WithType sets the authentication type IAM or EC2
func (a *VaultGcpAuth) WithType(_type interface{}) *VaultGcpAuth {
	if a.err == nil {
		i, err := helpers.GetEnumEnv(_type, []helpers.EnumEnvAllowedValues{
			{"gce", VaultGcpAuthTypeGCE},
			{"iam", VaultGcpAuthTypeIAM},
		})
		if err == nil {
			a._type = i
		} else if errors.Is(err, helpers.ErrIsNil) {
			a._type = VaultGcpAuthTypeGCE
		} else {
			a.err = errors.New("invalid type specified for Vault's Azure auth")
		}
	}
	return a
}

// WithTypeGCE sets the authentication type as GCE
func (a *VaultGcpAuth) WithTypeGCE() *VaultGcpAuth {
	if a.err == nil {
		a._type = VaultGcpAuthTypeGCE
	}
	return a
}

// WithTypeIAM sets the authentication type as IAM
func (a *VaultGcpAuth) WithTypeIAM() *VaultGcpAuth {
	if a.err == nil {
		a._type = VaultGcpAuthTypeIAM
	}
	return a
}

// WithIamServiceAccountEmail sets the service account email for IAM authentication type
func (a *VaultGcpAuth) WithIamServiceAccountEmail(email string) *VaultGcpAuth {
	if a.err == nil {
		email, a.err = helpers.ExpandEnvVars(email)
		if a.err == nil {
			a.iamServiceAccountEmail = email
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to gcp
func (a *VaultGcpAuth) WithMountPath(mountPath string) *VaultGcpAuth {
	if a.err == nil {
		mountPath, a.err = helpers.ExpandEnvVars(mountPath)
		if a.err == nil {
			a.mountPath = mountPath
		}
	}
	return a
}

func (a *VaultGcpAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	if len(a.role) == 0 {
		return nil, errors.New("no role specified for Vault's GCP auth")
	}

	opts := make([]gcp.LoginOption, 0)
	if len(a.mountPath) > 0 {
		opts = append(opts, gcp.WithMountPath(a.mountPath))
	}
	switch a._type {
	case VaultGcpAuthTypeGCE:
		opts = append(opts, gcp.WithGCEAuth())
	case VaultGcpAuthTypeIAM:
		opts = append(opts, gcp.WithIAMAuth(a.iamServiceAccountEmail))
	}

	// Return the authorization wrapper
	return gcp.NewGCPAuth(a.role, opts...)
}

func (a *VaultGcpAuth) hash() (res [32]byte) {
	h := sha256.New()
	_, _ = h.Write([]byte(a.role))
	_, _ = h.Write([]byte(a.mountPath))
	_, _ = h.Write([]byte{byte(a._type)})
	_, _ = h.Write([]byte(a.iamServiceAccountEmail))
	copy(res[:], h.Sum(nil))
	return
}
