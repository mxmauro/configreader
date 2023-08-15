package loader

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/hashicorp/vault/api"
	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// Vault wraps content to be loaded from a Hashicorp Vault instance
type Vault struct {
	host string
	path string

	headers map[string]string

	tlsConfig *tls.Config

	accessToken string
	auth        api.AuthMethod

	err error
}

// VaultAuthMethod is an interface used to set up an authentication mechanism
type VaultAuthMethod interface {
	create() (api.AuthMethod, error)
}

// -----------------------------------------------------------------------------

// NewVault create a new Hashicorp Vault loader
func NewVault() *Vault {
	return &Vault{
		headers: make(map[string]string),
	}
}

// WithHost sets the host address and, optionally, the port
func (l *Vault) WithHost(host string) *Vault {
	if l.err == nil {
		host, l.err = helpers.LoadAndReplaceEnvs(host)
		if l.err == nil {
			l.host = host
		}
	}
	return l
}

// WithPath sets the path
func (l *Vault) WithPath(path string) *Vault {
	if l.err == nil {
		path, l.err = helpers.LoadAndReplaceEnvs(path)
		if l.err == nil {
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}
			l.path = path
		}
	}
	return l
}

// WithHeaders sets the request headers
func (l *Vault) WithHeaders(headers map[string]string) *Vault {
	if l.err == nil {
		var err error

		headersCopy := make(map[string]string)
		for key, value := range headers {
			if len(key) == 0 {
				err = errors.New("invalid header value")
				break
			}

			value, err = helpers.LoadAndReplaceEnvs(value)
			if err != nil {
				break
			}

			if len(value) > 0 {
				headersCopy[key] = value
			}
		}

		if l.err == nil {
			l.headers = headersCopy
		} else {
			l.err = err
		}
	}
	return l
}

// WithHeaderItem sets a single request header
func (l *Vault) WithHeaderItem(key string, value string) *Vault {
	if l.err == nil {
		var err error

		value, err = helpers.LoadAndReplaceEnvs(value)
		if err == nil {
			if l.headers == nil {
				l.headers = make(map[string]string)
			}
			l.headers[key] = value
		} else {
			l.err = err
		}
	}
	return l
}

// WithDefaultTLS sets a default tls.Config object
func (l *Vault) WithDefaultTLS() *Vault {
	if l.err == nil {
		l.tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	return l
}

// WithTLS sets a tls.Config object
func (l *Vault) WithTLS(tlsConfig *tls.Config) *Vault {
	if l.err == nil {
		l.tlsConfig = tlsConfig
	}
	return l
}

// WithAccessToken sets the access token to use as authorization
func (l *Vault) WithAccessToken(token string) *Vault {
	if l.err == nil {
		l.accessToken = token
	}
	return l
}

// WithAuth sets the authorization method to use
func (l *Vault) WithAuth(auth VaultAuthMethod) *Vault {
	if l.err == nil {
		if auth != nil {
			var err error

			l.auth, err = auth.create()
			if err != nil {
				l.err = err
			}
		} else {
			l.auth = nil
		}
	}
	return l
}

// Load loads the content from the web
func (l *Vault) Load(ctx context.Context) ([]byte, error) {
	var vaultClient *api.Client
	var secret *api.Secret
	var buf bytes.Buffer
	var err error

	// Create the http client object and set up transport
	client := http.Client{
		Transport: httpTransport,
	}
	if l.tlsConfig != nil {
		// Clone our default transport
		t := httpTransport.Clone()
		t.TLSClientConfig = l.tlsConfig

		client.Transport = t
	}

	vaultCfg := api.Config{
		MaxRetries: 3,
		Address:    "http",
	}

	// Get host
	if len(l.host) == 0 {
		return nil, errors.New("host not set")
	}
	if l.tlsConfig != nil {
		vaultCfg.Address += "s"
	}
	vaultCfg.Address += "://" + l.host

	// Create Vault client accessor
	vaultClient, err = api.NewClient(&vaultCfg)
	if err != nil {
		return nil, err
	}
	// Remove some settings that can be overridden with environment variables
	vaultClient.SetToken("")
	vaultClient.SetNamespace("")

	// Add custom headers if provided
	if len(l.headers) > 0 {
		for key, value := range l.headers {
			vaultClient.AddHeader(key, value)
		}
	}

	// Get access token
	if len(l.accessToken) > 0 {
		// If a token was provided, use it
		vaultClient.SetToken(l.accessToken)
	} else if l.auth != nil {
		// Else attempt to log in using the provided authentication method
		_, err = vaultClient.Auth().Login(ctx, l.auth)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("no authorization method was established")
	}

	// Read secret
	if len(l.path) == 0 {
		return nil, errors.New("path not set")
	}
	secret, err = vaultClient.Logical().ReadWithContext(ctx, l.path)
	if err != nil {
		return nil, err
	}

	// If we don't have a secret but also no errors
	if secret == nil {
		return nil, errors.New("data not found")
	}

	// Extract data and re-encode as JSON
	data, ok := secret.Data["data"]
	if !ok || data == nil {
		return nil, errors.New("data not found")
	}

	// Re-encode all as the original received json
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	err = enc.Encode(data)
	if err != nil {
		return nil, err
	}

	// Done
	return buf.Bytes(), nil
}
