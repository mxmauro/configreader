package loader

import (
	"crypto/sha256"
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
	VaultAwsAuthSignatureRSA2048
)

// VaultAwsAuth contains the options to access vault with the AWS authentication mechanism
type VaultAwsAuth struct {
	role        string
	mountPath   string
	_type       int
	signature   int
	serverIdHdr string
	nonce       string
	region      string

	err error
}

// -----------------------------------------------------------------------------

// NewVaultAwsAuthMethod creates a new AWS authentication method helper
func NewVaultAwsAuthMethod() *VaultAwsAuth {
	return &VaultAwsAuth{
		_type:     VaultAwsAuthTypeIAM,        //aws.WithIAMAuth(),
		signature: VaultAwsAuthSignaturePKCS7, //aws.WithPKCS7Signature(),
	}
}

// WithRole sets the role
func (a *VaultAwsAuth) WithRole(role string) *VaultAwsAuth {
	if a.err == nil {
		role, a.err = helpers.LoadAndReplaceEnvs(role)
		if a.err == nil {
			a.role = role
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
			a._type = i
		} else if errors.Is(err, helpers.ErrIsNil) {
			a._type = VaultAwsAuthTypeIAM
		} else {
			a.err = errors.New("invalid type specified for Vault's AWS auth")
		}
	}
	return a
}

// WithTypeIAM sets the authentication type as IAM
func (a *VaultAwsAuth) WithTypeIAM() *VaultAwsAuth {
	if a.err == nil {
		a._type = VaultAwsAuthTypeIAM
	}
	return a
}

// WithTypeEC2 sets the authentication type as EC2
func (a *VaultAwsAuth) WithTypeEC2() *VaultAwsAuth {
	if a.err == nil {
		a._type = VaultAwsAuthTypeEC2
	}
	return a
}

// WithIamServerID sets the server id header when authenticating as IAM
func (a *VaultAwsAuth) WithIamServerID(id string) *VaultAwsAuth {
	if a.err == nil {
		id, a.err = helpers.LoadAndReplaceEnvs(id)
		if a.err == nil {
			a.serverIdHdr = id
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
			{"rsa2048", VaultAwsAuthSignatureRSA2048},
		})
		if err == nil {
			a.signature = i
		} else if errors.Is(err, helpers.ErrIsNil) {
			a.signature = VaultAwsAuthSignaturePKCS7
		} else {
			a.err = errors.New("invalid signature type specified for Vault's AWS auth")
		}
	}
	return a
}

// WithIdentitySignature tells the client to use the cryptographic identity document signature to verify EC2 auth logins
func (a *VaultAwsAuth) WithIdentitySignature() *VaultAwsAuth {
	if a.err == nil {
		a.signature = VaultAwsAuthSignatureIdentity
	}
	return a
}

// WithPKCS7Signature tells the client to use the PKCS #7 signature to verify EC2 auth logins
func (a *VaultAwsAuth) WithPKCS7Signature() *VaultAwsAuth {
	if a.err == nil {
		a.signature = VaultAwsAuthSignaturePKCS7
	}
	return a
}

// WithRSA2048Signature tells the client to use the RSA 2048 signature to verify EC2 auth logins
func (a *VaultAwsAuth) WithRSA2048Signature() *VaultAwsAuth {
	if a.err == nil {
		a.signature = VaultAwsAuthSignatureRSA2048
	}
	return a
}

// WithNonce sets nonce to use. Defaults to generate a random uuid
func (a *VaultAwsAuth) WithNonce(nonce string) *VaultAwsAuth {
	if a.err == nil {
		nonce, a.err = helpers.LoadAndReplaceEnvs(nonce)
		if a.err == nil {
			a.nonce = nonce
		}
	}
	return a
}

// WithRegion sets the region to use. Defaults to us-east-1
func (a *VaultAwsAuth) WithRegion(region string) *VaultAwsAuth {
	if a.err == nil {
		region, a.err = helpers.LoadAndReplaceEnvs(region)
		if a.err == nil {
			a.region = region
		}
	}
	return a
}

// WithMountPath sets an optional mount path. Defaults to aws
func (a *VaultAwsAuth) WithMountPath(mountPath string) *VaultAwsAuth {
	if a.err == nil {
		mountPath, a.err = helpers.LoadAndReplaceEnvs(mountPath)
		if a.err == nil {
			a.mountPath = mountPath
		}
	}
	return a
}

func (a *VaultAwsAuth) create() (api.AuthMethod, error) {
	// If an error was set by a With... function, return it
	if a.err != nil {
		return nil, a.err
	}

	if len(a.role) == 0 {
		return nil, errors.New("no role specified for Vault's AWS auth")
	}

	opts := make([]aws.LoginOption, 0)
	opts = append(opts, aws.WithRole(a.role))
	if len(a.mountPath) > 0 {
		opts = append(opts, aws.WithMountPath(a.mountPath))
	}
	switch a._type {
	case VaultAwsAuthTypeIAM:
		opts = append(opts, aws.WithIAMAuth())
	case VaultAwsAuthTypeEC2:
		opts = append(opts, aws.WithEC2Auth())
	}
	switch a.signature {
	case VaultAwsAuthSignatureIdentity:
		opts = append(opts, aws.WithIdentitySignature())
	case VaultAwsAuthSignaturePKCS7:
		opts = append(opts, aws.WithPKCS7Signature())
	case VaultAwsAuthSignatureRSA2048:
		opts = append(opts, aws.WithRSA2048Signature())

	}
	if len(a.serverIdHdr) > 0 {
		opts = append(opts, aws.WithIAMServerIDHeader(a.serverIdHdr))
	}
	if len(a.nonce) > 0 {
		opts = append(opts, aws.WithNonce(a.nonce))
	}
	if len(a.region) > 0 {
		opts = append(opts, aws.WithRegion(a.region))
	}

	// Return the authorization wrapper
	return aws.NewAWSAuth(opts...)
}

func (a *VaultAwsAuth) hash() (res [32]byte) {
	h := sha256.New()
	_, _ = h.Write([]byte(a.role))
	_, _ = h.Write([]byte(a.mountPath))
	_, _ = h.Write([]byte{byte(a._type)})
	_, _ = h.Write([]byte{byte(a.signature)})
	_, _ = h.Write([]byte(a.serverIdHdr))
	_, _ = h.Write([]byte(a.nonce))
	_, _ = h.Write([]byte(a.region))
	copy(res[:], h.Sum(nil))
	return
}
