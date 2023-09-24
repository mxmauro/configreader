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
	cmdLineOption := addressOfString("--settings")
	cmdLineOptionShort := addressOfString("-S")

	if CmdLineParameter != nil {
		if len(*CmdLineParameter) > 0 {
			cmdLineOption = addressOfString("--" + *CmdLineParameter)
		} else {
			cmdLineOption = nil
		}
	}
	if CmdLineParameterShort != nil {
		if len(*CmdLineParameterShort) > 0 {
			cmdLineOptionShort = addressOfString("-" + *CmdLineParameterShort)
		} else {
			cmdLineOptionShort = nil
		}
	}

	// Lookup for the parameter's value.
	if cmdLineOption != nil || cmdLineOptionShort != nil {
		for idx, value := range os.Args[1:] {
			if (cmdLineOption != nil && value == *cmdLineOption) || (cmdLineOptionShort != nil && value == *cmdLineOptionShort) {
				if idx+1 < len(os.Args) {
					l.WithFilename(os.Args[idx+1])
				} else {
					l.err = errors.New("missing filename in '" + value + "' parameter")
				}
				break
			}
		}
		if l.err == nil && len(l.filename) == 0 {
			l.err = errors.New("command line parameter not found")
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
func (l *File) Load(ctx context.Context) ([]byte, error) {
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

func addressOfString(s string) *string {
	return &s
}
