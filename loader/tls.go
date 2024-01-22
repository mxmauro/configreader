package loader

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"os"
)

// -----------------------------------------------------------------------------

func createTlsConfig(names TlsEnvVarNames) (*tls.Config, error) {
	var certBytes, keyBytes []byte
	var err error

	cfg := tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Check for a CA certificate and load
	if len(names.CaCert) > 0 {
		certBytes, err = loadFromEnvVar(names.CaCert)
	} else {
		certBytes, err = loadFromEnvVar("SSL_CA_CERT")
	}
	if err != nil {
		return nil, err
	}
	if certBytes != nil {
		cfg.RootCAs = x509.NewCertPool()
		for len(certBytes) > 0 {
			var block *pem.Block
			var cert *x509.Certificate

			// Decode block
			block, certBytes = pem.Decode(certBytes)
			if block == nil {
				break
			}
			if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
				continue
			}

			// Parse certificate
			cert, err = x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}

			// Add to pool
			cfg.RootCAs.AddCert(cert)
		}
	}

	// Check for a client certificate and load
	if len(names.ClientCert) > 0 {
		certBytes, err = loadFromEnvVar(names.ClientCert)
	} else {
		certBytes, err = loadFromEnvVar("SSL_CLIENT_CERT")
	}
	if err != nil {
		return nil, err
	}
	if len(names.ClientKey) > 0 {
		keyBytes, err = loadFromEnvVar(names.ClientKey)
	} else {
		keyBytes, err = loadFromEnvVar("SSL_CLIENT_KEY")
	}
	if err != nil {
		return nil, err
	}
	if certBytes != nil && keyBytes != nil {
		var cert tls.Certificate

		// Build client certificate pair
		cert, err = tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, err
		}

		// Add to TLS config
		cfg.Certificates = []tls.Certificate{cert}
	}

	// Done
	return &cfg, nil
}

func loadFromEnvVar(varName string) ([]byte, error) {
	value := []byte(os.Getenv(varName))
	if len(value) == 0 {
		return nil, nil
	}

	if isPEM(value) {
		return value, nil
	}

	// Assume a filename
	valueBytes, err := os.ReadFile(string(value))
	if err != nil {
		return nil, err
	}

	// Done
	return valueBytes, nil
}

func isPEM(value []byte) bool {
	idx := 0
	l := len(value)
	for idx < l {
		if value[idx] != '\n' && value[idx] != '\r' && value[idx] != '\t' && value[idx] != ' ' {
			break
		}
		idx += 1
	}
	if bytes.Compare(value[idx:], []byte("-----")) == 0 {
		return true
	}
	return false
}
