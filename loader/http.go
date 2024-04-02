package loader

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mxmauro/configreader/internal/helpers"
	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

const (
	httpRequestTimeout         = 10 * time.Second
	httpResponseHeadersTimeout = 5 * time.Second
)

// -----------------------------------------------------------------------------

var (
	httpTransport *http.Transport
)

// -----------------------------------------------------------------------------

// Http wraps content to be loaded from a URL
type Http struct {
	host string
	path string

	credentials *url.Userinfo

	query map[string][]string

	headers map[string]string

	tlsConfig *tls.Config

	err error
}

// -----------------------------------------------------------------------------

func init() {
	// Create custom http transport
	// From: https://www.loginradius.com/blog/async/tune-the-go-http-client-for-high-performance/
	httpTransport = http.DefaultTransport.(*http.Transport).Clone()
	httpTransport.MaxIdleConns = 10
	httpTransport.MaxConnsPerHost = 10
	httpTransport.IdleConnTimeout = 60 * time.Second
	httpTransport.MaxIdleConnsPerHost = 10
	httpTransport.ResponseHeaderTimeout = httpResponseHeadersTimeout
}

// -----------------------------------------------------------------------------

// NewHttp create a new web loader
func NewHttp() *Http {
	return &Http{
		query:   make(map[string][]string),
		headers: make(map[string]string),
	}
}

// WithHost sets the host address and, optionally, the port
func (l *Http) WithHost(host string) *Http {
	if l.err == nil {
		host, l.err = helpers.ExpandEnvVars(host)
		if l.err == nil {
			l.host = host
		}
	}
	return l
}

// WithPath sets the resource path
func (l *Http) WithPath(path string) *Http {
	if l.err == nil {
		path, l.err = helpers.ExpandEnvVars(path)
		if l.err == nil {
			l.path = path
		}
	}
	return l
}

// WithCredentials sets the username and password
func (l *Http) WithCredentials(username string, password string) *Http {
	if l.err == nil {
		if len(username) > 0 || len(password) > 0 {
			var err error

			username, err = helpers.ExpandEnvVars(username)
			if err == nil {
				password, err = helpers.ExpandEnvVars(password)
			}
			if err == nil {
				l.credentials = url.UserPassword(username, password)
			} else {
				l.err = err
			}
		} else {
			l.credentials = nil
		}
	}
	return l
}

// WithQuery sets the query parameters
func (l *Http) WithQuery(query map[string][]string) *Http {
	if l.err == nil {
		var err error

		queryCopy := make(map[string][]string)
		for key, values := range query {
			if len(key) == 0 {
				err = errors.New("invalid query parameter")
				break
			}

			copiedValues := make([]string, 0)
			for _, value := range values {
				value, err = helpers.ExpandEnvVars(value)
				if err != nil {
					break
				}
				copiedValues = append(copiedValues, value)
			}

			queryCopy[key] = copiedValues
		}

		if err == nil {
			l.query = queryCopy
		} else {
			l.err = err
		}
	}
	return l
}

// WithQueryItem sets a single query parameter
func (l *Http) WithQueryItem(key string, values []string) *Http {
	if l.err == nil {
		if len(key) > 0 {
			var err error

			copiedValues := make([]string, 0)
			for _, value := range values {
				value, err = helpers.ExpandEnvVars(value)
				if err != nil {
					break
				}
				copiedValues = append(copiedValues, value)
			}

			if err == nil {
				if l.query == nil {
					l.query = make(map[string][]string)
				}
				l.query[key] = copiedValues
			} else {
				l.err = err
			}
		} else {
			l.err = errors.New("invalid query parameter")
		}
	}
	return l
}

// WithHeaders sets the request headers
func (l *Http) WithHeaders(headers map[string]string) *Http {
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
func (l *Http) WithHeaderItem(key string, value string) *Http {
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
func (l *Http) WithDefaultTLS() *Http {
	if l.err == nil {
		l.tlsConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}
	}
	return l
}

// WithTLS sets a tls.Config object
func (l *Http) WithTLS(tlsConfig *tls.Config) *Http {
	if l.err == nil {
		if tlsConfig != nil {
			l.tlsConfig = tlsConfig.Clone()
		} else {
			l.tlsConfig = nil
		}
	}
	return l
}

// WithURL sets the host, port, path and other settings from the provided url
func (l *Http) WithURL(rawURL string) *Http {
	if l.err == nil {
		// Parse url
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

		if u.User != nil {
			password, _ := u.User.Password()
			_ = l.WithCredentials(u.User.Username(), password)
		}

		l.WithQuery(u.Query())
	}

	// Done
	return l
}

// Load loads the content from the web
func (l *Http) Load(ctx context.Context) (model.Values, error) {
	var resp *http.Response

	// If an error was set by a With... function, return it
	if l.err != nil {
		return nil, l.err
	}

	// Start building URL
	u := &url.URL{
		Scheme: "http",
	}
	if l.tlsConfig != nil {
		u.Scheme = "https"
	}

	// Set host
	if len(l.host) == 0 {
		return nil, errors.New("host not set")
	}
	u.Host = l.host

	// Set path
	if len(l.path) == 0 {
		return nil, errors.New("path not set")
	}
	u.Path = l.path

	// Set credentials
	u.User = l.credentials

	// Set query parameters if provided
	if len(l.query) > 0 {
		query := url.Values{}
		for key, values := range l.query {
			for _, value := range values {
				query.Add(key, value)
			}
		}
		u.RawQuery = query.Encode()
	}

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

	// Create a new request
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Add custom headers if provided
	if len(l.headers) > 0 {
		for key, value := range l.headers {
			req.Header.Add(key, value)
		}
	}

	// Execute request
	ctxWithTimeout, ctxCancel := context.WithTimeout(ctx, httpRequestTimeout)
	defer ctxCancel()

	resp, err = client.Do(req.WithContext(ctxWithTimeout))
	if resp != nil && resp.Body != nil {
		defer func() {
			_ = resp.Body.Close()
		}()
	}
	if err != nil {
		return nil, err
	}

	// Check if the request succeeded
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unexpected HTTP status code [http-status=%v]", resp.Status)
	}

	// Read response body
	var responseBody []byte
	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse data
	if strings.HasSuffix(u.Path, ".env") {
		return parseData(responseBody, parseDataHintIsDotEnv)
	}
	return parseData(responseBody, 0)
}
