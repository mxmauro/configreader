package loader

import (
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
	mountPath              gcp.LoginOption
	typeId                 int
	_type                  gcp.LoginOption
	iamServiceAccountEmail string

	err error
}

// -----------------------------------------------------------------------------

// NewVaultGcpAuthMethod creates a new GCP authentication method helper
func NewVaultGcpAuthMethod() *VaultGcpAuth {
	return &VaultGcpAuth{
		typeId: VaultGcpAuthTypeGCE,
		_type:  gcp.WithGCEAuth(),
	}
}

// WithRole sets the role
func (a *VaultGcpAuth) WithRole(role string) *VaultGcpAuth {
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

// WithType sets the authentication type IAM or EC2
func (a *VaultGcpAuth) WithType(_type interface{}) *VaultGcpAuth {
	if a.err == nil {
		i, err := helpers.GetEnumEnv(_type, []helpers.EnumEnvAllowedValues{
			{"gce", VaultGcpAuthTypeGCE},
			{"iam", VaultGcpAuthTypeIAM},
		})
		if err == nil {
			switch i {
			case VaultGcpAuthTypeGCE:
				a.typeId = VaultGcpAuthTypeGCE
				a._type = gcp.WithGCEAuth()
			case VaultGcpAuthTypeIAM:
				a.typeId = VaultGcpAuthTypeIAM
				a._type = gcp.WithIAMAuth(a.iamServiceAccountEmail)
			}
		} else if errors.Is(err, helpers.ErrIsNil) {
			a.typeId = VaultGcpAuthTypeGCE
			a._type = gcp.WithGCEAuth()
		} else {
			a.err = errors.New("invalid type specified for Vault's Azure auth")
		}
	}
	return a
}

// WithTypeGCE sets the authentication type as GCE
func (a *VaultGcpAuth) WithTypeGCE() *VaultGcpAuth {
	if a.err == nil {
		a.typeId = VaultGcpAuthTypeGCE
		a._type = gcp.WithGCEAuth()
	}
	return a
}

// WithTypeIAM sets the authentication type as IAM
func (a *VaultGcpAuth) WithTypeIAM() *VaultGcpAuth {
	if a.err == nil {
		a.typeId = VaultGcpAuthTypeIAM
		a._type = gcp.WithIAMAuth(a.iamServiceAccountEmail)
	}
	return a
}

// WithIamServiceAccountEmail sets the service account email for IAM authentication type
func (a *VaultGcpAuth) WithIamServiceAccountEmail(email string) *VaultGcpAuth {
	if a.err == nil {
		var err error

		email, err = helpers.LoadAndReplaceEnvs(email)
		if err == nil {
			a.iamServiceAccountEmail = email
			if a.typeId == VaultGcpAuthTypeIAM {
				a._type = gcp.WithIAMAuth(a.iamServiceAccountEmail)
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to gcp
func (a *VaultGcpAuth) WithMountPath(mountPath string) *VaultGcpAuth {
	if a.err == nil {
		var err error

		mountPath, err = helpers.LoadAndReplaceEnvs(mountPath)
		if err == nil {
			if len(mountPath) > 0 {
				a.mountPath = gcp.WithMountPath(mountPath)
			} else {
				a.mountPath = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

func (a *VaultGcpAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	opts := make([]gcp.LoginOption, 0)
	if len(a.role) == 0 {
		return nil, errors.New("no role specified for Vault's GCP auth")
	}
	if a.mountPath != nil {
		opts = append(opts, a.mountPath)
	}
	opts = append(opts, a._type)

	// Return the authorization wrapper
	return gcp.NewGCPAuth(a.role, opts...)
}
