package loader

import (
	"errors"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/api/auth/aws"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

const (
	VaultAwsAuthTypeIAM = iota + 1
	VaultAwsAuthTypeEC2
)

const (
	VaultAwsAuthSignatureIdentity = iota + 1
	VaultAwsAuthSignaturePKCS7
)

// VaultAwsAuth contains the options to access vault with the AWS authentication mechanism
type VaultAwsAuth struct {
	role        aws.LoginOption
	mountPath   aws.LoginOption
	_type       aws.LoginOption
	signature   aws.LoginOption
	serverIdHdr aws.LoginOption
	nonce       aws.LoginOption
	region      aws.LoginOption

	err error
}

// -----------------------------------------------------------------------------

// NewVaultAwsAuthMethod creates a new AWS authentication method helper
func NewVaultAwsAuthMethod() *VaultAwsAuth {
	return &VaultAwsAuth{
		_type:     aws.WithIAMAuth(),
		signature: aws.WithPKCS7Signature(),
	}
}

// WithRole sets the role
func (a *VaultAwsAuth) WithRole(role string) *VaultAwsAuth {
	if a.err == nil {
		var err error

		role, err = helpers.LoadAndReplaceEnvs(role)
		if err == nil {
			if len(role) > 0 {
				a.role = aws.WithRole(role)
			} else {
				a.role = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithType sets the authentication type IAM or EC2
func (a *VaultAwsAuth) WithType(_type interface{}) *VaultAwsAuth {
	if a.err == nil {
		i, err := helpers.GetEnumEnv(_type, []helpers.EnumEnvAllowedValues{
			{"iam", VaultAwsAuthTypeIAM},
			{"ec2", VaultAwsAuthTypeEC2},
		})
		if err == nil {
			switch i {
			case VaultAwsAuthTypeIAM:
				a._type = aws.WithIAMAuth()
			case VaultAwsAuthTypeEC2:
				a._type = aws.WithEC2Auth()
			}
		} else if errors.Is(err, helpers.ErrIsNil) {
			a._type = aws.WithIAMAuth()
		} else {
			a.err = errors.New("invalid type specified for Vault's AWS auth")
		}
	}
	return a
}

// WithTypeIAM sets the authentication type as IAM
func (a *VaultAwsAuth) WithTypeIAM() *VaultAwsAuth {
	if a.err == nil {
		a._type = aws.WithIAMAuth()
	}
	return a
}

// WithTypeEC2 sets the authentication type as EC2
func (a *VaultAwsAuth) WithTypeEC2() *VaultAwsAuth {
	if a.err == nil {
		a._type = aws.WithEC2Auth()
	}
	return a
}

// WithIamServerID sets the server id header when authenticating as IAM
func (a *VaultAwsAuth) WithIamServerID(id string) *VaultAwsAuth {
	if a.err == nil {
		var err error

		id, err = helpers.LoadAndReplaceEnvs(id)
		if err == nil {
			if len(id) > 0 {
				a.serverIdHdr = aws.WithIAMServerIDHeader(id)
			} else {
				a.serverIdHdr = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithSignature tells the client which type of signature to use when verifying EC2 auth logins
func (a *VaultAwsAuth) WithSignature(signature interface{}) *VaultAwsAuth {
	if a.err == nil {
		i, err := helpers.GetEnumEnv(signature, []helpers.EnumEnvAllowedValues{
			{"identity", VaultAwsAuthSignatureIdentity},
			{"pkcs7", VaultAwsAuthSignaturePKCS7},
		})
		if err == nil {
			switch i {
			case VaultAwsAuthSignatureIdentity:
				a.signature = aws.WithIdentitySignature()
			case VaultAwsAuthSignaturePKCS7:
				a.signature = aws.WithPKCS7Signature()
			}
		} else if errors.Is(err, helpers.ErrIsNil) {
			a.signature = aws.WithPKCS7Signature()
		} else {
			a.err = errors.New("invalid signature type specified for Vault's AWS auth")
		}
	}
	return a
}

// WithIdentitySignature tells the client to use the cryptographic identity document signature to verify EC2 auth logins
func (a *VaultAwsAuth) WithIdentitySignature() *VaultAwsAuth {
	if a.err == nil {
		a.signature = aws.WithIdentitySignature()
	}
	return a
}

// WithPKCS7Signature tells the client to use the PKCS #7 signature to verify EC2 auth logins
func (a *VaultAwsAuth) WithPKCS7Signature() *VaultAwsAuth {
	if a.err == nil {
		a.signature = aws.WithPKCS7Signature()
	}
	return a
}

// WithNonce sets nonce to use. Defaults to generate a random uuid
func (a *VaultAwsAuth) WithNonce(nonce string) *VaultAwsAuth {
	if a.err == nil {
		var err error

		nonce, err = helpers.LoadAndReplaceEnvs(nonce)
		if err == nil {
			if len(nonce) > 0 {
				a.nonce = aws.WithNonce(nonce)
			} else {
				a.nonce = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithRegion sets the region to use. Defaults to us-east-1
func (a *VaultAwsAuth) WithRegion(region string) *VaultAwsAuth {
	if a.err == nil {
		var err error

		region, err = helpers.LoadAndReplaceEnvs(region)
		if err == nil {
			if len(region) > 0 {
				a.region = aws.WithRegion(region)
			} else {
				a.region = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to aws
func (a *VaultAwsAuth) WithMountPath(mountPath string) *VaultAwsAuth {
	if a.err == nil {
		var err error

		mountPath, err = helpers.LoadAndReplaceEnvs(mountPath)
		if err == nil {
			if len(mountPath) > 0 {
				a.mountPath = aws.WithMountPath(mountPath)
			} else {
				a.mountPath = nil
			}
		} else {
			a.err = err
		}
	}
	return a
}

func (a *VaultAwsAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	opts := make([]aws.LoginOption, 0)
	if a.role == nil {
		return nil, errors.New("no role specified for Vault's AWS auth")
	}
	opts = append(opts, a.role)
	if a.mountPath != nil {
		opts = append(opts, a.mountPath)
	}
	opts = append(opts, a._type)
	opts = append(opts, a.signature)
	if a.serverIdHdr != nil {
		opts = append(opts, a.serverIdHdr)
	}
	if a.nonce != nil {
		opts = append(opts, a.nonce)
	}
	if a.region != nil {
		opts = append(opts, a.region)
	}

	// Return the authorization wrapper
	return aws.NewAWSAuth(opts...)
}
