package pasloe

import (
	"github.com/Fesaa/Media-Provider/providers/pasloe/api"
	"github.com/Fesaa/Media-Provider/providers/pasloe/mangadex"
)

func Init(cfg api.Config) {
	c = newClient(cfg)
	mangadex.Init()
}

var c api.Client

func I() api.Client {
	return c
}
