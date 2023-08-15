package loader

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mxmauro/configreader/internal/helpers"
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

	query map[string]string

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
		query:   make(map[string]string),
		headers: make(map[string]string),
	}
}

// WithHost sets the host address and, optionally, the port
func (l *Http) WithHost(host string) *Http {
	if l.err == nil {
		host, l.err = helpers.LoadAndReplaceEnvs(host)
		if l.err == nil {
			l.host = host
		}
	}
	return l
}

// WithPath sets the path
func (l *Http) WithPath(path string) *Http {
	if l.err == nil {
		path, l.err = helpers.LoadAndReplaceEnvs(path)
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

			username, err = helpers.LoadAndReplaceEnvs(username)
			if err == nil {
				password, err = helpers.LoadAndReplaceEnvs(password)
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
func (l *Http) WithQuery(query map[string]string) *Http {
	if l.err == nil {
		var err error

		queryCopy := make(map[string]string)
		for key, value := range query {
			if len(key) == 0 {
				err = errors.New("invalid query parameter")
				break
			}

			value, err = helpers.LoadAndReplaceEnvs(value)
			if err != nil {
				break
			}

			if len(value) > 0 {
				queryCopy[key] = value
			}
		}

		if l.err == nil {
			l.query = queryCopy
		} else {
			l.err = err
		}
	}
	return l
}

// WithQueryItem sets a single query parameter
func (l *Http) WithQueryItem(key string, value string) *Http {
	if l.err == nil {
		var err error

		value, err = helpers.LoadAndReplaceEnvs(value)
		if err == nil {
			if l.query == nil {
				l.query = make(map[string]string)
			}
			l.query[key] = value
		} else {
			l.err = err
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
func (l *Http) WithHeaderItem(key string, value string) *Http {
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
		l.tlsConfig = tlsConfig
	}
	return l
}

// WithUrl sets the options from the provided url
func (l *Http) WithUrl(rawURL string) *Http {
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

		if u.User != nil {
			password, _ := u.User.Password()
			_ = l.WithCredentials(u.User.Username(), password)
		}

		_ = l.WithPath(u.Path)

		l.query = make(map[string]string)
		for key, value := range u.Query() {
			if len(value) > 0 && len(value[0]) != 0 {
				l.query[key] = value[0]
			}
		}
	}

	// Done
	return l
}

// Load loads the content from the web
func (l *Http) Load(ctx context.Context) ([]byte, error) {
	var resp *http.Response
	var err error

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
		values := url.Values{}
		for key, value := range l.query {
			values.Add(key, value)
		}
		u.RawQuery = values.Encode()
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
	if err != nil {
		return nil, err
	}

	// Check if the request succeeded
	if resp.StatusCode != 200 {
		_ = resp.Body.Close()

		return nil, fmt.Errorf("unexpected HTTP status code [http-status=%v]", resp.Status)
	}

	// Read response body
	var responseBody []byte
	responseBody, err = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// Done
	return responseBody, nil
}
