package loader

import (
	"errors"
	"os"
	"strings"
)

// -----------------------------------------------------------------------------

// AutoDetectOptions indicates the behavior of the auto-detect module
type AutoDetectOptions struct {
	// CmdLine specifies if command line parameters are checked in first place
	CmdLine struct {
		Check bool
		// Long is the long version of the command line parameter. Defaults to --settings
		Long string
		// Short is the short version of the command line parameter. Defaults to -S
		Short string
	}

	// EnvVar is the environment variable that contains a file or url of the configuration settings
	EnvVar string

	// TlsEnvVar can specify and override where to look for client side certificates
	TlsEnvVar TlsEnvVarNames
}

// TlsEnvVarNames establishes environment variables to look for client certificates for http requests
type TlsEnvVarNames struct {
	CaCert     string // Defaults to SSL_CA_CERT environment variable
	ClientCert string // Defaults to SSL_CLIENT_CERT environment variable
	ClientKey  string // Defaults to SSL_CLIENT_KEY environment variable
}

// -----------------------------------------------------------------------------

// NewAutoDetect tries to create a new loader based on the origin type, like a file or Hashicorp Vault
func NewAutoDetect(opts AutoDetectOptions) Loader {
	location := ""

	// Try to get the location from the command line options
	if opts.CmdLine.Check {
		var errParam string

		longOpt := "--settings"
		shortOpt := "-S"
		if len(opts.CmdLine.Long) > 0 || len(opts.CmdLine.Short) > 0 {
			longOpt = opts.CmdLine.Long
			shortOpt = opts.CmdLine.Short
		}

		location, errParam = getCmdLineParamValue(longOpt, shortOpt)
		if len(errParam) > 0 {
			return &errorLoader{
				err: errors.New("missing filename/url in '" + errParam + "' parameter"),
			}
		}
	}

	// Try to obtain the location from an environment variable
	if len(location) == 0 && len(opts.EnvVar) > 0 {
		location = os.Getenv(opts.EnvVar)
	}

	// Check if a location was found
	if len(location) == 0 {
		return &errorLoader{
			err: errors.New("unable to find configuration source location"),
		}
	}

	// Check for a Vault URL
	if strings.HasPrefix(location, "vault://") {
		loader := NewVault()
		loader.WithURL(location)
		return loader
	}
	if strings.HasPrefix(location, "vaults://") {
		tlsCfg, err := createTlsConfig(opts.TlsEnvVar)
		if err != nil {
			return &errorLoader{
				err: err,
			}
		}

		loader := NewVault()
		loader.WithURL(location)
		loader.WithTLS(tlsCfg)
		return loader
	}
	if strings.HasPrefix(location, "http://") {
		loader := NewHttp()
		loader.WithURL(location)
		return loader
	}
	if strings.HasPrefix(location, "https://") {
		tlsCfg, err := createTlsConfig(opts.TlsEnvVar)
		if err != nil {
			return &errorLoader{
				err: err,
			}
		}

		loader := NewHttp()
		loader.WithURL(location)
		loader.WithTLS(tlsCfg)
		return loader
	}

	if strings.HasPrefix(location, "file://") {
		ofs := 7
		for ofs < len(location) && location[ofs] == '/' {
			ofs += 1
		}
		location = "file:" + strings.Repeat("/", fileSlashesCount) + location[ofs:]
	}

	loader := NewFile()
	loader.WithFilename(location)
	return loader
}
