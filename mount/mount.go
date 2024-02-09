package mount

import (
	"os"
)

var user, pass, domain, url string

func Init() {
	var check bool

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
