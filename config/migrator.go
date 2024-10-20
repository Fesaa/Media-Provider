package config

import (
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type versionMap struct {
	old int
	new int
}

const (
	currentVersion = 3
)

var (
	versionMappers map[versionMap]func(c Config) Config
)

func init() {
	versionMappers = make(map[versionMap]func(c Config) Config)

	versionMappers[versionMap{0, 1}] = func(c Config) Config {
		c.Cache = CacheConfig{Type: MEMORY}
		return c
	}
	versionMappers[versionMap{1, 2}] = func(c Config) Config {
		apiKey, err := ApiKey()
		if err != nil {
			panic(err)
		}

		c.ApiKey = apiKey
		return c
	}
	versionMappers[versionMap{2, 3}] = func(c Config) Config {
		hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
		if err != nil {
			panic(err)
		}
		c.Password = base64.StdEncoding.EncodeToString(hash)
		return c
	}
}

func update(c Config) Config {
	if c.Version >= currentVersion {
		return c
	}

	fmt.Println("Version is lower than wanted, updating...")
	for c.Version < currentVersion {
		m := versionMap{c.Version, c.Version + 1}
		if f, ok := versionMappers[m]; ok {
			c = f(c)
		}
		c.Version = m.new
	}
	fmt.Println("Migration finished, saving...")

	if err := c.Save(); err != nil {
		panic(err)
	}

	return c
}
