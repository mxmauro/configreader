package loader

import (
	"os"
)

// -----------------------------------------------------------------------------

func getCmdLineParamValue(longOpt string, shortOpt string) (location string, errParam string) {
	if len(longOpt) > 0 || len(shortOpt) > 0 {
		if len(longOpt) > 0 {
			longOpt = "--" + longOpt
		} else {
			longOpt = ""
		}
		if len(shortOpt) > 0 {
			shortOpt = "-" + shortOpt
		} else {
			shortOpt = ""
		}
	}

	if len(longOpt) > 0 {
		for idx, value := range os.Args[1:] {
			if longOpt == value {
				if idx+1 < len(os.Args) {
					location = os.Args[idx+1]
				} else {
					errParam = longOpt
				}
				return
			}
		}
	}

	if len(shortOpt) > 0 {
		for idx, value := range os.Args[1:] {
			if shortOpt == value {
				if idx+1 < len(os.Args) {
					location = os.Args[idx+1]
				} else {
					errParam = shortOpt
				}
				return
			}
		}
	}

	return
}
