package loader

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"github.com/mxmauro/configreader/internal/helpers"
	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

// Vault wraps content to be loaded from a Hashicorp Vault instance
type Vault struct {
	host string
	path []string

	headers map[string]string

	tlsConfig *tls.Config

	accessToken string
	auth        VaultAuthMethod

	client *vaultClient

	err error
}

// VaultAuthMethod is an interface used to set up an authentication mechanism
type VaultAuthMethod interface {
	create() (api.AuthMethod, error)
	hash() [32]byte
}

// -----------------------------------------------------------------------------

var (
	regexMultipleSlashes = regexp.MustCompile(`/+`)
)

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
		host, l.err = helpers.ExpandEnvVars(host)
		if l.err == nil {
			l.host = host
		}
	}
	return l
}

// WithPath sets the path
func (l *Vault) WithPath(path string) *Vault {
	if l.err == nil {
		l.path = make([]string, 0)
		l.AddPath(path)
	}
	return l
}

// AddPath adds a new path to the list
func (l *Vault) AddPath(path string) *Vault {
	if l.err == nil {
		path, l.err = helpers.ExpandEnvVars(path)
		if l.err == nil {
			path = strings.Replace(path, "\\", "/", -1)
			path = regexMultipleSlashes.ReplaceAllString(path, "/")
			if strings.HasPrefix(path, "/") {
				path = path[1:]
			}
			if strings.HasSuffix(path, "/") {
				path = path[:len(path)-1]
			}
			if len(path) > 0 {
				ignore := false
				for _, p := range l.path {
					if p == path {
						ignore = true
						break
					}
				}
				if !ignore {
					if l.path == nil {
						l.path = make([]string, 0)
					}
					l.path = append(l.path, path)
				}
			} else {
				l.err = errors.New("invalid path")
			}
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

			value, err = helpers.ExpandEnvVars(value)
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

		value, err = helpers.ExpandEnvVars(value)
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
		if tlsConfig != nil {
			l.tlsConfig = tlsConfig.Clone()
		} else {
			l.tlsConfig = nil
		}
	}
	return l
}

// WithAccessToken sets the access token to use as authorization
func (l *Vault) WithAccessToken(token string) *Vault {
	if l.err == nil {
		l.accessToken = token
		l.auth = nil
	}
	return l
}

// WithAuth sets the authorization method to use
func (l *Vault) WithAuth(auth VaultAuthMethod) *Vault {
	if l.err == nil {
		l.auth = auth
		l.accessToken = ""
	}
	return l
}

// WithURL sets the host, port, path and other settings from the provided url
func (l *Vault) WithURL(rawURL string) *Vault {
	if l.err == nil {
		// Parse url
		if strings.HasPrefix(rawURL, "vault://") {
			rawURL = "http://" + rawURL[8:]
		} else if strings.HasPrefix(rawURL, "vaults://") {
			rawURL = "https://" + rawURL[9:]
		}
		u, err := url.Parse(rawURL)
		if err != nil {
			l.err = err
			return l
		}

		// Replace settings
		if u.Scheme == "https" {
			_ = l.WithDefaultTLS()
		} else if u.Scheme == "http" {
			_ = l.WithTLS(nil)
		} else {
			l.err = errors.New("unsupported scheme")
			return l
		}

		_ = l.WithHost(u.Host)

		_ = l.WithPath(u.Path)

		// Get and validate path locations to read
		for _, p := range findQueryValues(u, "path") {
			_ = l.AddPath(p)
		}
		if l.err != nil {
			return l
		}
		if len(l.path) == 0 {
			l.err = errors.New("invalid Vault url (path not specified or invalid)")
			return l
		}

		// Figure out the auth login mount path
		mountPath := findFirstQueryValue(u, "mountPath")

		// Check if a role name was provided
		roleName := findFirstQueryValue(u, "roleName")

		// Check if AppRole credentials were provided (both or none must be specified)
		roleID := findFirstQueryValue(u, "roleId")
		secretID := findFirstQueryValue(u, "secretId")

		// Determine the auth method (or autodetect)
		method := findFirstQueryValue(u, "method")
		if len(method) > 0 {
			if method != "approle" && method != "iam" && method != "k8s" {
				l.err = errors.New("invalid Vault url (method not supported)")
				return l
			}
		} else {
			// Try to guess
			if len(roleID) > 0 && len(secretID) > 0 {
				method = "approle"
			} else if len(os.Getenv("KUBERNETES_SERVICE_HOST")) > 0 {
				method = "k8s"
			} else if len(os.Getenv("EC2_INSTANCE_ID")) > 0 || len(os.Getenv("ECS_CONTAINER_METADATA_URI_V4")) > 0 || len(os.Getenv("AWS_ROLE_ARN")) > 0 || len(os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE")) > 0 {
				method = "iam"
			} else {
				var req *http.Request
				var resp *http.Response

				client := http.Client{
					Transport: httpTransport,
				}

				// Create a new request
				req, err = http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/", nil)
				if err != nil {
					l.err = err
					return l
				}

				// Execute request
				ctxWithTimeout, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer ctxCancel()

				resp, err = client.Do(req.WithContext(ctxWithTimeout))
				if resp != nil && resp.Body != nil {
					defer func() {
						_ = resp.Body.Close()
					}()
				}

				// If no error, no matter the reason, we are on something running at AWS
				if err == nil {
					method = "iam"
				}
			}
		}

		// Prepare payload and set mount path if not provided
		switch method {
		case "approle":
			if len(roleID) == 0 || len(secretID) == 0 {
				l.err = errors.New("invalid Vault url (both roleId and secretId parameters are required)")
				return l
			}

			auth := NewVaultAppRoleAuthMethod()

			_ = auth.WithRoleId(roleID)
			_ = auth.WithSecretId(secretID)

			if len(mountPath) > 0 {
				_ = auth.WithMountPath(mountPath)
			} else {
				_ = auth.WithMountPath("approle")
			}

			_ = l.WithAuth(auth)

		case "k8s":
			if len(roleName) == 0 {
				l.err = errors.New("invalid Vault url (roleName parameter is required)")
				return l
			}

			auth := NewVaultKubernetesAuthMethod()

			_ = auth.WithRole(roleName)

			if len(mountPath) > 0 {
				_ = auth.WithMountPath(mountPath)
			} else {
				_ = auth.WithMountPath("kubernetes")
			}

			_ = l.WithAuth(auth)

		case "iam":
			if len(roleName) == 0 {
				l.err = errors.New("invalid Vault url (roleName parameter is required)")
				return l
			}

			auth := NewVaultAwsAuthMethod()

			_ = auth.WithRole(roleName)

			_ = auth.WithTypeIAM()

			serverId := findFirstQueryValue(u, "serverId")
			if len(serverId) > 0 {
				_ = auth.WithIamServerID(serverId)
			}

			region := findFirstQueryValue(u, "region")
			if len(region) > 0 {
				_ = auth.WithRegion(region)
			}

			if len(mountPath) > 0 {
				_ = auth.WithMountPath(mountPath)
			} else {
				_ = auth.WithMountPath("aws")
			}

			_ = l.WithAuth(auth)
		}
	}

	// Done
	return l
}

// Load loads the content from Vault
func (l *Vault) Load(ctx context.Context) (model.Values, error) {
	var secret *api.Secret
	var ret model.Values
	var err error

	if l.err != nil {
		return nil, l.err
	}
	if len(l.path) == 0 {
		l.err = errors.New("path not set")
		return nil, l.err
	}

	if l.client == nil {
		l.client, err = newVaultClient(l.host, l.headers, l.tlsConfig, l.accessToken, l.auth)
		if err != nil {
			l.err = err
			return nil, l.err
		}
	}

	// Read secrets
	ret = make(model.Values)
	for _, p := range l.path {
		secret, err = l.client.readWithContext(ctx, p)
		if err != nil {
			return nil, err
		}

		// If we don't have a secret but also no errors, skip
		if secret != nil {
			// Extract data
			values := secret.Data
			if data, ok := values["data"]; ok && data != nil {
				_values, ok2 := data.(map[string]interface{}) // Using 'map[string]interface{}' instead of model.Values else casting won't work
				if ok2 {
					values = _values
				} else {
					values = nil
				}
			}

			// Merge
			for k, v := range values {
				ret[k] = v
			}
		}
	}

	// Done
	return ret, nil
}

// -----------------------------------------------------------------------------

func findQueryValues(u *url.URL, param string) []string {
	for k, values := range u.Query() {
		if strings.EqualFold(k, param) {
			return values
		}
	}
	return nil
}

func findFirstQueryValue(u *url.URL, param string) string {
	values := findQueryValues(u, param)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}
