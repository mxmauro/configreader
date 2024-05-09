package loader

import (
	"os"
	"strings"
)

// -----------------------------------------------------------------------------

func getCmdLineParamValue(longOpt string, shortOpt string) (location string, errParam string) {
	if len(longOpt) == 0 && len(shortOpt) == 0 {
		return
	}

	opts := make([]string, 0)
	if len(longOpt) > 0 {
		opts = append(opts, "--"+longOpt, "/"+longOpt)
	}
	if len(shortOpt) > 0 {
		opts = append(opts, "-"+shortOpt, "/"+shortOpt)
	}

	for idx, value := range os.Args[1:] {
		for _, opt := range opts {
			if value == opt {
				// Offset +2 because we range from the second argument
				if idx+2 < len(os.Args) {
					location = os.Args[idx+2]
					if len(location) == 0 {
						errParam = opt
					}
				} else {
					errParam = opt
				}
				return
			}
			if strings.HasPrefix(value, opt+"=") || strings.HasPrefix(value, opt+":") {
				if len(value) > len(opt)+1 {
					location = value[len(opt)+1:]
				} else {
					errParam = opt
				}
				return
			}
		}
	}

	return
}
