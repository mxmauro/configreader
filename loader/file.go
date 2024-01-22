package loader

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/mxmauro/configreader/internal/helpers"
)

// -----------------------------------------------------------------------------

// File wraps content to be loaded from a file on disk
type File struct {
	filename string

	err error
}

// -----------------------------------------------------------------------------

// NewFile create a new file loader
func NewFile() *File {
	return &File{}
}

// NewFileFromCommandLine create a new file loader from a command line parameter
func NewFileFromCommandLine(CmdLineParameter *string, CmdLineParameterShort *string) *File {
	l := &File{}

	// Setup command-line parameters to look for.
	longOpt := "--settings"
	shortOpt := "-S"

	if CmdLineParameter != nil {
		longOpt = *CmdLineParameter
	}
	if CmdLineParameterShort != nil {
		shortOpt = *CmdLineParameterShort
	}

	// Lookup for the parameter's value.
	if len(longOpt) > 0 || len(shortOpt) > 0 {
		filename, errParam := getCmdLineParamValue(longOpt, shortOpt)
		if len(filename) > 0 {
			l.WithFilename(filename)
		} else if len(errParam) == 0 {
			l.err = errors.New("command line parameter not found")
		} else {
			l.err = errors.New("missing filename in '" + errParam + "' parameter")
		}
	} else {
		l.err = errors.New("no command-line option was specified")
	}

	// Done
	return l
}

// NewFileFromEnvironmentVariable create a new file loader from an environment variable
func NewFileFromEnvironmentVariable(Name string) *File {
	l := &File{}

	filename, ok := os.LookupEnv(Name)
	if ok {
		l.WithFilename(filename)
	} else {
		l.err = errors.New("environment variable '" + Name + "' not found")
	}

	// Done
	return l
}

// WithFilename sets the file name
func (l *File) WithFilename(filename string) *File {
	if l.err == nil {
		filename, l.err = expandAndNormalizeFilename(filename)
		if l.err == nil {
			l.filename = filename
		}
	}
	return l
}

// Load loads the content from the file
// NOTE: We are not making use of the context assuming configuration files will be small and on a local disk
func (l *File) Load(_ context.Context) ([]byte, error) {
	// If an error was set by a With... function, return it
	if l.err != nil {
		return nil, l.err
	}

	// Get filename
	if len(l.filename) == 0 {
		return nil, errors.New("filename not set")
	}

	// Load file
	return os.ReadFile(l.filename)
}

func expandAndNormalizeFilename(filename string) (string, error) {
	var err error

	filename, err = helpers.LoadAndReplaceEnvs(filename)
	if err != nil {
		return "", err
	}

	// Convert slashes
	filename = filepath.FromSlash(filename)

	// Convert path to absolute
	if !filepath.IsAbs(filename) {
		var currentPath string

		currentPath, err = os.Getwd()
		if err != nil {
			return "", err
		}
		filename = filepath.Join(currentPath, filename)
	}

	// Done
	return filename, err
}
