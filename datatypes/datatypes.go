package datatypes

import "io"

type Book struct  {
	Title string
	Price float32
	Site string
}

type ResponseDetails struct {
	Address string
	Content string
	Reader io.Reader
}
