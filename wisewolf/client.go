package wisewolf

import (
	"net/http"
	"time"
)

var Client http.Client

func Init() {
	Client = http.Client{
		Transport: &loggingTransport{},
		Timeout:   time.Second * 30,
	}
}
