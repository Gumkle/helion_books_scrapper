package sinks

import (
	"github.com/Gumkle/helion_price_scrapper/datatypes"
	"net/http"
)

var Responses chan *http.Response
var Links chan string
var Books chan datatypes.Book

func InitSinks() {
	Books = make(chan datatypes.Book)
	Responses = make(chan *http.Response)
	Links = make(chan string)
}