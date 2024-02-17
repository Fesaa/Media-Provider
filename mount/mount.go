package mount

import (
	"os"
)

var user, pass, domain, url string

// If doMount is set to false, the mount package will not attempt to mount the share.
// And a local temp will be used to download to. This is useful for testing.
var doMount bool = true

// Initializes the mount package.
// It will panic if any of the required environment variables are not set.
func Init() {
	if os.Getenv("MOUNT") == "false" {
		doMount = false
		return
	}

	var check bool = false

	user, check = os.LookupEnv("USER")
	if !check {
		panic("USER not found. Please set USER environment variable")
	}

	pass, check = os.LookupEnv("PASS")
	if !check {
		panic("PASS not found. Please set PASS environment variable")
	}

	domain, check = os.LookupEnv("DOMAIN")
	if !check {
		panic("DOMAIN not found. Please set DOMAIN environment variable")
	}

	url, check = os.LookupEnv("URL")
	if !check {
		panic("URL not found. Please set URL environment variable")
	}
}

func WantsMount() bool {
	return doMount
}
