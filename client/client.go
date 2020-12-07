package client

import "net/http"

var httpClient *http.Client

func InitClient() {
	httpClient = new(http.Client)
}

func Get() *http.Client {
	return httpClient
}